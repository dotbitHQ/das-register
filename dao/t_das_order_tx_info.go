package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetOrderTxListByOrderId(orderId string) (list []tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("order_id=?", orderId).Order("id DESC").Find(&list).Error
	return
}

func (d *DbDao) CreateOrderTx(orderTx *tables.TableDasOrderTxInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"action", "status", "timestamp",
		}),
	}).Create(&orderTx).Error
}

func (d *DbDao) GetOrderTxByHash(action tables.OrderTxAction, hash string) (tx tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("action=? AND hash=?", action, hash).Limit(1).Find(&tx).Error
	return
}

func (d *DbDao) GetOrderTxByActionAndHashList(action tables.OrderTxAction, hashList []string) (list []tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("action=? AND hash IN(?)", action, hashList).Find(&list).Error
	return
}

func (d *DbDao) GetOrderTxByAction(orderId string, action tables.OrderTxAction) (tx tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("order_id=? AND action=? AND status=?",
		orderId, action, tables.OrderTxStatusConfirm).Order("id DESC").
		Limit(1).Find(&tx).Error
	return
}

func (d *DbDao) GetMaybeRejectedRegisterTxs(timestamp int64) (list []tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("timestamp<? AND status=?", timestamp, tables.OrderTxStatusDefault).Find(&list).Error
	return
}
