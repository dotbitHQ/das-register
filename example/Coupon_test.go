package example

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/timer"
	"sync"

	"fmt"
	"testing"
)

func TestResetCoupon(t *testing.T) {

	dbConfig := config.DbMysql{
		Addr:        "127.0.0.1",
		User:        "root",
		Password:    "123456",
		DbName:      "das_register",
		MaxIdleConn: 20,
		MaxOpenConn: 20,
	}
	dbDao, err := dao.NewGormDB(dbConfig, dbConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	wgServer := sync.WaitGroup{}
	txTimer := timer.NewTxTimer(timer.TxTimerParam{
		Ctx:           context.Background(),
		Wg:            &wgServer,
		DbDao:         dbDao,
		DasCore:       nil,
		DasCache:      nil,
		TxBuilderBase: nil,
	})
	if err := txTimer.DoResetCoupon(); err != nil {
		fmt.Println(err)
	}
}
