package timer

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

var log = logger.NewLogger("timer", logger.LevelDebug)

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
	return nil
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
	tickerResetCoupon := time.NewTicker(time.Minute * 1)
	t.wg.Add(5)
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerToken.C:
				log.Debug("doUpdateTokenMap start ...")
				if err := t.doUpdateTokenMap(); err != nil {
					log.Error("doUpdateTokenMap err:", err)
				}
				log.Debug("doUpdateTokenMap end ...")
			case <-tickerRejected.C:
				log.Debug("checkRejected start ...")
				if err := t.checkRejected(); err != nil {
					log.Error("checkRejected err: ", err.Error())
				}
				log.Debug("checkRejected end ...")
			case <-tickerTxRejected.C:
				log.Debug("doTxRejected start ...")
				if err := t.doTxRejected(); err != nil {
					log.Error("doTxRejected err: ", err.Error())
				}
				log.Debug("doTxRejected end ...")
			case <-t.ctx.Done():
				log.Debug("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerExpired.C:
				log.Debug("checkExpired start ...")
				if err := t.checkExpired(); err != nil {
					log.Error("checkExpired err: ", err.Error())
				}
				log.Debug("checkExpired end ...")
			case <-t.ctx.Done():
				log.Debug("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerRefundApply.C:
				log.Debug("doRefundApply start ...")
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
				log.Debug("doRefundApply end ...")
			case <-t.ctx.Done():
				log.Debug("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerRecover.C:
				log.Debug("doRecoverCkb start ...")
				if err := t.doRecoverCkb(); err != nil {
					log.Error("doRecoverCkb err: ", err.Error())
				}
				log.Debug("doRecoverCkb end ...")
			case <-t.ctx.Done():
				log.Debug("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerClosedAndUnRefund.C:
				log.Debug("doCheckClosedAndUnRefund start ...")
				if err := t.doCheckClosedAndUnRefund(); err != nil {
					log.Error("doCheckClosedAndUnRefund err: ", err.Error())
				}
				log.Debug("doCheckClosedAndUnRefund end ...")
			case <-tickerResetCoupon.C:
				log.Debug("")
				if err := t.DoResetCoupon(); err != nil {
					log.Error("doResetCoupon err: ", err.Error())
				}
				log.Debug("doResetCoupon end ...")
			case <-t.ctx.Done():
				log.Debug("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	return nil
}
