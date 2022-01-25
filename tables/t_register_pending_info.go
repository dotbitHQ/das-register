package tables

import (
	"github.com/DeAccountSystems/das-lib/common"
	"time"
)

type TableRegisterPendingInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number"`
	Account        string           `json:"account" gorm:"column:account"`
	Action         string           `json:"action" gorm:"column:action"`
	ChainType      common.ChainType `json:"chain_type" gorm:"column:chain_type"`
	Address        string           `json:"address" gorm:"column:address"`
	Capacity       uint64           `json:"capacity" gorm:"column:capacity"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp"`
	Status         int              `json:"status" gorm:"column:status"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameRegisterPendingInfo = "t_register_pending_info"
	StatusRejected               = -1
	StatusConfirm                = 1
	StatusPending                = 0
)

func (t *TableRegisterPendingInfo) TableName() string {
	return TableNameRegisterPendingInfo
}
