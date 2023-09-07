package dao

import (
	"das_register_server/config"
	"das_register_server/tables"
	"github.com/dotbitHQ/das-lib/common"
	"gorm.io/gorm"
	"time"
)

func (d *DbDao) SearchAccountList(chainType common.ChainType, address string) (list []tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" owner_chain_type=? AND owner=? ", chainType, address).
		Or(" manager_chain_type=? AND manager=? ", chainType, address).
		Where(" status=? ", tables.AccountStatusNormal).
		Find(&list).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) SearchAccountListWithPage(chainType common.ChainType, address, keyword string, limit, offset int, category tables.Category) (list []tables.TableAccountInfo, err error) {
	db := d.parserDb.Where("((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?))", chainType, address, chainType, address)
	db = db.Where("status!=?", tables.AccountStatusOnCross)

	switch category {
	//case tables.CategoryDefault:
	case tables.CategoryMainAccount:
		db = db.Where("parent_account_id='' AND enable_sub_account=1")
	case tables.CategoryMainAccountDisableSecondLevelDID:
		db = db.Where("parent_account_id='' AND enable_sub_account=0")
	case tables.CategorySubAccount:
		db = db.Where("parent_account_id!=''")
	case tables.CategoryOnSale:
		db = db.Where("status=?", tables.AccountStatusOnSale)
	case tables.CategoryExpireSoon:
		expiredAt := time.Now().Unix()
		expiredAt30Days := time.Now().Add(time.Hour * 24 * 30).Unix()
		db = db.Where("expired_at>=? AND expired_at<=?", expiredAt, expiredAt30Days)
	case tables.CategoryToBeRecycled:
		expiredAt := time.Now().Unix()
		recycledAt := time.Now().Add(-time.Hour * 24 * 90).Unix()
		if config.Cfg.Server.Net != common.DasNetTypeMainNet {
			recycledAt = time.Now().Add(-time.Hour * 24 * 30).Unix()
		}
		db = db.Where("expired_at<=? AND expired_at>=?", expiredAt, recycledAt)
	}

	if keyword != "" {
		db = db.Where("account LIKE ?", "%"+keyword+"%")
	}
	err = db.Order("expired_at").Limit(limit).Offset(offset).Find(&list).Error

	return
}

func (d *DbDao) GetAccountsCount(chainType common.ChainType, address, keyword string, category tables.Category) (count int64, err error) {
	db := d.parserDb.Model(tables.TableAccountInfo{}).Where("((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?))", chainType, address, chainType, address)
	db = db.Where("status!=?", tables.AccountStatusOnCross)

	switch category {
	//case tables.CategoryDefault:
	case tables.CategoryMainAccount:
		db = db.Where("parent_account_id='' AND enable_sub_account=1")
	case tables.CategoryMainAccountDisableSecondLevelDID:
		db = db.Where("parent_account_id='' AND enable_sub_account=0")
	case tables.CategorySubAccount:
		db = db.Where("parent_account_id!=''")
	case tables.CategoryOnSale:
		db = db.Where("status=?", tables.AccountStatusOnSale)
	case tables.CategoryExpireSoon:
		expiredAt := time.Now().Unix()
		expiredAt30Days := time.Now().Add(time.Hour * 24 * 30).Unix()
		db = db.Where("expired_at>=? AND expired_at<=?", expiredAt, expiredAt30Days)
	case tables.CategoryToBeRecycled:
		expiredAt := time.Now().Unix()
		recycledAt := time.Now().Add(-time.Hour * 24 * 90).Unix()
		if config.Cfg.Server.Net != common.DasNetTypeMainNet {
			recycledAt = time.Now().Add(-time.Hour * 24 * 30).Unix()
		}
		db = db.Where("expired_at<=? AND expired_at>=?", expiredAt, recycledAt)
	}

	if keyword != "" {
		db = db.Where("account LIKE ?", "%"+keyword+"%")
	}
	err = db.Count(&count).Error

	return
}

func (d *DbDao) GetAccounts(accounts []string) (list []tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" account IN(?) ", accounts).Find(&list).Error
	return
}

func (d *DbDao) GetAccountInfoByAccountId(accountId string) (acc tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" account_id=? ", accountId).Find(&acc).Error
	return
}

func (d *DbDao) GetAccountInfoByAccountIds(accountIds []string) (list []tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" account_id IN(?) ", accountIds).Find(&list).Error
	return
}

func (d *DbDao) GetNameDaoAccountInfoByAccountIds(accountIds []string) (list []tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" parent_account_id='' AND account_id IN(?) ", accountIds).Find(&list).Error
	return
}

func (d *DbDao) GetPreAccount(accountId string) (info tables.TableAccountInfo, err error) {
	err = d.parserDb.Where("parent_account_id='' AND account_id<?", accountId).
		Order("account_id DESC").Limit(1).Find(&info).Error
	return
}
