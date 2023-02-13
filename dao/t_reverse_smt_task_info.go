package dao

import (
	"das_register_server/tables"
	"fmt"
	"gorm.io/gorm"
)

func (d *DbDao) SearchReverseSmtTaskInfo(outpoint string) (reverse tables.ReverseSmtTaskInfo, err error) {
	err = d.db.Where(" outpoint=? ", outpoint).Order(" block_number DESC, outpoint DESC ").First(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) FindReverseSmtTaskInfo(fn func(db *gorm.DB) *gorm.DB) (reverse []*tables.ReverseSmtTaskInfo, err error) {
	err = fn(d.db).Find(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) UpdateReverseSmtTaskInfoStatus(smtStatus, txStatus int, where string, args ...interface{}) (err error) {
	err = d.db.Model(&tables.ReverseSmtTaskInfo{}).Where(where, args...).Updates(map[string]interface{}{
		"smt_status": smtStatus,
		"tx_status":  txStatus,
	}).Error
	return
}

func (d *DbDao) UpdateAllReverseSmtRollbackToTxPending() (err error) {
	err = d.db.Exec(fmt.Sprintf("update %s set smt_status=?, tx_status=?, retry=retry+1 where smt_status=? AND tx_status=?", tables.TableNameReverseSmtTaskInfo),
		tables.ReverseSmtStatusPending, tables.ReverseSmtTxStatusDefault, tables.ReverseSmtStatusRollback, tables.ReverseSmtTxStatusReject).Error
	return
}

func (d *DbDao) CreateSmtTaskInfo(smtTask *tables.ReverseSmtTaskInfo) (err error) {
	err = d.db.Create(smtTask).Error
	return
}

func (d *DbDao) UpdateReverseSmtStatusRejectToPending(tasks []*tables.ReverseSmtTaskInfo) (err error) {
	err = d.db.Transaction(func(tx *gorm.DB) error {
		for _, v := range tasks {
			if err := tx.Exec(fmt.Sprintf("update %s set retry=retry+1 where task_id=?", tables.TableNameReverseSmtRecordInfo), v.TaskID).Error; err != nil {
				return err
			}
			if err := tx.Model(&tables.ReverseSmtTaskInfo{}).Where("id=?", v.ID).Updates(map[string]interface{}{
				"smt_status": tables.ReverseSmtStatusPending,
				"tx_status":  tables.ReverseSmtTxStatusDefault,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func (d *DbDao) UpdateReverseSmtTaskInfo(updates map[string]interface{}, where string, args ...interface{}) (err error) {
	err = d.db.Model(&tables.ReverseSmtTaskInfo{}).Where(where, args...).Updates(updates).Error
	return
}

func (d *DbDao) GetLatestReverseSmtInfoByTaskID(taskID string) (reverseInfo tables.ReverseSmtTaskInfo, err error) {
	err = d.db.Where(" task_id=? ", taskID).First(&reverseInfo).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}
