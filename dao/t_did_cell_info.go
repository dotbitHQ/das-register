package dao

import (
	"das_register_server/tables"
)

func (d *DbDao) GetDidAccountList(args string, limit, offset int) (list []tables.TableDidCellInfo, err error) {
	expiredAt := tables.GetDidCellRecycleExpiredAt()
	err = d.parserDb.Where("args=? AND expired_at>=?", args, expiredAt).
		Limit(limit).Offset(offset).Find(&list).Error
	return
}

func (d *DbDao) GetDidAccountListTotal(args string) (count int64, err error) {
	expiredAt := tables.GetDidCellRecycleExpiredAt()
	err = d.parserDb.Model(tables.TableDidCellInfo{}).
		Where("args=? AND expired_at>=?", args, expiredAt).Count(&count).Error
	return
}

func (d *DbDao) GetDidAccountByAccountId(accountId, args string) (info tables.TableDidCellInfo, err error) {
	err = d.parserDb.Where("account_id=? AND args=?",
		accountId, args).Order("expired_at").Find(&info).Error
	return
}
