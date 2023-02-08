package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm"
)

func (d *DbDao) GetLatestReverseSmtInfo() (reverse tables.ReverseSmtInfo, err error) {
	err = d.parserDb.Order(" id DESC").First(&reverse).Error
	return
}

func (d *DbDao) FindReverseRecordInfoUnassigned(limit int) (reverse []*tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" task_id='' ").Limit(limit).Find(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) SaveReverseSmtRecordInfo(reverse []*tables.ReverseSmtRecordInfo) (err error) {
	err = d.db.Where(" task_id='' ").Save(reverse).Error
	return
}

func (d *DbDao) FindReverseSmtRecordInfoByTaskID(taskID string) (reverse []*tables.ReverseSmtRecordInfo, err error) {
	err = d.db.Where(" task_id=? ", taskID).Find(&reverse).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}
