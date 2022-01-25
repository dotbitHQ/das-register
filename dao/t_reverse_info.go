package dao

import (
	"das_register_server/tables"
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm"
)

func (d *DbDao) SearchLatestReverse(chainType common.ChainType, address string) (reverse tables.TableReverseInfo, err error) {
	err = d.parserDb.Where(" chain_type=? AND address=? ", chainType, address).Order(" block_number DESC, outpoint DESC ").First(&reverse).Error
	return
}

func (d *DbDao) SearchReverseList(chainType common.ChainType, address string) (list []tables.TableReverseInfo, err error) {
	err = d.parserDb.Where(" chain_type=? AND address=? ", chainType, address).Order(" block_number DESC, outpoint DESC ").Find(&list).Error
	if err == gorm.ErrRecordNotFound {
		err = nil
	}
	return
}

func (d *DbDao) SearchReverseByAccount(chainType common.ChainType, address, account string) (reverse tables.TableReverseInfo, err error) {
	err = d.parserDb.Where(" chain_type=? AND address=? AND account=? ", chainType, address, account).Order(" block_number DESC, outpoint DESC ").First(&reverse).Error
	return
}
