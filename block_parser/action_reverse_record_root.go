package block_parser

import (
	"das_register_server/internal/reverse_smt"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (b *BlockParser) ActionReverseRecordRoot(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameReverseRecordRootCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	outpoint := common.OutPoint2String(req.TxHash, 0)
	res, err := b.DasCore.Client().GetTransaction(b.Ctx, types.HexToHash(req.TxHash))
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	if res.TxStatus.Status == types.TransactionStatusRejected {
		resp.Err = b.DbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusRollback, tables.ReverseSmtTxStatusReject, "outpoint=?", outpoint)
		return
	}

	// find local tx exist or not
	reverseSmtTaskInfo, err := b.DbDao.SearchReverseSmtTaskInfo(outpoint)
	if err != nil {
		resp.Err = err
		return
	}
	if reverseSmtTaskInfo.ID == 0 {
		// TODO 一期不存在这种情况，解析witness同步交易到本地 task_info reverse_record 表
		// TODO 更新SMT
		return
	}

	tree := reverse_smt.GetReverseSmt()
	rootH256, err := tree.GetSmtRoot()
	if err != nil {
		resp.Err = err
		return
	}

	smtRoot := string(res.Transaction.OutputsData[0])
	// update smt_status=2 and tx_status=2
	if smtRoot == rootH256.String() {
		resp.Err = b.DbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusConfirm, tables.ReverseSmtTxStatusConfirm, "outpoint=?", outpoint)
		return
	}
	return
}
