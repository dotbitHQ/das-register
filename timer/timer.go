package timer

import (
	"context"
	"das_register_server/dao"
	"fmt"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

var log = mylog.NewLogger("timer", mylog.LevelDebug)

type TxTimer struct {
	ctx     context.Context
	wg      *sync.WaitGroup
	dbDao   *dao.DbDao
	dasCore *core.DasCore
}

type TxTimerParam struct {
	DbDao   *dao.DbDao
	Ctx     context.Context
	Wg      *sync.WaitGroup
	DasCore *core.DasCore
}

func NewTxTimer(p TxTimerParam) *TxTimer {
	var t TxTimer
	t.ctx = p.Ctx
	t.wg = p.Wg
	t.dbDao = p.DbDao
	t.dasCore = p.DasCore
	return &t
}

func (t *TxTimer) Run() error {
	if err := t.doUpdateTokenMap(); err != nil {
		return fmt.Errorf("doUpdateTokenMap init token err: %s", err.Error())
	}
	tickerToken := time.NewTicker(time.Second * 50)
	//tickerPending := time.NewTicker(time.Minute)
	//pendingLimit := 10
	tickerRejected := time.NewTicker(time.Minute * 3)
	t.wg.Add(1)

	go func() {
		for {
			select {
			case <-tickerToken.C:
				log.Info("doUpdateTokenMap start ...")
				if err := t.doUpdateTokenMap(); err != nil {
					log.Error("doUpdateTokenMap err:", err)
				}
				log.Info("doUpdateTokenMap end ...")
			//case <-tickerPending.C:
			//	log.Info("checkPending start ...", pendingLimit)
			//	if limit, err := t.checkPending(pendingLimit); err != nil {
			//		log.Error("checkPending err: ", err.Error())
			//	} else if limit > pendingLimit {
			//		pendingLimit = limit
			//	}
			//	log.Info("checkPending end ...", pendingLimit)
			case <-tickerRejected.C:
				log.Info("checkRejected start ...")
				if err := t.checkRejected(); err != nil {
					log.Error("checkRejected err: ", err.Error())
				}
				log.Info("checkRejected end ...")
			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()
	return nil
}
