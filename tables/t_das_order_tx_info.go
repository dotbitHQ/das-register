package tables

import "time"

type TableDasOrderTxInfo struct {
	Id        uint64        `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	OrderId   string        `json:"order_id" gorm:"column:order_id;uniqueIndex:uk_order_id_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Hash      string        `json:"hash" gorm:"column:hash;index:k_hash;uniqueIndex:uk_order_id_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Action    OrderTxAction `json:"action" gorm:"column:action;index:k_action;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Status    OrderTxStatus `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-default 1-confirm'"`
	Timestamp int64         `json:"timestamp" gorm:"column:timestamp;type:bigint(20) NOT NULL DEFAULT '0' COMMENT ''"`
	CreatedAt time.Time     `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
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
