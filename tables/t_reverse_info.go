package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableReverseInfo struct {
	Id             uint64                `json:"id" gorm:"column:id"`
	BlockNumber    uint64                `json:"block_number" gorm:"column:block_number"`
	BlockTimestamp uint64                `json:"block_timestamp" gorm:"column:block_timestamp"`
	Outpoint       string                `json:"outpoint" gorm:"column:outpoint"`
	AlgorithmId    common.DasAlgorithmId `json:"algorithm_id" gorm:"column:algorithm_id"`
	ChainType      common.ChainType      `json:"chain_type" gorm:"column:chain_type"`
	Address        string                `json:"address" gorm:"column:address"`
	AccountId      string                `json:"account_id" gorm:"account_id"`
	Account        string                `json:"account" gorm:"column:account"`
	Capacity       uint64                `json:"capacity" gorm:"column:capacity"`
	ReverseType    uint32                `json:"reverse_type" gorm:"column:reverse_type;type:tinyint(1) NOT NULL DEFAULT '0' COMMENT '0: old reverse type，1：new outpoint struct'"`
	CreatedAt      time.Time             `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time             `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameReverseInfo = "t_reverse_info"

	ReverseTypeOld = 0
	ReverseTypeSmt = 1
)

func (t *TableReverseInfo) TableName() string {
	return TableNameReverseInfo
}
