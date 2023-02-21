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

func (d *DbDao) GetPreRecordByAddressAndNonce(address string, algorithmId uint8, nonce uint32) (reverse tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" address=? and algorithm_id=? and nonce=?", address, algorithmId, nonce-1).
		Order("id desc").First(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) GetReverseSmtRecordByAddress(address string, algorithmId uint8) (reverse tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" address=? and algorithm_id=?", address, algorithmId).Order("id desc").First(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) CreateReverseSmtRecord(reverse *tables.ReverseSmtRecordInfo) (err error) {
	err = d.db.Create(reverse).Error
	return
}
