package timer

import (
	"das_register_server/internal/reverse_smt"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/smt"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const (
	ReverseRecordMaxTaskNum = 300
)

func (t *TxTimer) doReverseSmtTask() error {
	// check reverse_smt_info latest smt_root == online_smt_root
	rt, err := t.reverseSmtCheckIsLatestTask()
	if err != nil {
		return fmt.Errorf("reverseSmtCheckIsLatestTask err: %s", err)
	}
	if rt {
		return nil
	}

	// update smt_status=3 and tx_status=3 to smt_status=1 and tx_status=0 and retry=retry+1
	log.Infof("doReverseSmtTask update smt_status=3 and tx_status=3 to smt_status=1 and tx_status=0 and retry=retry+1")
	if err := t.dbDao.UpdateAllReverseSmtRollbackToTxPending(); err != nil {
		return fmt.Errorf("UpdateReverseSmtToTxPending err: %s", err)
	}

	// rollback smt if retry>=tables.ReverseSmtMaxRetryNum
	rt, err = t.reverseSmtTaskRollback()
	if err != nil {
		return fmt.Errorf("doReverseSmtTaskRollback err: %s", err)
	}
	if rt {
		return nil
	}

	// task assignment
	if err := t.reverseSmtTaskAssignment(); err != nil {
		return fmt.Errorf("reverseSmtTaskAssignment err: %s", err)
	}

	// get pending task
	smtPendingTask, err := t.reverseSmtGetPendingTask()
	if err != nil {
		return fmt.Errorf("reverseSmtGetPendingTask err: %s", err)
	}
	// not pending task, return go to next cycle
	if smtPendingTask.ID == 0 {
		return nil
	}

	// reverseSmtPendingCheck
	rt, err = t.reverseSmtPendingCheck(&smtPendingTask)
	if err != nil {
		return fmt.Errorf("reverseSmtPendingCheck err: %s", err)
	}
	if rt {
		return nil
	}

	// update local smt
	reverseRecordsByTaskID, err := t.dbDao.FindReverseSmtRecordInfoByTaskID(smtPendingTask.TaskID)
	if err != nil {
		return fmt.Errorf("FindReverseSmtRecordInfoByTaskID err: %s", err)
	}
	// update smt_status=2, tx_status=0
	smtOut, err := t.doReverseSmtUpdateSmt(smtPendingTask.ID, reverseRecordsByTaskID)
	if err != nil {
		return fmt.Errorf("doReverseSmtUpdateSmt err: %s", err)
	}

	// assembly transaction
	reverseRecordSmtLiveCell, err := t.dasCore.GetReverseRecordSmtCell()
	if err != nil {
		return fmt.Errorf("GetReverseRecordSmtCell err: %s", err)
	}
	txBuilderParams, err := t.reverseSmtAssemblyTx(reverseRecordSmtLiveCell, reverseRecordsByTaskID, smtOut)
	if err != nil {
		return fmt.Errorf("reverseSmtAssemblyTx err: %s", err)
	}

	// build transaction
	txBuilder := txbuilder.NewDasTxBuilderFromBase(t.reverseSmtTxBuilder, nil)
	if err := txBuilder.BuildTransaction(txBuilderParams); err != nil {
		return fmt.Errorf("BuildTransaction err: %s", err)
	}
	log.Debugf("doReverseSmtTask tx struct: %s", txBuilder.TxString())

	txHash, err := txBuilder.Transaction.ComputeHash()
	if err != nil {
		return fmt.Errorf("BuildTransaction err: %s", err)
	}

	// update task_info ref_outpoint
	refOutpoint := common.OutPointStruct2String(reverseRecordSmtLiveCell.OutPoint)
	outpoint := common.OutPoint2String(txHash.Hex(), 0)
	if err := t.dbDao.UpdateReverseSmtTaskInfo(map[string]interface{}{
		"ref_outpoint": refOutpoint,
		"outpoint":     outpoint,
	}, "id=?", smtPendingTask.ID); err != nil {
		return fmt.Errorf("UpdateReverseSmtTaskInfo err: %s", err)
	}
	log.Infof("doReverseSmtTask update task_info ref_outpoint: %s, outpoint: %s", refOutpoint, outpoint)

	// send transaction
	if _, sendTxErr := txBuilder.SendTransaction(); sendTxErr != nil {
		// update to smt_status=3 and tx_status=3
		if updateErr := t.dbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusRollback, tables.ReverseSmtTxStatusReject, "id=?", smtPendingTask.ID); updateErr != nil {
			return fmt.Errorf("SendTransaction err: %s UpdateReverseSmtTaskInfoStatus err: %s", sendTxErr, updateErr)
		}
		return fmt.Errorf("SendTransaction err: %s", sendTxErr)
	}
	log.Infof("doReverseSmtTask send tx: %s", txHash)

	// smt_status=2 and tx_status=1
	if err := t.dbDao.UpdateReverseSmtTaskInfo(map[string]interface{}{
		"tx_status": tables.ReverseSmtTxStatusPending,
	}, "id=?", smtPendingTask.ID); err != nil {
		return fmt.Errorf("UpdateReverseSmtTaskInfo err: %s", err)
	}
	log.Infof("doReverseSmtTask UpdateReverseSmtTaskInfo, tx_hash: %s tx_status: %d", txHash, tables.ReverseSmtTxStatusPending)
	return nil
}

// doReverseSmtUpdateSmt update local smt
func (t *TxTimer) doReverseSmtUpdateSmt(id uint64, reverseRecordsByTaskID []*tables.ReverseSmtRecordInfo) (*smt.UpdateMiddleSmtOut, error) {
	smtKvs := make([]smt.SmtKv, 0)
	for _, record := range reverseRecordsByTaskID {
		smtKey, err := record.GetSmtKey()
		if err != nil {
			return nil, fmt.Errorf("GetSmtKey err: %s", err)
		}
		smtVal, err := record.GetSmtValue()
		if err != nil {
			return nil, fmt.Errorf("GetSmtValue err: %s", err)
		}
		smtKvs = append(smtKvs, smt.SmtKv{
			Key:   smtKey,
			Value: smtVal,
		})
	}

	smtKvsData, _ := json.Marshal(smtKvs)
	log.Infof("doReverseSmtTask doReverseSmtUpdateSmt update local smt: %s", string(smtKvsData))

	// update smt local
	opt := smt.SmtOpt{GetProof: true, GetRoot: true}
	smtTree := reverse_smt.GetReverseSmt()
	smtOut, err := smtTree.UpdateMiddleSmt(smtKvs, opt)
	if err != nil {
		return nil, fmt.Errorf("UpdateMiddleSmt err: %s", err)
	}

	// smt_status=2 and tx_status=0
	if err := t.dbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusConfirm, tables.ReverseSmtTxStatusDefault, "id=?", id); err != nil {
		return nil, fmt.Errorf("UpdateReverseSmtTaskInfoStatus err: %s", err)
	}
	return smtOut, nil
}

// reverseSmtTaskRollback
func (t *TxTimer) reverseSmtTaskRollback() (bool, error) {
	// find all smt_status=1 and tx_status=0 and retry>=tables.ReverseSmtMaxRetryNum, rollback it
	rollbackTaskInfos, err := t.dbDao.FindReverseSmtTaskInfo(func(db *gorm.DB) *gorm.DB {
		return db.Where("smt_status=? and tx_status=? and retry>=?",
			tables.ReverseSmtStatusPending, tables.ReverseSmtTxStatusDefault, tables.ReverseSmtMaxRetryNum)
	})
	if err != nil {
		return false, fmt.Errorf("FindReverseSmtTaskInfo err: %s", err)
	}
	if len(rollbackTaskInfos) == 0 {
		return false, nil
	}

	rollbackTaskInfosData, _ := json.Marshal(rollbackTaskInfos)
	log.Infof("doReverseSmtTask reverseSmtTaskRollback: %s", string(rollbackTaskInfosData))

	// check reverse_smt_info latest smt_root == online_smt_root
	reverseRootCell, err := t.dasCore.GetReverseRecordSmtCell()
	if err != nil {
		return false, fmt.Errorf("GetReverseRecordSmtCell err: %s", err)
	}
	onlineSmtRoot := string(reverseRootCell.OutputData)

	rollbackKv := make([]smt.SmtKv, 0)
	for _, task := range rollbackTaskInfos {
		task.SmtStatus = tables.ReverseSmtStatusRollbackConfirm
		task.TxStatus = tables.ReverseSmtTxStatusReject

		recordInfos, err := t.dbDao.FindReverseSmtRecordInfoByTaskID(task.TaskID)
		if err != nil {
			return false, fmt.Errorf("FindReverseSmtRecordInfoByTaskID err: %s", err)
		}

		for _, record := range recordInfos {
			reverseSmtInfo, err := t.dbDao.GetLatestReverseSmtInfoByAddress(record.Address)
			if err != nil {
				return false, fmt.Errorf("GetLatestReverseSmtInfoByAddress err: %s", err)
			}
			smtKey, err := record.GetSmtKey()
			if err != nil {
				return false, fmt.Errorf("GetSmtKey err: %s", err)
			}

			leafDataHash := smt.H256Zero()
			if reverseSmtInfo.ID > 0 {
				leafDataHash = smt.ToSmtH256(reverseSmtInfo.LeafDataHash)
			}

			rollbackKv = append(rollbackKv, smt.SmtKv{
				Key:   smtKey,
				Value: leafDataHash,
			})
			log.Infof("doReverseSmtTask rollback reverse address: %s account: %s toLeafDataHash: %s", record.Address, record.Account, leafDataHash)
		}
	}

	opt := smt.SmtOpt{GetRoot: true}
	smtTree := reverse_smt.GetReverseSmt()

	// update smt_status=4 and tx_status=3
	if err := t.dbDao.DbTransaction(func(tx *gorm.DB) error {
		for _, task := range rollbackTaskInfos {
			if err := tx.Model(&tables.ReverseSmtTaskInfo{}).Where("id=?", task.ID).Updates(map[string]interface{}{
				"smt_status": tables.ReverseSmtStatusRollbackConfirm,
				"tx_status":  tables.ReverseSmtTxStatusReject,
			}).Error; err != nil {
				return err
			}
		}
		smtOut, err := smtTree.UpdateSmt(rollbackKv, opt)
		if err != nil {
			return fmt.Errorf("UpdateSmt err: %s", err)
		}
		if smtOut.Root.String() != onlineSmtRoot {
			log.Warnf("rollback warn, local smtRoot: %s != online smtRoot: %s", smtOut.Root, onlineSmtRoot)
		}
		return nil
	}); err != nil {
		return false, err
	}
	return false, nil
}

// reverseSmtCheckIsLatestTask check reverse_smt_info latest smt_root == online_smt_root
func (t *TxTimer) reverseSmtCheckIsLatestTask() (bool, error) {
	log.Info("doReverseSmtTask reverseSmtCheckIsLatestTask check smt local root == online_root")

	reverseInfo, err := t.dbDao.GetLatestReverseSmtInfo()
	if err != nil {
		return false, fmt.Errorf("GetLatestReverseSmtInfo err: %s", err)
	}
	if reverseInfo.ID == 0 {
		return false, nil
	}

	reverseRootCell, err := t.dasCore.GetReverseRecordSmtCell()
	if err != nil {
		return false, fmt.Errorf("GetReverseRecordSmtCell err: %s", err)
	}
	onlineSmtRoot := string(reverseRootCell.OutputData)

	if reverseInfo.RootHash != onlineSmtRoot {
		log.Warnf("doReverseSmtTask reverseSmtCheckIsLatestTask online_smt_root: %s != reverse_smt_info.smt_root: %s", onlineSmtRoot, reverseInfo.RootHash)
		return true, nil
	}
	return false, nil
}

// reverseSmtTaskAssignment
func (t *TxTimer) reverseSmtTaskAssignment() error {
	// smt_status=0, tx_status=0
	reverseRecordsByUnassigned, err := t.dbDao.FindReverseRecordInfoUnassigned(ReverseRecordMaxTaskNum)
	if err != nil {
		return fmt.Errorf("FindReverseRecordInfo err: %s", err)
	}
	if len(reverseRecordsByUnassigned) == 0 {
		return nil
	}

	// have ending task, don't assign tasks for now
	pendingTask, err := t.reverseSmtGetPendingTask()
	if err != nil {
		return fmt.Errorf("reverseSmtGetPendingTask err: %s", err)
	}
	if pendingTask.ID > 0 {
		return nil
	}

	// update to smt_status=1, tx_status=0
	taskInfo := &tables.ReverseSmtTaskInfo{
		SmtStatus: tables.ReverseSmtStatusPending,
	}
	if err := taskInfo.InitTaskId(); err != nil {
		return fmt.Errorf("InitTaskId err: %s", err)
	}
	if err := t.dbDao.DbTransaction(func(tx *gorm.DB) error {
		if err := tx.Create(taskInfo).Error; err != nil {
			return err
		}
		for _, v := range reverseRecordsByUnassigned {
			v.TaskID = taskInfo.TaskID
		}
		if err := tx.Save(reverseRecordsByUnassigned).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// reverseSmtGetPendingTask
func (t *TxTimer) reverseSmtGetPendingTask() (smtPendingTask tables.ReverseSmtTaskInfo, err error) {
	smtPendingTasks, err := t.dbDao.FindReverseSmtTaskInfo(func(db *gorm.DB) *gorm.DB {
		return db.Where(" ( smt_status!=? or tx_status!=? ) and smt_status!=? ",
			tables.ReverseSmtStatusRollbackConfirm, tables.ReverseSmtTxStatusConfirm, tables.ReverseSmtStatusRollbackConfirm).Limit(1)
	})
	if err != nil {
		err = fmt.Errorf("FindReverseSmtTaskInfo err: %s", err)
		return
	}
	if len(smtPendingTasks) == 0 {
		return
	}
	smtPendingTask = *smtPendingTasks[0]
	return
}

// reverseSmtAssemblyTx
func (t *TxTimer) reverseSmtAssemblyTx(reverseRecordSmtLiveCell *indexer.LiveCell, reverseRecordsByTaskID []*tables.ReverseSmtRecordInfo, smtOut *smt.UpdateMiddleSmtOut) (*txbuilder.BuildTransactionParams, error) {
	// balance check
	configCellBuilder, err := t.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		return nil, fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	feeCapacity, _ := configCellBuilder.RecordCommonFee()

	liveCells, totalCapacity, err := t.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          t.dasCache,
		LockScript:        t.reverseSmtServerScript,
		CapacityNeed:      feeCapacity,
		CapacityForChange: common.DasLockWithBalanceTypeOccupiedCkb,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCells err: %s", err)
	}

	// get pre tx
	preTx, err := t.dasCore.Client().GetTransaction(t.ctx, reverseRecordSmtLiveCell.OutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("preTx GetTransaction err: %s", err)
	}

	// Assembly transaction
	latestSmtKey, err := reverseRecordsByTaskID[len(reverseRecordsByTaskID)-1].GetSmtKey()
	if err != nil {
		return nil, fmt.Errorf("reverseRecordsByTaskID get latest smt key err: %s", err)
	}
	latestRootHash := smtOut.Roots[latestSmtKey.String()]

	txBuilderParams, err := reverse_smt.BuildReverseSmtTx(&reverse_smt.ReverseSmtParams{
		ServerLock:    t.reverseSmtServerScript,
		BalanceCells:  liveCells,
		TotalCapacity: totalCapacity,
		FeeCapacity:   feeCapacity,
		SmtRoot:       latestRootHash.String(),
		PreTx:         preTx,
	})
	if err != nil {
		return nil, fmt.Errorf("doReverseSmtTask BuildReverseSmtTx err: %s", err)
	}

	// append witness
	errWg := &errgroup.Group{}
	reverseSmtWitnessBuilder := witness.NewReverseSmtBuilder()

	oldWitnessLen := len(txBuilderParams.Witnesses)
	for range reverseRecordsByTaskID {
		txBuilderParams.Witnesses = append(txBuilderParams.Witnesses, nil)
	}

	for i, v := range reverseRecordsByTaskID {
		idx := i
		record := v
		errWg.Go(func() error {
			preRecord, err := t.dbDao.GetPreRecordByAddressAndNonce(record.Address, record.AlgorithmID, record.Nonce)
			if err != nil {
				return fmt.Errorf("GetPreRecordByAddressAndNonce err: %s", err)
			}

			smtKey, err := record.GetSmtKey()
			if err != nil {
				return fmt.Errorf("record GetSmtKey err: %s", err)
			}
			proof, ok := smtOut.Proofs[smtKey.String()]
			if !ok {
				return fmt.Errorf("GetProof err: %s not in smtOut", smtKey)
			}
			smtRoot, ok := smtOut.Roots[smtKey.String()]
			if !ok {
				return fmt.Errorf("GetRoots err: %s not in smtOut", smtKey)
			}

			witnessData, err := reverseSmtWitnessBuilder.GenWitness(&witness.ReverseSmtRecord{
				Version:     witness.ReverseSmtRecordVersion1,
				Action:      witness.ReverseSmtRecordAction(record.SubAction),
				Signature:   record.Sign,
				SignType:    record.AlgorithmID,
				Address:     common.Hex2Bytes(record.Address),
				Proof:       proof,
				PrevNonce:   preRecord.Nonce,
				PrevAccount: preRecord.Account,
				NextRoot:    smtRoot,
				NextAccount: record.Account,
			})
			if err != nil {
				return fmt.Errorf("GenWitness err: %s", err)
			}
			txBuilderParams.Witnesses[idx+oldWitnessLen] = witnessData
			return nil
		})
	}
	if err := errWg.Wait(); err != nil {
		return nil, err
	}
	return txBuilderParams, nil
}

// reverseSmtNeedProcess need goto process task
func (t *TxTimer) reverseSmtNeedProcess() (bool, error) {
	smtPendingTask, err := t.reverseSmtGetPendingTask()
	if err != nil {
		return false, fmt.Errorf("reverseSmtGetPendingTask err: %s", err)
	}
	if smtPendingTask.ID > 0 {
		return true, nil
	}

	// find have enough task record to process
	reverseRecord, err := t.dbDao.FindReverseRecordInfoUnassigned(ReverseRecordMaxTaskNum)
	if err != nil {
		return false, fmt.Errorf("FindReverseSmtTaskInfo err: %s", err)
	}
	if len(reverseRecord) >= ReverseRecordMaxTaskNum {
		return true, nil
	}
	return false, nil
}

// reverseSmtPendingCheck
func (t *TxTimer) reverseSmtPendingCheck(smtPendingTask *tables.ReverseSmtTaskInfo) (bool, error) {
	if smtPendingTask.SmtStatus != tables.ReverseSmtStatusConfirm {
		// should never into this case
		if smtPendingTask.SmtStatus != tables.ReverseSmtStatusPending ||
			smtPendingTask.TxStatus != tables.ReverseSmtTxStatusDefault {
			return false, fmt.Errorf("smt and tx status err, now smt_status=%d and tx_status=%d, by want smt_status=1 and tx_status=0", smtPendingTask.SmtStatus, smtPendingTask.TxStatus)
		}
		return false, nil
	}

	// mean tx no send, continue to assembly tx
	if smtPendingTask.Outpoint == "" {
		return false, nil
	}

	txHash := common.String2OutPointStruct(smtPendingTask.Outpoint).TxHash
	txStatus, err := t.dasCore.Client().GetTransaction(t.ctx, txHash)
	if err != nil {
		return false, fmt.Errorf("GetTransaction err: %s", err)
	}

	switch txStatus.TxStatus.Status {
	case types.TransactionStatusUnknown:
		return false, nil
	case types.TransactionStatusCommitted:
		err = t.dbDao.UpdateReverseSmtTaskInfo(map[string]interface{}{
			"smt_status": tables.ReverseSmtStatusConfirm,
			"tx_status":  tables.ReverseSmtTxStatusConfirm,
		}, "id=?", smtPendingTask.ID)
	case types.TransactionStatusPending, types.TransactionStatusProposed:
		err = t.dbDao.UpdateReverseSmtTaskInfo(map[string]interface{}{
			"smt_status": tables.ReverseSmtStatusConfirm,
			"tx_status":  tables.ReverseSmtTxStatusPending,
		}, "id=?", smtPendingTask.ID)
	case types.TransactionStatusRejected:
		err = t.dbDao.UpdateReverseSmtTaskInfo(map[string]interface{}{
			"smt_status": tables.ReverseSmtStatusRollback,
			"tx_status":  tables.ReverseSmtTxStatusReject,
		}, "id=?", smtPendingTask.ID)
	default:
		return false, fmt.Errorf("GetTransaction status unknown: %s", txStatus.TxStatus.Status)
	}
	if err != nil {
		return false, fmt.Errorf("update status err: %s", err)
	}
	return true, nil
}
