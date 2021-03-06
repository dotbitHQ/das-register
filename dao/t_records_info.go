package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm"
)

func (d *DbDao) SearchAccountReverseRecords(account, address string) (record tables.TableRecordsInfo, err error) {
	err = d.parserDb.Where(" account=? AND type='address' AND value=? ", account, address).Find(&record).Error
	return
}

func (d *DbDao) SearchRecordsByAccount(accountId string) (list []tables.TableRecordsInfo, err error) {
	err = d.parserDb.Where(" account_id=? ", accountId).Find(&list).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}
