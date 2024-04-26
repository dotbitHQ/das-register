package dao

import "das_register_server/tables"

func (d *DbDao) GetDidAccountList(args string, limit, offset int) (list []tables.TableDidCellInfo, err error) {
	err = d.parserDb.Where("args=?", args).
		Limit(limit).Offset(offset).Find(&list).Error
	return
}

func (d *DbDao) GetDidAccountListTotal(args string) (count int64, err error) {
	err = d.db.Model(tables.TableDidCellInfo{}).
		Where("args=?", args).Count(&count).Error
	return
}
