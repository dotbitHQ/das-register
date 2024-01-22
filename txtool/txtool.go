package txtool

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/notify"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"sync"
	"time"
)

var (
	log = logger.NewLogger("txtool", logger.LevelDebug)
)

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
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerApply.C:
				log.Debug("doOrderApplyTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderApplyTx(); err != nil {
						log.Error("doOrderApplyTx err: ", err.Error())
						notify.SendLarkErrNotify(common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("doOrderApplyTx", "", err.Error()))
					}
				}
				log.Debug("doOrderApplyTx end ...")
			case <-tickerPreRegister.C:
				log.Debug("doOrderPreRegisterTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderPreRegisterTx(); err != nil {
						log.Error("doOrderPreRegisterTx err: ", err.Error())
						notify.SendLarkErrNotify(common.DasActionPreRegister, notify.GetLarkTextNotifyStr("doOrderPreRegisterTx", "", err.Error()))
					}
				}
				log.Debug("doOrderPreRegisterTx end ...")
			case <-tickerRenew.C:
				log.Debug("doOrderRenewTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderRenewTx(); err != nil {
						log.Error("doOrderRenewTx err: ", err.Error())
						notify.SendLarkErrNotify(common.DasActionRenewAccount, notify.GetLarkTextNotifyStr("doOrderRenewTx", "", err.Error()))
					}
				}
				log.Debug("doOrderRenewTx end ...")
			case <-t.Ctx.Done():
				log.Debug("tx tool done")
				t.Wg.Done()
				return
			}
		}
	}()
}
