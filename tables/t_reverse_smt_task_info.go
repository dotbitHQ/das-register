package tables

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	ReverseSmtStatusDefault         = 0
	ReverseSmtStatusPending         = 1
	ReverseSmtStatusConfirm         = 2
	ReverseSmtStatusRollback        = 3
	ReverseSmtStatusRollbackConfirm = 4

	ReverseSmtTxStatusDefault = 0
	ReverseSmtTxStatusPending = 1
	ReverseSmtTxStatusConfirm = 2
	ReverseSmtTxStatusReject  = 3

	ReverseSmtMaxRetryNum = 3
)

// ReverseSmtTaskInfo reverse task info
type ReverseSmtTaskInfo struct {
	ID          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	TaskID      string    `gorm:"column:task_id;NOT NULL;index:idx_task_id,unique"` // 批处理任务ID
	RefOutpoint string    `gorm:"column:ref_outpoint;NOT NULL"`                     // ref sub_account cell outpoint
	BlockNumber uint64    `gorm:"column:block_number;default:0;NOT NULL"`           // tx block number
	Outpoint    string    `gorm:"column:outpoint;NOT NULL;index:idx_outpoint"`      // new sub_account cell outpoint
	Timestamp   int64     `gorm:"column:timestamp;default:0;NOT NULL"`              // record timestamp
	SmtStatus   int       `gorm:"column:smt_status;default:0;NOT NULL"`             // smt的状态
	TxStatus    int       `gorm:"column:tx_status;default:0;NOT NULL"`              // 交易状态
	Retry       int       `gorm:"column:retry;default:0;NOT NULL"`                  // 失败重试次数
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const TableNameReverseSmtTaskInfo = "t_reverse_smt_task_info"

func (m *ReverseSmtTaskInfo) TableName() string {
	return TableNameReverseSmtTaskInfo
}

func (m *ReverseSmtTaskInfo) InitTaskId() error {
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	m.TaskID = strings.ReplaceAll(uid.String(), "-", "")
	return nil
}
