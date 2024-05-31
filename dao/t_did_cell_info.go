package dao

import (
	"das_register_server/tables"
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

func (d *DbDao) GetAccountUpgradableList(chainType common.ChainType, address, keyword string, limit, offset int) (list []tables.TableAccountInfo, err error) {
	db := d.parserDb.Where("owner_chain_type=? AND owner=?", chainType, address)
	db = db.Where("status!=? and status!=? and expired_at >= ?", tables.AccountStatusOnCross, tables.AccountStatusOnUpgrade, time.Now().Unix())
	db = db.Where("parent_account_id=''")

	if keyword != "" {
		db = db.Where("account LIKE ?", "%"+keyword+"%")
	}
	err = db.Order("expired_at").Limit(limit).Offset(offset).Find(&list).Error

	return
}

func (d *DbDao) GetAccountUpgradableListTotal(chainType common.ChainType, address, keyword string) (count int64, err error) {
	db := d.parserDb.Model(tables.TableAccountInfo{}).
		Where("owner_chain_type=? AND owner=?", chainType, address)
	db = db.Where("status!=? and status!=? and expired_at >= ?", tables.AccountStatusOnCross, tables.AccountStatusOnUpgrade, time.Now().Unix())
	db = db.Where("parent_account_id=''")

	if keyword != "" {
		db = db.Where("account LIKE ?", "%"+keyword+"%")
	}
	err = db.Count(&count).Error

	return
}

func (d *DbDao) GetUpgradeOrder(accountIds []string) (list []tables.TableDasOrderInfo, err error) {
	if len(accountIds) == 0 {
		return
	}
	err = d.db.Where("account_id IN(?) AND action=? AND pay_status=? AND order_status=?",
		accountIds, common.DasActionTransferAccount, tables.TxStatusSending, tables.OrderStatusDefault).Find(&list).Error
	return
}

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
