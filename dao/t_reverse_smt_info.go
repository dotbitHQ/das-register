package dao

import (
	"das_register_server/tables"
)

func (d *DbDao) GetLatestReverseSmtInfo() (reverse tables.ReverseSmtInfo, err error) {
	err = d.parserDb.Order(" id DESC").First(&reverse).Error
	return
}
