package example

import (
	"das_register_server/dao"
	"das_register_server/timer"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestCheckNameDaoMember(t *testing.T) {
	addr, user, password, dbName := "", "", "", "das_database"
	parserDb, err := toolib.NewGormDB(addr, user, password, dbName, 100, 100)
	if err != nil {
		t.Fatal(err)
	}

	var dbDao dao.DbDao
	dbDao.InitDb(nil, parserDb)

	nameDaoTimer := timer.NameDaoTimer{DbDao: &dbDao}
	if err := nameDaoTimer.CheckNameDaoMember("", 2); err != nil {
		t.Fatal(err)
	}
}
