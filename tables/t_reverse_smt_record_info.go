package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/smt"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"time"
)

const (
	SubActionUpdate = "update"
	SubActionRemove = "remove"
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
	SubAction   string    `gorm:"column:sub_action;NOT NULL"`          // 交易的子类型：update, remove
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt   time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"`
}

const TableNameReverseSmtRecordInfo = "t_reverse_smt_record_info"

func (m *ReverseSmtRecordInfo) TableName() string {
	return "t_reverse_smt_record_info"
}

func (m *ReverseSmtRecordInfo) GetSmtKey() (smt.H256, error) {
	smtKeyBlake256, err := blake2b.Blake256(common.Hex2Bytes(m.Address))
	if err != nil {
		return nil, err
	}
	smtKey := smt.ToSmtH256(common.Bytes2Hex(smtKeyBlake256))
	return smtKey, nil
}

func (m *ReverseSmtRecordInfo) GetSmtValue() (smt.H256, error) {
	valBs := make([]byte, 0)
	nonce := molecule.GoU32ToMoleculeU32(uint32(m.Nonce))
	valBs = append(valBs, nonce.RawData()...)
	valBs = append(valBs, []byte(m.Account)...)

	smtValBlake256, err := blake2b.Blake256(valBs)
	if err != nil {
		return nil, err
	}
	smtVal := smt.ToSmtH256(common.Bytes2Hex(smtValBlake256))
	return smtVal, nil
}
