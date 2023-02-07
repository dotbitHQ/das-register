package tables

import (
	"time"
)

// ReverseSmtRecordInfo reverse smt record info
type ReverseSmtRecordInfo struct {
	ID          uint64    `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Address     string    `gorm:"column:address;NOT NULL"` // 设置反向解析的地址
	AlgorithmID int       `gorm:"column:algorithm_id;default:0;NOT NULL"`
	Nonce       int       `gorm:"column:nonce;default:0;NOT NULL"`
	TaskID      string    `gorm:"column:task_id;NOT NULL"`             // 批处理任务ID
	Account     string    `gorm:"column:account;default:0;NOT NULL"`   // 子账户名
	Signature   string    `gorm:"column:signature;NOT NULL"`           // 用户操作的签名
	Timestamp   int64     `gorm:"column:timestamp;default:0;NOT NULL"` // record timestamp
	SubAction   string    `gorm:"column:sub_action;NOT NULL"`          // 交易的子类型：create, update, delete
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (m *ReverseSmtRecordInfo) TableName() string {
	return "t_reverse_smt_record_info"
}
