package dao

import (
	"das_register_server/tables"
)

func (d *DbDao) SearchReverseSmtTaskInfo(outpoint string) (reverse tables.ReverseSmtTaskInfo, err error) {
	err = d.db.Where(" outpoint=? ", outpoint).Order(" block_number DESC, outpoint DESC ").First(&reverse).Error
	return
}

func (d *DbDao) UpdateReverseSmtTaskInfoStatus(outpoint string, smtStatus, txStatus int) (err error) {
	err = d.db.Where(" outpoint=? ", outpoint).Updates(map[string]interface{}{
		"smt_status": smtStatus,
		"tx_status":  txStatus,
	}).Error
	return
}
