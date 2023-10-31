package dao

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
)

func (d *DbDao) GetPendingAuctionOrder(algorithmId common.DasAlgorithmId, subAlgithmId common.DasSubAlgorithmId, addr string) (list []tables.TableAuctionOrder, err error) {
	sql := fmt.Sprintf(`SELECT o.id,o.account,o.outpoint,o.basic_price,o.premium_price,o.order_id,p.status FROM %s o LEFT JOIN %s p ON o.outpoint=p.outpoint WHERE   o.algorithm_id = %d and o.sub_algorithm_id = %d and o.address = "%s" and p.status = 0 order by bid_time desc`, tables.TableNameAuctionOrder, tables.TableNameRegisterPendingInfo, algorithmId, subAlgithmId, addr)
	err = d.db.Raw(sql).Find(&list).Error
	return list, nil
}

func (d *DbDao) GetAuctionOrderByAccount(account string) (list []tables.TableAuctionOrder, err error) {
	sql := fmt.Sprintf(`SELECT o.account,o.outpoint,o.address,o.algorithm_id,o.sub_algorithm_id,p.status FROM %s o 
LEFT JOIN %s p ON o.outpoint=p.outpoint
WHERE o.account= "%s" and p.status != %d`, tables.TableNameAuctionOrder, tables.TableNameRegisterPendingInfo, account, tables.StatusRejected)
	err = d.db.Raw(sql).Find(&list).Error
	return list, nil
}

func (d *DbDao) GetAuctionOrderStatus(algorithmId common.DasAlgorithmId, subAlgithmId common.DasSubAlgorithmId, addr, account string) (list tables.TableAuctionOrder, err error) {
	sql := fmt.Sprintf(`SELECT o.id,o.account,o.outpoint,o.basic_price,o.premium_price,o.order_id,p.status FROM %s o LEFT JOIN %s p ON o.outpoint=p.outpoint WHERE o.account="%s" and o.algorithm_id = %d and o.sub_algorithm_id = %d and o.address = "%s" order by bid_time desc`, tables.TableNameAuctionOrder, tables.TableNameRegisterPendingInfo, account, algorithmId, subAlgithmId, addr)
	fmt.Println(sql)
	err = d.db.Raw(sql).First(&list).Error
	return list, nil
}
