package dao

import (
	"das_register_server/tables"
	"fmt"
)

func (d *DbDao) CreateCoupon(data []tables.TableCoupon) (err error) {
	if len(data) == 0 {
		return fmt.Errorf("coupon is empty")
	}
	err = d.db.Create(&data).Error
	return
}

func (d *DbDao) GetCouponByCode(code string) (coupon tables.TableCoupon, err error) {
	err = d.db.Where("code=?", code).Limit(1).Find(&coupon).Error
	return
}

func (d *DbDao) GetUsedCouponNotChecked() (list []tables.TableDasOrderInfo, err error) {
	sql := fmt.Sprintf("select o.register_status,o.order_status,o.order_id from %s as c join %s as o on c.order_id = o.order_id where length(c.order_id) > 0 and c.is_check=0 ; ", tables.TableNameCoupon, tables.TableNameDasOrderInfo)
	err = d.db.Raw(sql).Find(&list).Error
	return
}

//reset coupon when the order is
func (d *DbDao) ResetCoupon(order_id []string) error {
	return d.db.Model(tables.TableCoupon{}).
		Where(" order_id in ? AND is_check=0 ", order_id).
		Updates(map[string]interface{}{
			"order_id": "",
			"use_at":   0,
		}).Error
}

func (d *DbDao) MarkChecked(order_id []string) error {
	return d.db.Model(tables.TableCoupon{}).
		Where(" order_id in ? AND is_check=0 ", order_id).
		Updates(map[string]interface{}{
			"is_check": 1,
		}).Error
}
