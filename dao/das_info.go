package dao

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
)

type AccountNumRegisterNum struct {
	Num   int `json:"num" gorm:"column:num"`
	Total int `json:"total" gorm:"column:total"`
}

func (d *DbDao) GetAccountNumRegisterNum() (list []AccountNumRegisterNum, err error) {
	sql := fmt.Sprintf(`SELECT a.num,COUNT(*) AS total FROM(
SELECT CHAR_LENGTH(account)-4 AS num FROM %s WHERE account!=''
)a GROUP BY a.num ORDER BY a.num`, tables.TableNameAccountInfo)
	err = d.parserDb.Raw(sql).Find(&list).Error
	return
}

type OrderTotalAmount struct {
	PayTokenId tables.PayTokenId `json:"pay_token_id" gorm:"column:pay_token_id"`
	Amount     decimal.Decimal   `json:"amount" gorm:"column:amount"`
	Num        int               `json:"num" gorm:"column:num"`
}

func (d *DbDao) GetOrderTotalAmount() (list []OrderTotalAmount, err error) {
	err = d.db.Model(tables.TableDasOrderInfo{}).
		Select("pay_token_id,SUM(pay_amount) amount,count(*) num").
		Where("order_type=? AND pay_status=?", tables.OrderTypeSelf, tables.TxStatusOk).
		Group("pay_token_id").Find(&list).Error
	return list, nil
}

func (d *DbDao) GetOrderRefundTotalAmount() (list []OrderTotalAmount, err error) {
	sql := fmt.Sprintf(`SELECT o.pay_token_id,SUM(o.pay_amount)amount,count(*)num FROM %s p 
LEFT JOIN %s o ON o.order_id=p.order_id 
WHERE p.refund_status=2
GROUP BY o.pay_token_id`, tables.TableNameDasOrderPayInfo, tables.TableNameDasOrderInfo)
	err = d.db.Raw(sql).Find(&list).Error
	return list, nil
}

//

func (d *DbDao) GetAccountCount() (count int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).Where("account!=''").Count(&count).Error
	return
}

func (d *DbDao) GetOwnerCount() (count int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).Where("account!=''").Group("owner_chain_type,`owner`").Count(&count).Error
	return
}

type RegisterStatusCount struct {
	RegisterStatus tables.RegisterStatus `json:"register_status" gorm:"column:register_status"`
	CountNum       int64                 `json:"count_num" gorm:"column:count_num"`
}

func (d *DbDao) GetRegisterStatusCount() (list []RegisterStatusCount, err error) {
	err = d.db.Model(tables.TableDasOrderInfo{}).Select("register_status,count(*) AS count_num").
		Where("order_type=? AND action=? AND order_status=? AND register_status>? AND register_status<?",
			tables.OrderTypeSelf, common.DasActionApplyRegister, tables.OrderStatusDefault,
			tables.RegisterStatusConfirmPayment, tables.RegisterStatusRegistered).
		Group("register_status").Find(&list).Error
	return
}
