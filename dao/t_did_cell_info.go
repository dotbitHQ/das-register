package dao

import (
	"das_register_server/tables"
	"time"
)

func (d *DbDao) GetDidAccountList(args string, limit, offset int) (list []tables.TableDidCellInfo, err error) {
	expiredAt := time.Now().Unix() - 90*86400
	err = d.parserDb.Where("args=? AND expired_at>=?", args, expiredAt).
		Limit(limit).Offset(offset).Find(&list).Error
	return
}

func (d *DbDao) GetDidAccountListTotal(args string) (count int64, err error) {
	expiredAt := time.Now().Unix() - 90*86400
	err = d.db.Model(tables.TableDidCellInfo{}).
		Where("args=? AND expired_at>=?", args, expiredAt).Count(&count).Error
	return
}
