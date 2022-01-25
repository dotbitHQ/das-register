package dao

import (
	"das_register_server/tables"
	"gorm.io/gorm/clause"
)

func (d *DbDao) FindBlockInfo(parserType tables.ParserType) (block tables.TableBlockParserInfo, err error) {
	err = d.db.Where("parser_type=?", parserType).
		Order("block_number DESC").Limit(1).Find(&block).Error
	return
}

func (d *DbDao) FindBlockInfoByBlockNumber(parserType tables.ParserType, blockNumber uint64) (block tables.TableBlockParserInfo, err error) {
	err = d.db.Where("parser_type=? AND block_number=?", parserType, blockNumber).
		Order("block_number DESC").Limit(1).Find(&block).Error
	return
}

func (d *DbDao) CreateBlockInfo(parserType tables.ParserType, blockNumber uint64, blockHash, parentHash string) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"block_hash", "parent_hash"}),
	}).Create(&tables.TableBlockParserInfo{
		ParserType:  parserType,
		BlockNumber: blockNumber,
		BlockHash:   blockHash,
		ParentHash:  parentHash,
	}).Error
}

func (d *DbDao) DeleteBlockInfo(parserType tables.ParserType, blockNumber uint64) error {
	return d.db.Where("parser_type=? AND block_number < ?", parserType, blockNumber).
		Delete(&tables.TableBlockParserInfo{}).Error
}
