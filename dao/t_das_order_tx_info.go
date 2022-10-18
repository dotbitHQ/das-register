package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm"
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

func (d *DbDao) GetMaybeRejectedRegisterTxs(start, end int64) (list []tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("timestamp>? AND timestamp<? AND status=?", start, end, tables.OrderTxStatusDefault).Find(&list).Error
	return
}

func (d *DbDao) UpdateRejectedTx(action tables.OrderTxAction, orderId string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderTxInfo{}).
			Where("order_id=? AND status=?", orderId, tables.OrderTxStatusDefault).
			Updates(map[string]interface{}{
				"status": tables.OrderTxStatusRejected,
			}).Error; err != nil {
			return err
		}

		switch action {
		case tables.TxActionApplyRegister, tables.TxActionRenewAccount:
			if err := tx.Model(tables.TableDasOrderInfo{}).
				Where("order_id=? AND pay_status=? AND order_status=?", orderId, tables.TxStatusOk, tables.OrderStatusDefault).
				Updates(map[string]interface{}{
					"pay_status": tables.TxStatusSending,
				}).Error; err != nil {
				return err
			}
		case tables.TxActionPreRegister:
			if err := tx.Model(tables.TableDasOrderInfo{}).
				Where("order_id=? AND pre_register_status=? AND order_status=?", orderId, tables.TxStatusOk, tables.OrderStatusDefault).
				Updates(map[string]interface{}{
					"pre_register_status": tables.TxStatusSending,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
