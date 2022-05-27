package dao

import (
	"das_register_server/tables"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	actionList = []string{
		common.DasActionTransferAccount,
		common.DasActionEditManager,
		common.DasActionEditRecords,
		common.DasActionRedeclareReverseRecord,
		common.DasActionEditAccountSale,
		common.DasActionConfirmProposal,
		common.DasActionRenewAccount,
		tables.DasActionOfferAccepted,
		common.DasActionEditSubAccount,
		tables.DasActionBalanceDeposit,
		common.DasActionLockAccountForCrossChain,
	}
)

func (d *DbDao) GetTransactionList(chainType common.ChainType, address string, limit, offset int) (list []tables.TableTransactionInfo, err error) {

	err = d.parserDb.Where(" chain_type=? AND address=? AND action NOT IN(?) ", chainType, address, actionList).
		Order(" id DESC ").
		Limit(limit).Offset(offset).
		Find(&list).Error

	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) GetTransactionListByOutpoints(outpoints []string) (list []tables.TableTransactionInfo, err error) {
	err = d.parserDb.Where(" outpoint IN(?) ", outpoints).Find(&list).Error
	return
}

func (d *DbDao) GetTransactionListTotal(chainType common.ChainType, address string) (count int64, err error) {
	err = d.parserDb.Model(tables.TableTransactionInfo{}).Where(" chain_type=? AND address=? AND action NOT IN(?) ", chainType, address, actionList).Count(&count).Error
	return
}

func (d *DbDao) GetTransactionListByAction(chainType common.ChainType, address, action string, limit, offset int) (list []tables.TableTransactionInfo, err error) {
	err = d.parserDb.Where(" chain_type=? AND address=? AND action=? ", chainType, address, action).
		Order("id DESC").
		Limit(limit).Offset(offset).Find(&list).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

type TransactionTotal struct {
	CountNumber   int64           `json:"count_number" gorm:"column:count_number"`
	TotalCapacity decimal.Decimal `json:"total_capacity" gorm:"column:total_capacity"`
}

func (d *DbDao) GetTransactionTotalCapacityByAction(chainType common.ChainType, address, action string) (tt TransactionTotal, err error) {
	err = d.parserDb.Model(tables.TableTransactionInfo{}).
		Select(" COUNT(*) AS count_number,SUM(capacity) AS total_capacity ").
		Where(" chain_type=? AND address=? AND action=? ", chainType, address, action).
		Find(&tt).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) GetAccountTransactionByActions(accountId string, actions []common.DasAction) (t tables.TableTransactionInfo, err error) {
	err = d.parserDb.Where("account_id=? AND action IN(?)", accountId, actions).
		Order("id DESC").Limit(1).Find(&t).Error
	return
}
