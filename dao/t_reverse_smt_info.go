package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm"
)

func (d *DbDao) GetLatestReverseSmtInfo() (reverseInfo tables.ReverseSmtInfo, err error) {
	err = d.parserDb.Order(" id DESC ").First(&reverseInfo).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) GetLatestReverseSmtInfoByAddress(address string) (reverseInfo tables.ReverseSmtInfo, err error) {
	err = d.parserDb.Where(" address=? ", address).Order(" block_number DESC, outpoint DESC ").First(&reverseInfo).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}
