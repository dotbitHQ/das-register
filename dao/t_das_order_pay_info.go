package dao

import (
	"das_register_server/tables"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetPayInfoByOrderId(orderId string) (pay tables.TableDasOrderPayInfo, err error) {
	err = d.db.Where("order_id=? and `status`!=?", orderId, tables.OrderTxStatusRejected).Order("id DESC").Limit(1).Find(&pay).Error
	return
}

func (d *DbDao) CreateOrderPayInfo(orderPay *tables.TableDasOrderPayInfo) error {
	if orderPay == nil {
		return fmt.Errorf("order pay info is nil")
	}
	return d.db.Create(&orderPay).Error
}

func (d *DbDao) UpdatePayToRefund(orderId string) error {
	return d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("order_id=? AND `status`=? AND refund_status=?",
			orderId, tables.OrderTxStatusConfirm, tables.TxStatusDefault).
		Updates(map[string]interface{}{
			"refund_status": tables.TxStatusSending,
		}).Error
}

func (d *DbDao) GetUnRefundTxCount() (count int64, err error) {
	err = d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("`status`=? AND refund_status=?", tables.OrderTxStatusConfirm, tables.TxStatusSending).Count(&count).Error
	return
}

// unipay
func (d *DbDao) UpdateUniPayRefundStatusToRefunded(payHash, orderId, refundHash string) error {
	return d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("hash=? AND order_id=? AND `status`=? AND uni_pay_refund_status=?",
			payHash, orderId, tables.OrderTxStatusConfirm, tables.UniPayRefundStatusRefunding).
		Updates(map[string]interface{}{
			"uni_pay_refund_status": tables.UniPayRefundStatusRefunded,
			"refund_hash":           refundHash,
		}).Error
}

func (d *DbDao) UpdatePayment(paymentInfo tables.TableDasOrderPayInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_type=? AND pay_status=?",
				paymentInfo.OrderId, tables.OrderTypeSelf, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"pay_status":      tables.TxStatusSending,
				"register_status": tables.RegisterStatusApplyRegister,
			}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&paymentInfo).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderPayInfo{}).
			Where("`hash`=? AND order_id=? AND `status`=?",
				paymentInfo.Hash, paymentInfo.OrderId, tables.OrderTxStatusDefault).
			Updates(map[string]interface{}{
				"chain_type": paymentInfo.ChainType,
				"address":    paymentInfo.Address,
				"status":     paymentInfo.Status,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) GetPayHashStatusPendingList() (list []tables.TableDasOrderPayInfo, err error) {
	timestamp := tables.GetPaymentInfoTimestamp()
	err = d.db.Where("timestamp>=? AND `status`=?",
		timestamp, tables.OrderTxStatusDefault).Find(&list).Error
	return
}

func (d *DbDao) GetRefundStatusRefundingList() (list []tables.TableDasOrderPayInfo, err error) {
	timestamp := tables.GetPaymentInfoTimestamp()
	err = d.db.Where("timestamp>=? AND `status`=? AND uni_pay_refund_status=?",
		timestamp, tables.OrderTxStatusConfirm, tables.UniPayRefundStatusRefunding).Find(&list).Error
	return
}

func (d *DbDao) UpdateUniPayUnconfirmedToRejected(orderId, payHash string) error {
	return d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("`hash`=? AND order_id=? AND `status`=? AND uni_pay_refund_status=?",
			payHash, orderId, tables.OrderTxStatusDefault, tables.UniPayRefundStatusDefault).
		Updates(map[string]interface{}{
			"status": tables.OrderTxStatusRejected,
		}).Error
}

func (d *DbDao) GetUnRefundList() (list []tables.TableDasOrderPayInfo, err error) {
	timestamp := tables.GetPaymentInfoTimestamp()
	err = d.db.Where("timestamp>=? AND `status`=? AND uni_pay_refund_status=?",
		timestamp, tables.OrderTxStatusConfirm, tables.UniPayRefundStatusUnRefund).Find(&list).Error
	return
}

func (d *DbDao) UpdateRefundStatusToRefundIng(ids []uint64) error {
	return d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("id IN(?) AND `status`=? AND uni_pay_refund_status=?",
			ids, tables.OrderTxStatusConfirm, tables.UniPayRefundStatusUnRefund).
		Updates(map[string]interface{}{
			"uni_pay_refund_status": tables.UniPayRefundStatusRefunded,
		}).Error
}
