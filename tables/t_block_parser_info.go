package tables

import "time"

type TableBlockParserInfo struct {
	Id          uint64     `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	ParserType  ParserType `json:"parser_type" gorm:"column:parser_type"`
	BlockNumber uint64     `json:"block_number" gorm:"column:block_number"`
	BlockHash   string     `json:"block_hash" gorm:"column:block_hash"`
	ParentHash  string     `json:"parent_hash" gorm:"column:parent_hash"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameBlockParserInfo = "t_block_parser_info"
)

func (t *TableBlockParserInfo) TableName() string {
	return TableNameBlockParserInfo
}

type ParserType int

const (
	ParserTypeDAS     = 99
	ParserTypeCKB     = 0
	ParserTypeETH     = 1
	ParserTypeTRON    = 3
	ParserTypeBSC     = 5
	ParserTypePOLYGON = 6
)
