package dao

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
)

func (d *DbDao) GetPendingAuctionOrder(chainType common.ChainType, addr string) (list []tables.TableAuctionOrder, err error) {
	sql := fmt.Sprintf(`SELECT o.id,o.account,o.outpoint,o.basic_price,o.premium_price,o.order_id,p.status FROM %s o LEFT JOIN %s p ON o.outpoint=p.outpoint WHERE   o.chain_type = %d  and o.address = "%s" and p.status = 0 order by bid_time desc`, tables.TableNameAuctionOrder, tables.TableNameRegisterPendingInfo, chainType, addr)
	err = d.db.Raw(sql).Find(&list).Error
	return list, nil
}

func (d *DbDao) GetAuctionOrderByAccount(account string, createTime int64) (list []tables.TableAuctionOrder, err error) {
	sql := fmt.Sprintf(`SELECT o.account,o.outpoint,o.address,o.chain_type,p.status FROM %s o 
LEFT JOIN %s p ON o.outpoint=p.outpoint
WHERE o.account= "%s" and p.status != %d and o.created_at > %d `, tables.TableNameAuctionOrder, tables.TableNameRegisterPendingInfo, account, tables.StatusRejected, createTime)
	err = d.db.Raw(sql).Find(&list).Error
	return list, nil
}

func (d *DbDao) GetAuctionOrderStatus(chainType common.ChainType, addr, hash string) (list tables.TableAuctionOrder, err error) {
	sql := fmt.Sprintf(`SELECT o.id,o.account,o.outpoint,o.basic_price,o.premium_price,o.order_id,p.status FROM %s o LEFT JOIN %s p ON o.outpoint=p.outpoint WHERE p.outpoint="%s" and o.chain_type = %d and o.address = "%s" order by bid_time desc`, tables.TableNameAuctionOrder, tables.TableNameRegisterPendingInfo, fmt.Sprintf("%s-0", hash), chainType, addr)
	err = d.db.Raw(sql).First(&list).Error
	return list, nil
}

func (d *DbDao) CreateAuctionOrder(auctionOrder tables.TableAuctionOrder) (err error) {
	return d.db.Create(&auctionOrder).Error
}
