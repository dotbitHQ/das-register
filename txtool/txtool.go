package txtool

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/notify"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

var log = mylog.NewLogger("txtool", mylog.LevelDebug)

type TxTool struct {
	Ctx           context.Context
	Wg            *sync.WaitGroup
	DbDao         *dao.DbDao
	DasCore       *core.DasCore
	DasCache      *dascache.DasCache
	TxBuilderBase *txbuilder.DasTxBuilderBase
	ServerScript  *types.Script
	RebootTime    time.Time
}

func (t *TxTool) Run() {
	tickerApply := time.NewTicker(time.Second * 5)
	tickerPreRegister := time.NewTicker(time.Second * 6)
	tickerRenew := time.NewTicker(time.Second * 7)
	t.Wg.Add(1)
	errCountApply, errCountPre, errCountRenew := 0, 0, 0
	go func() {
		for {
			select {
			case <-tickerApply.C:
				log.Info("doOrderApplyTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderApplyTx(); err != nil {
						log.Error("doOrderApplyTx err: ", err.Error())
						errCountApply++
						if errCountApply < 50 {
							notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("doOrderApplyTx", "", err.Error()))
						}
					} else {
						errCountApply = 0
					}
				}
				log.Info("doOrderApplyTx end ...")
			case <-tickerPreRegister.C:
				log.Info("doOrderPreRegisterTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderPreRegisterTx(); err != nil {
						log.Error("doOrderPreRegisterTx err: ", err.Error())
						errCountPre++
						if errCountPre < 50 {
							notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionPreRegister, notify.GetLarkTextNotifyStr("doOrderPreRegisterTx", "", err.Error()))
						}
					} else {
						errCountPre = 0
					}
				}
				log.Info("doOrderPreRegisterTx end ...")
			case <-tickerRenew.C:
				log.Info("doOrderRenewTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderRenewTx(); err != nil {
						log.Error("doOrderRenewTx err: ", err.Error())
						errCountRenew++
						if errCountRenew < 50 {
							notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionRenewAccount, notify.GetLarkTextNotifyStr("doOrderRenewTx", "", err.Error()))
						}
					} else {
						errCountRenew = 0
					}
				}
				log.Info("doOrderRenewTx end ...")
			case <-t.Ctx.Done():
				log.Info("tx tool done")
				t.Wg.Done()
				return
			}
		}
	}()
}
