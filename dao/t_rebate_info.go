package dao

import (
	"das_register_server/tables"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (d *DbDao) GetMyRewards(chainType common.ChainType, address string, serviceType, rewardType, limit, offset int) (list []tables.TableRebateInfo, err error) {
	err = d.parserDb.Where(" inviter_chain_type=? AND inviter_address=? AND service_type=? AND reward_type=? ", chainType, address, serviceType, rewardType).
		Order("id DESC").Limit(limit).Offset(offset).Find(&list).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

type RewardsCount struct {
	CountNumber int64           `json:"count_number" gorm:"column:count_number"`
	TotalReward decimal.Decimal `json:"total_reward" gorm:"column:total_reward"`
}

func (d *DbDao) GetMyRewardsCount(chainType common.ChainType, address string, serviceType, rewardType int) (rc RewardsCount, err error) {
	err = d.parserDb.Model(tables.TableRebateInfo{}).
		Select(" count(*) AS count_number,SUM(reward) AS total_reward ").
		Where(" inviter_chain_type=? AND inviter_address=? AND service_type=? AND reward_type=? ", chainType, address, serviceType, rewardType).
		Find(&rc).Error
	return
}
