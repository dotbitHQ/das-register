package dao

import (
	"das_register_server/tables"
)

func (d *DbDao) GetDidAccountList(args, keyword string, limit, offset int) (list []tables.TableDidCellInfo, err error) {
	expiredAt := tables.GetDidCellRecycleExpiredAt()

	db := d.parserDb.Where("args=? AND expired_at>=?", args, expiredAt)
	if keyword != "" {
		db = db.Where("account LIKE ?", "%"+keyword+"%")
	}
	err = db.Limit(limit).Offset(offset).Find(&list).Error
	return
}

func (d *DbDao) GetDidAccountListTotal(args, keyword string) (count int64, err error) {
	expiredAt := tables.GetDidCellRecycleExpiredAt()

	db := d.parserDb.Model(tables.TableDidCellInfo{}).Where("args=? AND expired_at>=?", args, expiredAt)
	if keyword != "" {
		db = db.Where("account LIKE ?", "%"+keyword+"%")
	}
	err = db.Count(&count).Error
	return
}

func (d *DbDao) GetDidAccountByAccountId(accountId, args string) (info tables.TableDidCellInfo, err error) {
	err = d.parserDb.Where("account_id=? AND args=?",
		accountId, args).Order("expired_at DESC").Limit(1).Find(&info).Error
	return
}

func (d *DbDao) GetDidAccountByAccountIdWithoutArgs(accountId string) (info tables.TableDidCellInfo, err error) {
	err = d.parserDb.Where("account_id=?", accountId).
		Order("expired_at DESC").Limit(1).Find(&info).Error
	return
}
