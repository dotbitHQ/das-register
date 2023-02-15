package block_parser

import (
	"das_register_server/internal/reverse_smt"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/smt"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"gorm.io/gorm"
)

func (b *BlockParser) ActionReverseRecordRoot(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameReverseRecordRootCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	outpoint := common.OutPoint2String(req.TxHash, 0)
	// find local tx exist or not
	reverseSmtTaskInfo, err := b.DbDao.SearchReverseSmtTaskInfo(outpoint)
	if err != nil {
		resp.Err = err
		return
	}

	// is other service provider
	if reverseSmtTaskInfo.ID == 0 {
		if err := b.syncOtherProvider(req); err != nil {
			resp.Err = err
			return
		}
		return
	}

	res, err := b.DasCore.Client().GetTransaction(b.Ctx, types.HexToHash(req.TxHash))
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	// ready to rollback
	if res.TxStatus.Status == types.TransactionStatusRejected {
		resp.Err = b.DbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusRollback, tables.ReverseSmtTxStatusReject, "outpoint=?", outpoint)
		return
	}

	// update smt_status=2 and tx_status=2
	resp.Err = b.DbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusConfirm, tables.ReverseSmtTxStatusConfirm, "outpoint=?", outpoint)
	return
}

// syncOtherProvider sync reverse record from other provider
func (b *BlockParser) syncOtherProvider(req FuncTransactionHandleReq) error {
	// This situation does not exist in the first phase, parse the witness synchronous transaction to the local task_info reverse_record table
	smtBuilder := witness.NewReverseSmtBuilder()
	txReverseSmtRecord, err := smtBuilder.FromTx(req.Tx)
	if err != nil {
		return err
	}

	outpoint := common.OutPoint2String(req.TxHash, 0)
	refOutput := req.Tx.Inputs[0].PreviousOutput
	taskInfo := &tables.ReverseSmtTaskInfo{
		RefOutpoint: common.OutPoint2String(refOutput.TxHash.String(), refOutput.Index),
		BlockNumber: req.BlockNumber,
		Outpoint:    outpoint,
		Timestamp:   int64(req.BlockTimestamp),
		SmtStatus:   tables.ReverseSmtStatusConfirm,
		TxStatus:    tables.ReverseSmtTxStatusConfirm,
	}
	if err := taskInfo.InitTaskId(); err != nil {
		return err
	}

	smtKv := make([]smt.SmtKv, 0)
	recordInfos := make([]*tables.ReverseSmtRecordInfo, 0)
	for _, v := range txReverseSmtRecord {
		recordInfo := &tables.ReverseSmtRecordInfo{
			Address:   v.Address,
			Nonce:     v.PrevNonce,
			TaskID:    taskInfo.TaskID,
			Account:   v.NextAccount,
			Sign:      v.Signature,
			SubAction: string(v.Action),
		}
		recordInfos = append(recordInfos, recordInfo)

		smtKey, err := recordInfo.GetSmtKey()
		if err != nil {
			return err
		}
		smtVal, err := recordInfo.GetSmtValue()
		if err != nil {
			return err
		}
		smtKv = append(smtKv, smt.SmtKv{
			Key:   smtKey,
			Value: smtVal,
		})
	}

	opt := smt.SmtOpt{GetRoot: true}
	if err := b.DbDao.DbTransaction(func(tx *gorm.DB) error {
		if err := tx.Create(taskInfo).Error; err != nil {
			return err
		}
		if err := tx.Create(recordInfos).Error; err != nil {
			return err
		}

		tree := reverse_smt.GetReverseSmt()
		// update SMT
		smtOutput, err := tree.UpdateSmt(smtKv, opt)
		if err != nil {
			return err
		}

		smtRootHash := string(req.Tx.OutputsData[0])
		if smtOutput.Root.String() != smtRootHash {
			log.Warnf("ActionReverseRecordRoot smtRoot: %s != txSmtRoot: %s", smtOutput.Root, smtRootHash)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}