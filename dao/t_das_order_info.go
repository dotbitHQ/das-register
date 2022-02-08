package dao

import (
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetLatestRegisterOrderByAddress(chainType common.ChainType, address, accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("chain_type=? AND address=? AND account_id=? AND action=?",
		chainType, address, accountId, common.DasActionApplyRegister).
		Order("register_status DESC,id DESC").Limit(1).Find(&order).Error
	return
}

func (d *DbDao) GetLatestRegisterOrderByLatest(accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("account_id=? AND action=? AND order_status=?",
		accountId, common.DasActionApplyRegister, tables.OrderStatusDefault).
		Order("register_status DESC,id DESC").Limit(1).Find(&order).Error
	return
}

func (d *DbDao) GetRegisteringOrders(chainType common.ChainType, address string) (list []tables.TableDasOrderInfo, err error) {
	// SELECT account,MAX(register_status)AS register_status FROM t_das_order_status_info WHERE chain_type=? AND address=? AND order_status=? GROUP BY account
	err = d.db.Select("account,MAX(register_status) AS register_status").
		Where("chain_type=? AND address=? AND action=? AND order_status=?", chainType, address, common.DasActionApplyRegister, tables.OrderStatusDefault).
		Group("account").Order("id DESC").Find(&list).Error
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

func (d *DbDao) DoActionConfirmProposal(orderIds, accountIds []string, txs []tables.TableDasOrderTxInfo) error {
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
				accountIds, orderIds, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"refund_status": tables.TxStatusSending,
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

func (d *DbDao) GetLatestRegisterOrderBySelf(chainType common.ChainType, address, accountId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("chain_type=? AND address=? AND account_id=? AND action=? AND order_type=?",
		chainType, address, accountId, common.DasActionApplyRegister, tables.OrderTypeSelf).
		Order("register_status DESC,id DESC").Limit(1).
		Find(&order).Error
	return
}

func (d *DbDao) GetOrderByOrderId(orderId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_id=?", orderId).Find(&order).Error
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
	err = d.db.Where("action=? AND order_type=? AND pay_status=? AND order_status=?",
		action, tables.OrderTypeSelf, tables.TxStatusSending, tables.OrderStatusDefault).
		Order("id").Limit(10).Find(&list).Error
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
		Order("id").Limit(10).Find(&list).Error
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
		if err := tx.Model(tables.TableDasOrderPayInfo{}).
			Where("order_id=? AND refund_status=?", orderId, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"refund_status": tables.TxStatusSending,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}