package dao

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func (d *DbDao) GetLatestRegisterOrderByAddress(chainType common.ChainType, address, accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("chain_type=? AND address=? AND account_id=? AND action=?",
		chainType, address, accountId, common.DasActionApplyRegister).
		Order("order_status,register_status DESC,id DESC").Limit(1).Find(&order).Error
	return
}

func (d *DbDao) GetLatestRegisterOrderByLatest(accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("account_id=? AND action=? AND order_status=?",
		accountId, common.DasActionApplyRegister, tables.OrderStatusDefault).
		Order("order_status,register_status DESC,id DESC").Limit(1).Find(&order).Error
	return
}

func (d *DbDao) GetRegisteringOrders(chainType common.ChainType, address string) (list []tables.TableDasOrderInfo, err error) {
	// SELECT account,MAX(register_status)AS register_status FROM t_das_order_status_info WHERE chain_type=? AND address=? AND order_status=? GROUP BY account
	//err = d.db.Select("account,MAX(register_status) AS register_status").
	//	Where("chain_type=? AND address=? AND action=? AND order_status=?", chainType, address, common.DasActionApplyRegister, tables.OrderStatusDefault).
	//	Group("account").Order("id DESC").Find(&list).Error
	//return
	err = d.db.Select("account,account_id,register_status,cross_coin_type").Where("chain_type=? AND address=? AND action=? AND order_status=?",
		chainType, address, common.DasActionApplyRegister, tables.OrderStatusDefault).Order("id DESC").Find(&list).Error
	return
}

func (d *DbDao) CreateOrderAndOrderTxs(order *tables.TableDasOrderInfo, list []tables.TableDasOrderTxInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"account_id", "account", "action", "chain_type", "address", "register_status",
			}),
		}).Create(&order).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"action", "status", "timestamp",
			}),
		}).Create(&list).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) DoActionPropose(orderIds []string, txs []tables.TableDasOrderTxInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id IN(?) AND register_status=?", orderIds, tables.RegisterStatusProposal).
			Updates(map[string]interface{}{
				"register_status": tables.RegisterStatusConfirmProposal,
			}).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"action", "status", "timestamp",
			}),
		}).Create(&txs).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) DoActionConfirmProposal(orderIds, okOrderIds, accountIds []string, txs []tables.TableDasOrderTxInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id IN(?)", orderIds).
			Updates(map[string]interface{}{
				"register_status": tables.RegisterStatusRegistered,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("account_id IN(?)", accountIds).
			Updates(map[string]interface{}{
				"order_status": tables.OrderStatusClosed,
			}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"action", "status", "timestamp",
			}),
		}).Create(&txs).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id IN(?) AND order_type=? AND hedge_status=?",
				orderIds, tables.OrderTypeSelf, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"hedge_status": tables.TxStatusSending,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderPayInfo{}).
			Where("account_id IN(?) AND order_id NOT IN(?) AND refund_status=?",
				accountIds, okOrderIds, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"refund_status":         tables.TxStatusSending,
				"uni_pay_refund_status": tables.UniPayRefundStatusUnRefund,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) CreateOrder(order *tables.TableDasOrderInfo) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	return d.db.Create(&order).Error
}

func (d *DbDao) CreateOrderWithPayment(order tables.TableDasOrderInfo, payment tables.TableDasOrderPayInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if payment.Hash != "" {
			if err := tx.Create(&payment).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *DbDao) CreateCouponOrder(order *tables.TableDasOrderInfo, coupon string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableCoupon{}).
			Where("code = ? and order_id= ? and use_at=? ", coupon, "", 0).
			Updates(map[string]interface{}{
				"order_id": order.OrderId,
				"use_at":   time.Now().Unix(),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) GetLatestRegisterOrderBySelf(chainType common.ChainType, address, accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("chain_type=? AND address=? AND account_id=? AND action=? AND order_type=? AND order_status=?",
		chainType, address, accountId, common.DasActionApplyRegister, tables.OrderTypeSelf, tables.OrderStatusDefault).
		Order("order_status,register_status DESC,id DESC").Limit(1).
		Find(&order).Error
	return
}

func (d *DbDao) GetOrderByOrderId(orderId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_id=?", orderId).Find(&order).Error
	return
}

func (d *DbDao) GetOrderListByOrderIds(orderIds []string) (list []tables.TableDasOrderInfo, err error) {
	if len(orderIds) == 0 {
		return
	}
	err = d.db.Where("order_id IN(?)", orderIds).Find(&list).Error
	return
}

func (d *DbDao) DoActionApplyRegister(orderId, hash string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderTxInfo{}).
			Where("order_id=? AND `hash`=?", orderId, hash).
			Updates(map[string]interface{}{
				"status": tables.OrderTxStatusConfirm,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND pre_register_status=? AND order_status=?",
				orderId, tables.TxStatusDefault, tables.OrderStatusDefault).
			Updates(map[string]interface{}{
				"pre_register_status": tables.TxStatusSending,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND register_status<?", orderId, tables.RegisterStatusPreRegister).
			Updates(map[string]interface{}{
				"register_status": tables.RegisterStatusPreRegister,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) DoActionPreRegister(orderId, hash string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderTxInfo{}).
			Where("order_id=? AND `hash`=?", orderId, hash).
			Updates(map[string]interface{}{
				"status": tables.OrderTxStatusConfirm,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND register_status<?", orderId, tables.RegisterStatusProposal).
			Updates(map[string]interface{}{
				"register_status": tables.RegisterStatusProposal,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) DoActionRenewAccount(orderId, hash string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := d.db.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_status=?", orderId, tables.OrderStatusDefault).
			Updates(map[string]interface{}{
				"order_status": tables.OrderStatusClosed,
			}).Error; err != nil {
			return err
		}

		if err := d.db.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_type=? AND hedge_status=?", orderId, tables.OrderTypeSelf, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"hedge_status": tables.TxStatusSending,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderTxInfo{}).
			Where("order_id=? AND `hash`=?", orderId, hash).
			Updates(map[string]interface{}{
				"status": tables.OrderTxStatusConfirm,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (d *DbDao) GetNeedSendPayOrderList(action common.DasAction) (list []tables.TableDasOrderInfo, err error) {
	err = d.db.Where("action=? AND order_type=? AND pay_status=? AND order_status=? AND is_did_cell=?",
		action, tables.OrderTypeSelf, tables.TxStatusSending, tables.OrderStatusDefault, tables.IsDidCellNo).
		Order("id").Limit(20).Find(&list).Error
	return
}

func (d *DbDao) GetNeedSendDidCellOrderList(actions []common.DasAction) (list []tables.TableDasOrderInfo, err error) {
	err = d.db.Where("action IN(?) AND order_type=? AND pay_status=? AND order_status=? AND is_did_cell=?",
		actions, tables.OrderTypeSelf, tables.TxStatusSending, tables.OrderStatusDefault, tables.IsDidCellYes).
		Order("id").Limit(20).Find(&list).Error
	return
}

func (d *DbDao) UpdatePayStatus(orderId string, oldTxStatus, newTxStatus tables.TxStatus) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_type=? AND pay_status=?", orderId, tables.OrderTypeSelf, oldTxStatus).
		Updates(map[string]interface{}{
			"pay_status": newTxStatus,
		}).Error
}

func (d *DbDao) GetNeedSendPreRegisterTxOrderList() (list []tables.TableDasOrderInfo, err error) {
	err = d.db.Where("action=? AND order_type=? AND pre_register_status=? AND order_status=?",
		common.DasActionApplyRegister, tables.OrderTypeSelf, tables.TxStatusSending, tables.OrderStatusDefault).
		Order("id").Limit(20).Find(&list).Error
	return
}

func (d *DbDao) UpdatePreRegisterStatus(orderId string, oldTxStatus, newTxStatus tables.TxStatus) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_type=? AND pre_register_status=?",
			orderId, tables.OrderTypeSelf, oldTxStatus).
		Updates(map[string]interface{}{
			"pre_register_status": newTxStatus,
		}).Error
}

func (d *DbDao) UpdateOrderToClosedAndRefund(orderId string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_type=? AND order_status=?",
				orderId, tables.OrderTypeSelf, tables.OrderStatusDefault).
			Updates(map[string]interface{}{
				"order_status": tables.OrderStatusClosed,
			}).Error; err != nil {
			return err
		}
		//if err := tx.Model(tables.TableDasOrderPayInfo{}).
		//	Where("order_id=? AND `status`=? AND refund_status=?",
		//		orderId, tables.OrderTxStatusConfirm, tables.TxStatusDefault).
		//	Updates(map[string]interface{}{
		//		"refund_status": tables.TxStatusSending,
		//	}).Error; err != nil {
		//	return err
		//}
		return nil
	})
}

func (d *DbDao) GetPreRegisteredOrderByAccountId(accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("account_id=? AND action=? AND register_status>?",
		accountId, common.DasActionApplyRegister, tables.RegisterStatusPreRegister).
		Limit(1).Find(&order).Error
	return
}

func (d *DbDao) UpdateOrderToRefund(orderId string) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_type=?", orderId, tables.OrderTypeSelf).
			Updates(map[string]interface{}{
				"order_status": tables.OrderStatusClosed,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

type ExpiredOrder struct {
	OrderId string `json:"order_id" gorm:"column:order_id"`
}

func (d *DbDao) GetNeedExpiredOrders() (list []ExpiredOrder, err error) {
	//SELECT o.order_id,p.order_id FROM(
	//	SELECT order_id FROM t_das_order_info WHERE order_type=1 AND register_status=1 AND order_status=0 AND `timestamp`<0
	//) o LEFT JOIN t_das_order_pay_info p ON p.order_id=o.order_id
	//WHERE p.order_id IS NULL
	sql := fmt.Sprintf("SELECT o.order_id FROM(SELECT order_id FROM %s WHERE order_type=? AND register_status=? AND order_status=? AND `timestamp`<?) o LEFT JOIN %s p ON p.order_id=o.order_id WHERE p.order_id IS NULL", tables.TableNameDasOrderInfo, tables.TableNameDasOrderPayInfo)
	timestamp := time.Now().Add(-time.Hour*24).UnixNano() / 1e6
	err = d.db.Raw(sql, tables.OrderTypeSelf, tables.RegisterStatusConfirmPayment, tables.OrderStatusDefault, timestamp).Find(&list).Error
	return
}

func (d *DbDao) GetNeedExpiredOrders2() (list []ExpiredOrder, err error) {
	sql := fmt.Sprintf("SELECT order_id FROM %s WHERE order_type=? AND register_status=? AND order_status=? AND `timestamp`<?", tables.TableNameDasOrderInfo)
	timestamp := time.Now().Add(-time.Hour*24*7).UnixNano() / 1e6
	err = d.db.Raw(sql, tables.OrderTypeSelf, tables.RegisterStatusConfirmPayment, tables.OrderStatusDefault, timestamp).Find(&list).Error
	return
}

func (d *DbDao) DoExpiredOrder(orderId string) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_type=? AND register_status=? AND order_status=?",
			orderId, tables.OrderTypeSelf, tables.RegisterStatusConfirmPayment, tables.OrderStatusDefault).
		Updates(map[string]interface{}{
			"order_status": tables.OrderStatusClosed,
		}).Error
}

func (d *DbDao) GetClosedAndUnRefundOrders() (list []tables.TableDasOrderInfo, err error) {
	sql := fmt.Sprintf("SELECT o.order_id,o.pay_status,o.pre_register_status FROM %s o JOIN %s p ON o.order_id=p.order_id AND o.action=? AND o.order_type=? AND o.order_status=? AND o.register_status<? AND p.`status`=? AND p.refund_status=?", tables.TableNameDasOrderInfo, tables.TableNameDasOrderPayInfo)
	err = d.db.Raw(sql, common.DasActionApplyRegister, tables.OrderTypeSelf, tables.OrderStatusClosed, tables.RegisterStatusProposal, tables.OrderTxStatusConfirm, tables.TxStatusDefault).Find(&list).Error
	return
}

func (d *DbDao) UpdateOrderRedoApply(orderId string) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_status=? AND pay_status=? AND pre_register_status=?",
			orderId, tables.OrderStatusDefault, tables.TxStatusOk, tables.TxStatusSending).
		Updates(map[string]interface{}{
			"pay_status":          tables.TxStatusSending,
			"pre_register_status": tables.TxStatusDefault,
		}).Error
}

func (d *DbDao) UpdateOrderStatusClosed(orderId string) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_status=?", orderId, tables.OrderStatusDefault).
		Updates(map[string]interface{}{
			"order_status": tables.OrderStatusClosed,
		}).Error
}

func (d *DbDao) GetUnPayOrderCount(chainType common.ChainType, address string) (count int64, err error) {
	err = d.db.Model(tables.TableDasOrderInfo{}).
		Where("chain_type=? AND address=? AND order_status=? AND pay_status=?",
			chainType, address, tables.OrderStatusDefault, tables.TxStatusDefault).Count(&count).Error
	return
}

func (d *DbDao) GetHistoryOkOrders(accountIds []string) (list []tables.TableDasOrderInfo, err error) {
	err = d.db.Where("account_id IN(?) AND order_type=? AND pay_status!=? AND order_status=?",
		accountIds, tables.OrderTypeSelf, tables.TxStatusDefault, tables.OrderStatusClosed).
		Find(&list).Error
	return
}

// unipay
func (d *DbDao) GetIsUniPayInfoByOrderIds(orderIds []string) (list []tables.TableDasOrderInfo, err error) {
	if len(orderIds) == 0 {
		return
	}
	err = d.db.Select("order_id,is_uni_pay").
		Where("order_id IN(?)", orderIds).Find(&list).Error
	return
}

func (d *DbDao) GetNeedHedgeOrderList() (list []tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_type=? AND hedge_status=?", tables.OrderTypeSelf, tables.TxStatusSending).Find(&list).Error
	return
}

func (d *DbDao) UpdateHedgeStatus(orderId string, oldStatus, newStatus tables.TxStatus) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_type=? AND hedge_status=?",
			orderId, tables.OrderTypeSelf, oldStatus).
		Updates(map[string]interface{}{
			"hedge_status": newStatus,
		}).Error
}
