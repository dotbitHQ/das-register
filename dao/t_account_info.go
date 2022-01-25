package dao

import (
	"das_register_server/tables"
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm"
)

func (d *DbDao) SearchAccount(account string) (acc tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" account=? ", account).First(&acc).Error
	return
}

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

func (d *DbDao) SearchAccountListWithPage(chainType common.ChainType, address string, limit, offset int) (list []tables.TableAccountInfo, err error) {
	err = d.parserDb.Where(" owner_chain_type=? AND owner=? ", chainType, address).
		Or(" manager_chain_type=? AND manager=? ", chainType, address).
		Order(" account ").
		Limit(limit).Offset(offset).
		Find(&list).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) GetAccountsCount(chainType common.ChainType, address string) (count int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).
		Where(" owner_chain_type=? AND owner=? ", chainType, address).
		Or(" manager_chain_type=? AND manager=? ", chainType, address).
		Count(&count).Error
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
