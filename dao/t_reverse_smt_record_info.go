package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm"
)

func (d *DbDao) FindReverseRecordInfoUnassigned(limit int) (reverse []*tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" task_id='' ").Limit(limit).Find(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) FindReverseSmtRecordInfoByTaskID(taskID string) (reverse []*tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" task_id=? ", taskID).Find(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) GetPreRecordByAddressAndNonce(address string, nonce int) (reverse tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" address=? and nonce=? ", address, nonce-1).Order("id desc").First(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}
