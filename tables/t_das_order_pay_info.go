package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableDasOrderPayInfo struct {
	Id                 uint64             `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	Hash               string             `json:"hash" gorm:"column:hash;uniqueIndex:uk_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	OrderId            string             `json:"order_id" gorm:"column:order_id;index:k_order_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ChainType          common.ChainType   `json:"chain_type" gorm:"column:chain_type;index:k_address;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Address            string             `json:"address" gorm:"column:address;index:k_address;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Status             OrderTxStatus      `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-default 1-confirm'"`
	Timestamp          int64              `json:"timestamp" gorm:"column:timestamp; index:idx_timestamp; type:bigint(20) NOT NULL DEFAULT '0' COMMENT ''"`
	AccountId          string             `json:"account_id" gorm:"account_id;index:k_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	RefundStatus       TxStatus           `json:"refund_status" gorm:"column:refund_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok'"`
	UniPayRefundStatus UniPayRefundStatus `json:"uni_pay_refund_status" gorm:"column:uni_pay_refund_status; type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok';"`
	RefundHash         string             `json:"refund_hash" gorm:"column:refund_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	CreatedAt          time.Time          `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt          time.Time          `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameDasOrderPayInfo = "t_das_order_pay_info"
)

func (t *TableDasOrderPayInfo) TableName() string {
	return TableNameDasOrderPayInfo
}

func GetPaymentInfoTimestamp() int64 {
	return time.Now().Add(-time.Hour * 24 * 3).UnixMilli()
}

func GetPaymentInfoTimestampBefore24h() int64 {
	return time.Now().Add(-time.Hour * 24).UnixMilli()
}

func (t *TableDasOrderPayInfo) PayHashUnconfirmedMin() float64 {
	return time.Now().Sub(time.UnixMilli(t.Timestamp)).Minutes()
}

type OrderTxStatus int

const (
	OrderTxStatusDefault  OrderTxStatus = 0
	OrderTxStatusConfirm  OrderTxStatus = 1
	OrderTxStatusRejected OrderTxStatus = 2
	OrderTxStatusDispute  OrderTxStatus = 3
)

// unipay

type UniPayRefundStatus int

const (
	UniPayRefundStatusDefault   UniPayRefundStatus = 0
	UniPayRefundStatusUnRefund  UniPayRefundStatus = 1
	UniPayRefundStatusRefunding UniPayRefundStatus = 2
	UniPayRefundStatusRefunded  UniPayRefundStatus = 3
)

type PayHashStatus int

const (
	PayHashStatusPending   PayHashStatus = 0
	PayHashStatusConfirmed PayHashStatus = 1
	PayHashStatusRejected  PayHashStatus = 2
)

type UniPayStatus int

const (
	UniPayStatusUnpaid UniPayStatus = 0
	UniPayStatusPaid   UniPayStatus = 1
)
