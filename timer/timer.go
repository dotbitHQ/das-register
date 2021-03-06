package timer

import (
	"context"
	"das_register_server/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
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
	txBuilderBase *txbuilder.DasTxBuilderBase
}

type TxTimerParam struct {
	DbDao         *dao.DbDao
	Ctx           context.Context
	Wg            *sync.WaitGroup
	DasCore       *core.DasCore
	TxBuilderBase *txbuilder.DasTxBuilderBase
}

func NewTxTimer(p TxTimerParam) *TxTimer {
	var t TxTimer
	t.ctx = p.Ctx
	t.wg = p.Wg
	t.dbDao = p.DbDao
	t.dasCore = p.DasCore
	t.txBuilderBase = p.TxBuilderBase
	return &t
}

func (t *TxTimer) Run() error {
	if err := t.doUpdateTokenMap(); err != nil {
		return fmt.Errorf("doUpdateTokenMap init token err: %s", err.Error())
	}
	tickerToken := time.NewTicker(time.Second * 50)
	tickerRejected := time.NewTicker(time.Minute * 3)

	tickerExpired := time.NewTicker(time.Minute * 30)
	tickerRecover := time.NewTicker(time.Minute * 30)
	tickerRefundApply := time.NewTicker(time.Minute * 15)
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
				if err := t.doRefundApply(); err != nil {
					log.Error("doRefundApply err: ", err.Error())
				}
				if err := t.doRefundPre(); err != nil {
					log.Error("doRefundPre err: %s", err.Error())
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
					log.Error("doCheckClosedAndUnRefund err: %s", err.Error())
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
