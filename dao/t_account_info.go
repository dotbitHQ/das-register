package dao

import (
	"das_register_server/tables"
	"github.com/DeAccountSystems/das-lib/common"
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
	switch category {
	default:
		//case tables.CategoryDefault:
		if keyword == "" {
			err = d.parserDb.Where(" owner_chain_type=? AND owner=? ", chainType, address).
				Or(" manager_chain_type=? AND manager=? ", chainType, address).
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		} else {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND account LIKE ? ",
				chainType, address, chainType, address, "%"+keyword+"%").
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		}
	case tables.CategoryMainAccount:
		if keyword == "" {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id='' ",
				chainType, address, chainType, address).
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		} else {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id='' AND account LIKE ? ",
				chainType, address, chainType, address, "%"+keyword+"%").
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		}
	case tables.CategorySubAccount:
		if keyword == "" {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id!='' ",
				chainType, address, chainType, address).
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		} else {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id!='' AND account LIKE ? ",
				chainType, address, chainType, address, "%"+keyword+"%").
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		}
	case tables.CategoryOnSale:
		if keyword == "" {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND status=? ",
				chainType, address, chainType, address, tables.AccountStatusOnSale).
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		} else {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND status=? AND account LIKE ? ",
				chainType, address, chainType, address, tables.AccountStatusOnSale, "%"+keyword+"%").
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		}
	case tables.CategoryExpireSoon:
		expiredAt := time.Now().Unix()
		expiredAt30Days := time.Now().Add(time.Hour * 24 * 30).Unix()
		if keyword == "" {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND expired_at>=? AND expired_at<=? ",
				chainType, address, chainType, address, expiredAt, expiredAt30Days).
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		} else {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND expired_at>=? AND expired_at<=? AND account LIKE ? ",
				chainType, address, chainType, address, expiredAt, expiredAt30Days, "%"+keyword+"%").
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		}
	case tables.CategoryToBeRecycled:
		expiredAt := time.Now().Unix()
		if keyword == "" {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND expired_at<=? ",
				chainType, address, chainType, address, expiredAt).
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		} else {
			err = d.parserDb.Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) expired_at<=? AND account LIKE ? ",
				chainType, address, chainType, address, expiredAt, "%"+keyword+"%").
				Order(" account ").
				Limit(limit).Offset(offset).
				Find(&list).Error
		}
	}

	return
}

func (d *DbDao) GetAccountsCount(chainType common.ChainType, address, keyword string, category tables.Category) (count int64, err error) {
	switch category {
	default:
		//case tables.CategoryDefault:
		if keyword == "" {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" owner_chain_type=? AND owner=? ", chainType, address).
				Or(" manager_chain_type=? AND manager=? ", chainType, address).
				Count(&count).Error
		} else {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND account LIKE ? ",
				chainType, address, chainType, address, "%"+keyword+"%").
				Count(&count).Error
		}
	case tables.CategoryMainAccount:
		if keyword == "" {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id='' ",
				chainType, address, chainType, address).
				Count(&count).Error
		} else {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id='' AND account LIKE ? ",
				chainType, address, chainType, address, "%"+keyword+"%").
				Count(&count).Error
		}
	case tables.CategorySubAccount:
		if keyword == "" {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id!='' ",
				chainType, address, chainType, address).
				Count(&count).Error
		} else {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND parent_account_id!='' AND account LIKE ? ",
				chainType, address, chainType, address, "%"+keyword+"%").
				Count(&count).Error
		}
	case tables.CategoryOnSale:
		if keyword == "" {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND status=? ",
				chainType, address, chainType, address, tables.AccountStatusOnSale).
				Count(&count).Error
		} else {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND status=? AND account LIKE ? ",
				chainType, address, chainType, address, tables.AccountStatusOnSale, "%"+keyword+"%").
				Count(&count).Error
		}
	case tables.CategoryExpireSoon:
		expiredAt := time.Now().Unix()
		expiredAt30Days := time.Now().Add(time.Hour * 24 * 30).Unix()
		if keyword == "" {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND expired_at>=? AND expired_at<=? ",
				chainType, address, chainType, address, expiredAt, expiredAt30Days).
				Count(&count).Error
		} else {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND expired_at>=? AND expired_at<=? AND account LIKE ? ",
				chainType, address, chainType, address, expiredAt, expiredAt30Days, "%"+keyword+"%").
				Count(&count).Error
		}
	case tables.CategoryToBeRecycled:
		expiredAt := time.Now().Unix()
		if keyword == "" {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) AND expired_at<=? ",
				chainType, address, chainType, address, expiredAt).
				Count(&count).Error
		} else {
			err = d.parserDb.Model(tables.TableAccountInfo{}).Where(" ((owner_chain_type=? AND owner=?)OR(manager_chain_type=? AND manager=?)) expired_at<=? AND account LIKE ? ",
				chainType, address, chainType, address, expiredAt, "%"+keyword+"%").
				Count(&count).Error
		}
	}

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
