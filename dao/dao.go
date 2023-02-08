package dao

import (
	"das_register_server/config"
	"das_register_server/tables"
	"fmt"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
)

type DbDao struct {
	db       *gorm.DB
	parserDb *gorm.DB
}

func (d *DbDao) InitDb(db, parserDb *gorm.DB) {
	d.parserDb = parserDb
	d.db = db
}

func NewGormDB(dbMysql, parserMysql config.DbMysql) (*DbDao, error) {
	db, err := toolib.NewGormDB(dbMysql.Addr, dbMysql.User, dbMysql.Password, dbMysql.DbName, dbMysql.MaxOpenConn, dbMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("toolib.NewGormDB err: %s", err.Error())
	}

	// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes.
	// It will change existing column’s type if its size, precision, nullable changed.
	// It WON’T delete unused columns to protect your data.
	if err = db.AutoMigrate(
		&tables.TableBlockParserInfo{},
		&tables.TableDasOrderInfo{},
		&tables.TableDasOrderPayInfo{},
		&tables.TableDasOrderTxInfo{},
		&tables.TableRegisterPendingInfo{},
	); err != nil {
		return nil, err
	}

	parserDb, err := toolib.NewGormDB(parserMysql.Addr, parserMysql.User, parserMysql.Password, parserMysql.DbName, parserMysql.MaxOpenConn, parserMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("toolib.NewGormDB err: %s", err.Error())
	}
	return &DbDao{db: db, parserDb: parserDb}, nil
}

func (d *DbDao) Transaction(fc func(tx *gorm.DB) error) error {
	return d.db.Transaction(fc)
}

type RecordTotal struct {
	Total int `json:"total" gorm:"column:total"`
}
