package tables

import "github.com/dotbitHQ/das-lib/common"

type TableReverseInfo struct {
	Id             uint64           `json:"id" gorm:"column:id"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint"`
	AlgorithmId    int              `json:"algorithm_id" gorm:"column:algorithm_id"`
	ChainType      common.ChainType `json:"chain_type" gorm:"column:chain_type"`
	Address        string           `json:"address" gorm:"column:address"`
	AccountId      string           `json:"account_id" gorm:"account_id"`
	Account        string           `json:"account" gorm:"column:account"`
	Capacity       uint64           `json:"capacity" gorm:"column:capacity"`
}

const (
	TableNameReverseInfo = "t_reverse_info"
)

func (t *TableReverseInfo) TableName() string {
	return TableNameReverseInfo
}
