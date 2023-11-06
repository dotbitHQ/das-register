package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
	"time"
)

const (
	TableNameAuctionOrder = "t_auction_order"
)

type TableAuctionOrder struct {
	Id             uint64                   `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	Account        string                   `json:"account" gorm:"column:account;type:varchar(255) DEFAULT NULL"`
	AccountId      string                   `json:"account_id" gorm:"column:account_id;type:varchar(255) DEFAULT NULL"`
	OrderId        string                   `json:"order_id" gorm:"column:order_id;type:varchar(255) DEFAULT NULL"`
	Address        string                   `json:"address" gorm:"column:address;type:varchar(255) DEFAULT NULL"`
	ChainType      common.ChainType         `json:"chain_type" gorm:"column:chain_type;index:k_chain_type_address;type:smallint(6) NOT NULL DEFAULT '0' COMMENT 'order chain type'"`
	AlgorithmId    common.DasAlgorithmId    `json:"algorithm_id" gorm:"column:algorithm_id"`
	SubAlgorithmId common.DasSubAlgorithmId `json:"sub_algorithm_id" gorm:"column:sub_algorithm_id"`
	BasicPrice     decimal.Decimal          `json:"basic_price" gorm:"column:basic_price;type:decimal(60,0) NOT NULL DEFAULT '0' COMMENT ''"`
	PremiumPrice   decimal.Decimal          `json:"premium_price" gorm:"column:premium_price;type:decimal(60,0) NOT NULL DEFAULT '0' COMMENT ''"`
	Status         int                      `json:"status"`
	BidTime        int64                    `json:"bid_time" gorm:"column:bid_time;type:bigint NOT NULL DEFAULT '0'"`
	Outpoint       string                   `json:"outpoint" gorm:"column:outpoint;type:varchar(255) DEFAULT NULL"`
	CreatedAt      time.Time                `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time                `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

func (t *TableAuctionOrder) TableName() string {
	return TableNameAuctionOrder
}

type BidStatus int

const (
	BidStatusNoOne    BidStatus = 0 //账号没有人竞
	BidStatusByOthers BidStatus = 1 //账号被其他人竞拍
	BidStatusByMe     BidStatus = 2 //账号被我竞拍
)

func (t *TableAuctionOrder) CreateOrderId() {
	t.OrderId = CreateOrderId(1, t.AccountId, common.DasBidExpiredAccountAuction, t.ChainType, t.Address, t.BidTime)
}
