package timer

import (
	"das_register_server/internal/reverse_smt"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/smt"
)

const (
	ReverseRecordMaxTaskNum = 300
)

func (t *TxTimer) doReverseSmtTask() error {
	if err := t.dbDao.UpdateReverseSmtRollbackToTxPending(); err != nil {
		return fmt.Errorf("UpdateReverseSmtToTxPending err: %s", err)
	}

	opt := smt.SmtOpt{GetProof: true, GetRoot: true}
	smtTree := reverse_smt.GetReverseSmt()

	// smt_status=0, tx_status=0
	reverseRecords, err := t.dbDao.FindReverseRecordInfoUnassigned(ReverseRecordMaxTaskNum)
	if err != nil {
		return fmt.Errorf("FindReverseRecordInfo err: %s", err)
	}
	if len(reverseRecords) > 0 {
		// smt_status=1, tx_status=0
		taskInfo := &tables.ReverseSmtTaskInfo{
			SmtStatus: tables.ReverseSmtStatusPending,
		}
		if err := taskInfo.InitTaskId(); err != nil {
			return err
		}
		if err := t.dbDao.CreateSmtTaskInfo(taskInfo); err != nil {
			return err
		}
		for _, v := range reverseRecords {
			v.TaskID = taskInfo.TaskID
		}
		if err := t.dbDao.SaveReverseSmtRecordInfo(reverseRecords); err != nil {
			return err
		}
	}

	// find smt_status=3 and tx_status=3
	rejectSmtTasks, err := t.dbDao.FindReverseSmtTaskInfoByStatus(tables.ReverseSmtStatusRollback, tables.ReverseSmtTxStatusReject)
	if err != nil {
		return err
	}
	if err := t.dbDao.UpdateReverseSmtStatusRejectToPending(rejectSmtTasks); err != nil {
		return err
	}

	// find all smt_status=1 and tx_status=0
	smtPendingTasks, err := t.dbDao.FindReverseSmtTaskInfoByStatus(tables.ReverseSmtStatusPending, tables.ReverseSmtTxStatusDefault)
	if err != nil {
		return err
	}
	for _, taskInfo := range smtPendingTasks {
		reverseRecords, err := t.dbDao.FindReverseSmtRecordInfoByTaskID(taskInfo.TaskID)
		if err != nil {
			return err
		}

		smtKvs := make([]smt.SmtKv, 0)
		for _, record := range reverseRecords {
			smtKey, err := record.GetSmtKey()
			if err != nil {
				return err
			}
			smtVal, err := record.GetSmtValue()
			if err != nil {
				return err
			}
			smtKvs = append(smtKvs, smt.SmtKv{
				Key:   smtKey,
				Value: smtVal,
			})
		}

		smtOut, err := smtTree.UpdateMiddleSmt(smtKvs, opt)
		if err != nil {
			return err
		}
		if err := t.dbDao.UpdateReverseSmtTaskInfoStatus(tables.ReverseSmtStatusConfirm, tables.ReverseSmtTxStatusPending, "id=?", taskInfo.ID); err != nil {
			return err
		}

		// Assembly transaction
		reverse_smt.BuildReverseSmtTx(&reverse_smt.ReverseSmtParams{
			DasLock:       nil,
			DasType:       nil,
			BalanceCells:  nil,
			TotalCapacity: 0,
			FeeCapacity:   0,
			SmtRoot:       "",
			LatestTx:      nil,
		})
	}
	return nil
}
