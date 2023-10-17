package txtool

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/notify"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"net"
	"sync"
	"time"
)

var (
	log          = logger.NewLogger("txtool", logger.LevelDebug)
	PromRegister = prometheus.NewRegistry()
)

var Tools *TxTool

type TxTool struct {
	Ctx           context.Context
	Wg            *sync.WaitGroup
	DbDao         *dao.DbDao
	DasCore       *core.DasCore
	DasCache      *dascache.DasCache
	TxBuilderBase *txbuilder.DasTxBuilderBase
	ServerScript  *types.Script
	RebootTime    time.Time
	pusher        *push.Pusher
	Metrics       Metric
}

type Metric struct {
	l         sync.Mutex
	api       *prometheus.SummaryVec
	errNotify *prometheus.CounterVec
}

func (m *Metric) Api() *prometheus.SummaryVec {
	if m.api == nil {
		m.l.Lock()
		defer m.l.Unlock()
		m.api = prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "api",
		}, []string{"method", "http_status", "err_no", "err_msg"})
		PromRegister.MustRegister(m.api)
	}
	return m.api
}

func (m *Metric) ErrNotify() *prometheus.CounterVec {
	if m.errNotify == nil {
		m.l.Lock()
		defer m.l.Unlock()
		m.errNotify = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "notify",
		}, []string{"title", "text"})
		PromRegister.MustRegister(m.errNotify)
	}
	return m.errNotify
}

func Init(txTool *TxTool) {
	Tools = txTool
}

func (t *TxTool) Run() {
	tickerApply := time.NewTicker(time.Second * 5)
	tickerPreRegister := time.NewTicker(time.Second * 6)
	tickerRenew := time.NewTicker(time.Second * 7)
	t.Wg.Add(1)
	errCountApply, errCountPre, errCountRenew := 0, 0, 0
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerApply.C:
				log.Debug("doOrderApplyTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderApplyTx(); err != nil {
						log.Error("doOrderApplyTx err: ", err.Error())
						errCountApply++
						if errCountApply < 50 {
							notify.SendLarkErrNotify(common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("doOrderApplyTx", "", err.Error()))
						}
					} else {
						errCountApply = 0
					}
				}
				log.Debug("doOrderApplyTx end ...")
			case <-tickerPreRegister.C:
				log.Debug("doOrderPreRegisterTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderPreRegisterTx(); err != nil {
						log.Error("doOrderPreRegisterTx err: ", err.Error())
						errCountPre++
						if errCountPre < 50 {
							notify.SendLarkErrNotify(common.DasActionPreRegister, notify.GetLarkTextNotifyStr("doOrderPreRegisterTx", "", err.Error()))
						}
					} else {
						errCountPre = 0
					}
				}
				log.Debug("doOrderPreRegisterTx end ...")
			case <-tickerRenew.C:
				log.Debug("doOrderRenewTx start ...")
				if config.Cfg.Server.TxToolSwitch {
					if err := t.doOrderRenewTx(); err != nil {
						log.Error("doOrderRenewTx err: ", err.Error())
						errCountRenew++
						if errCountRenew < 50 {
							notify.SendLarkErrNotify(common.DasActionRenewAccount, notify.GetLarkTextNotifyStr("doOrderRenewTx", "", err.Error()))
						}
					} else {
						errCountRenew = 0
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

	if config.Cfg.Server.PrometheusPushGateway != "" && config.Cfg.Server.Name != "" {
		t.pusher = push.New(config.Cfg.Server.PrometheusPushGateway, config.Cfg.Server.Name)
		t.pusher.Gatherer(PromRegister)
		t.pusher.Grouping("env", fmt.Sprint(config.Cfg.Server.Net))
		t.pusher.Grouping("instance", GetLocalIp("eth0"))

		go func() {
			ticker := time.NewTicker(time.Second * 5)
			defer ticker.Stop()

			for range ticker.C {
				_ = t.pusher.Push()
			}
		}()
	}
}

func GetLocalIp(interfaceName string) string {
	ief, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Error("GetLocalIp: ", err)
		return ""
	}
	addrs, err := ief.Addrs()
	if err != nil {
		log.Error("GetLocalIp: ", err)
		return ""
	}

	var ipv4Addr net.IP
	for _, addr := range addrs {
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		log.Errorf("GetLocalIp interface %s don't have an ipv4 address", interfaceName)
		return ""
	}
	return ipv4Addr.String()
}
