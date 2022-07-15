package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
	"time"
)

type TabletDutchAuctionInfo struct {
	Id            uint64           `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	Hash          string           `json:"hash" gorm:"column:hash;uniqueIndex:uk_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ChainType     common.ChainType `json:"chain_type" gorm:"column:chain_type;index:k_address;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Address       string           `json:"address" gorm:"column:address;index:k_address;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Account       string           `json:"account" gorm:"column:account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	AccountId     string           `json:"account_id" gorm:"account_id;index:k_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Price         decimal.Decimal  `json:"price" gorm:"column:price;decimal(60) NOT NULL DEFAULT '0' COMMENT ''"`
	Timestamp     int64            `json:"timestamp" gorm:"column:timestamp;type:bigint(20) NOT NULL DEFAULT '0' COMMENT ''"`
	RefundHash    string           `json:"refund_hash" gorm:"column:refund_hash;index:k_refund_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Status        DutchStatus      `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-default 1-pending 2-confirm 3-success 4-failed 5-refunding 6-refunded'"`
	AuctionStatus int              `json:"auction_status" gorm:"column:auction_status;index:k_auction_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-opened 1-closed'"`
	CreatedAt     time.Time        `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt     time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameDutchAuctionInfo = "t_dutch_auction_info"
)

func (t *TabletDutchAuctionInfo) TableName() string {
	return TableNameDutchAuctionInfo
}

type DutchStatus int

const (
	DutchStatusDefault   DutchStatus = 0
	DutchStatusPending   DutchStatus = 1
	DutchStatusConfirm   DutchStatus = 2
	DutchStatusSuccess   DutchStatus = 3
	DutchStatusFailed    DutchStatus = 4
	DutchStatusRefunding DutchStatus = 5
	DutchStatusRefund    DutchStatus = 6
)

type AuctionStatus int

const (
	AuctionStatusOpened AuctionStatus = 0
	AuctionStatusClosed AuctionStatus = 1
)
