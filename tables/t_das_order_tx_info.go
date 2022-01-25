package tables

type TableDasOrderTxInfo struct {
	Id        uint64        `json:"id" gorm:"column:id"`
	OrderId   string        `json:"order_id" gorm:"column:order_id"`
	Hash      string        `json:"hash" gorm:"column:hash"`
	Action    OrderTxAction `json:"action" gorm:"column:action"`
	Status    OrderTxStatus `json:"status" gorm:"column:status"`
	Timestamp int64         `json:"timestamp" gorm:"column:timestamp"`
	//CreatedAt    time.Time        `json:"created_at" gorm:"column:created_at"`
	//UpdatedAt    time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameDasOrderTxInfo = "t_das_order_tx_info"
)

func (t *TableDasOrderTxInfo) TableName() string {
	return TableNameDasOrderTxInfo
}

type OrderTxAction string

const (
	TxActionConfirmPayment  OrderTxAction = "confirm_payment"
	TxActionApplyRegister   OrderTxAction = "apply_register"
	TxActionPreRegister     OrderTxAction = "pre_register"
	TxActionPropose         OrderTxAction = "propose"
	TxActionConfirmProposal OrderTxAction = "confirm_proposal"
	TxActionRenewAccount    OrderTxAction = "renew_account"
)
