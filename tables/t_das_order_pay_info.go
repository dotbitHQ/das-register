package tables

import "github.com/DeAccountSystems/das-lib/common"

type TableDasOrderPayInfo struct {
	Id           uint64           `json:"id" gorm:"column:id"`
	Hash         string           `json:"hash" gorm:"column:hash"`
	OrderId      string           `json:"order_id" gorm:"column:order_id"`
	ChainType    common.ChainType `json:"chain_type" gorm:"column:chain_type"`
	Address      string           `json:"address" gorm:"column:address"`
	Status       OrderTxStatus    `json:"status" gorm:"column:status"`
	Timestamp    int64            `json:"timestamp" gorm:"column:timestamp"`
	AccountId    string           `json:"account_id" gorm:"account_id"`
	RefundStatus TxStatus         `json:"refund_status" gorm:"column:refund_status"`
	RefundHash   string           `json:"refund_hash" gorm:"column:refund_hash"`
	//CreatedAt    time.Time        `json:"created_at" gorm:"column:created_at"`
	//UpdatedAt    time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameDasOrderPayInfo = "t_das_order_pay_info"
)

func (t *TableDasOrderPayInfo) TableName() string {
	return TableNameDasOrderPayInfo
}

type OrderTxStatus int

const (
	OrderTxStatusDefault  OrderTxStatus = 0
	OrderTxStatusConfirm  OrderTxStatus = 1
	OrderTxStatusRejected OrderTxStatus = 2
)
