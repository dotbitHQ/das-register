package dao

import (
	"das_register_server/tables"
	"fmt"
)

func (d *DbDao) GetPayInfoByOrderId(orderId string) (pay tables.TableDasOrderPayInfo, err error) {
	err = d.db.Where("order_id=?", orderId).Order("id DESC").Limit(1).Find(&pay).Error
	return
}

func (d *DbDao) CreateOrderPayInfo(orderPay *tables.TableDasOrderPayInfo) error {
	if orderPay == nil {
		return fmt.Errorf("order pay info is nil")
	}
	return d.db.Create(&orderPay).Error
}
