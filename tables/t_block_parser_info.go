package tables

import "time"

type TableBlockParserInfo struct {
	Id          uint64     `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	ParserType  ParserType `json:"parser_type" gorm:"column:parser_type;uniqueIndex:uk_parser_number;type:smallint(6) NOT NULL DEFAULT '0' COMMENT 'das-99 ckb-0 eth-1 btc-2 tron-3 bsc-5 4-wx polygon-6'"`
	BlockNumber uint64     `json:"block_number" gorm:"column:block_number;uniqueIndex:uk_parser_number;type:bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT ''"`
	BlockHash   string     `json:"block_hash" gorm:"column:block_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ParentHash  string     `json:"parent_hash" gorm:"column:parent_hash;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	CreatedAt   time.Time  `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
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
