package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableRegisterPendingInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number;index:k_block_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Account        string           `json:"account" gorm:"column:account;index:k_a_a;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Action         string           `json:"action" gorm:"column:action;uniqueIndex:uk_a_o;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ChainType      common.ChainType `json:"chain_type" gorm:"column:chain_type;index:k_ct_a;type:smallint(6) NOT NULL DEFAULT '0' COMMENT ''"`
	Address        string           `json:"address" gorm:"column:address;index:k_ct_a;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Capacity       uint64           `json:"capacity" gorm:"column:capacity;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint;uniqueIndex:uk_a_o;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	Status         int              `json:"status" gorm:"column:status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-default 1-rejected'"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
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
