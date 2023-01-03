package timer

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/robfig/cron/v3"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

var log = mylog.NewLogger("timer", mylog.LevelDebug)

type TxTimer struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	dbDao         *dao.DbDao
	dasCore       *core.DasCore
	dasCache      *dascache.DasCache
	txBuilderBase *txbuilder.DasTxBuilderBase
	cron          *cron.Cron
}

type TxTimerParam struct {
	DbDao         *dao.DbDao
	Ctx           context.Context
	Wg            *sync.WaitGroup
	DasCore       *core.DasCore
	DasCache      *dascache.DasCache
	TxBuilderBase *txbuilder.DasTxBuilderBase
}

func NewTxTimer(p TxTimerParam) *TxTimer {
	var t TxTimer
	t.ctx = p.Ctx
	t.wg = p.Wg
	t.dbDao = p.DbDao
	t.dasCore = p.DasCore
	t.dasCache = p.DasCache
	t.txBuilderBase = p.TxBuilderBase
	return &t
}

func (t *TxTimer) Run() error {
	if err := t.doUpdateTokenMap(); err != nil {
		return fmt.Errorf("doUpdateTokenMap init token err: %s", err.Error())
	}
	tickerToken := time.NewTicker(time.Second * 50)
	tickerRejected := time.NewTicker(time.Minute * 3)
	tickerTxRejected := time.NewTicker(time.Minute * 5)

	tickerExpired := time.NewTicker(time.Minute * 30)
	tickerRecover := time.NewTicker(time.Minute * 3)
	if config.Cfg.Server.RecoverTime > 0 {
		tickerRecover = time.NewTicker(time.Minute * config.Cfg.Server.RecoverTime)
	}
	tickerRefundApply := time.NewTicker(time.Minute * 10)
	tickerClosedAndUnRefund := time.NewTicker(time.Minute * 20)

	t.wg.Add(5)
	go func() {
		for {
			select {
			case <-tickerToken.C:
				log.Info("doUpdateTokenMap start ...")
				if err := t.doUpdateTokenMap(); err != nil {
					log.Error("doUpdateTokenMap err:", err)
				}
				log.Info("doUpdateTokenMap end ...")
			case <-tickerRejected.C:
				log.Info("checkRejected start ...")
				if err := t.checkRejected(); err != nil {
					log.Error("checkRejected err: ", err.Error())
				}
				log.Info("checkRejected end ...")
			case <-tickerTxRejected.C:
				log.Info("doTxRejected start ...")
				if err := t.doTxRejected(); err != nil {
					log.Error("doTxRejected err: ", err.Error())
				}
				log.Info("doTxRejected end ...")
			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerExpired.C:
				log.Info("checkExpired start ...")
				if err := t.checkExpired(); err != nil {
					log.Error("checkExpired err: ", err.Error())
				}
				log.Info("checkExpired end ...")
			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerRefundApply.C:
				log.Info("doRefundApply start ...")
				//if err := t.doRefundApply(); err != nil {
				//	log.Error("doRefundApply err: ", err.Error())
				//}
				if err := t.doRecycleApply(); err != nil {
					log.Errorf("doRecycleApply err: %s", err.Error())
				}
				if config.Cfg.Server.RecycleAllPre {
					//if err := t.doRefundPre(); err != nil {
					//	log.Error("doRefundPre err: %s", err.Error())
					//}
					if err := t.doRecyclePre(); err != nil {
						log.Error("doRecyclePre err: ", err.Error())
					}
				}
				log.Info("doRefundApply end ...")
			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerRecover.C:
				log.Info("doRecoverCkb start ...")
				if err := t.doRecoverCkb(); err != nil {
					log.Error("doRecoverCkb err: ", err.Error())
				}
				log.Info("doRecoverCkb end ...")
			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tickerClosedAndUnRefund.C:
				log.Info("doCheckClosedAndUnRefund start ...")
				if err := t.doCheckClosedAndUnRefund(); err != nil {
					log.Error("doCheckClosedAndUnRefund err: ", err.Error())
				}
				log.Info("doCheckClosedAndUnRefund start ...")
			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	return nil
}
