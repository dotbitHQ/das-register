package unipay

import (
	"das_register_server/config"
	"das_register_server/notify"
	"das_register_server/tables"
	"das_register_server/timer"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

func (t *ToolUniPay) RunDoOrderHedge() {
	tickerHedge := time.NewTicker(time.Minute * 3)

	t.Wg.Add(1)
	go func() {
		for {
			select {
			case <-tickerHedge.C:
				log.Info("doOrderHedge start")
				if err := t.doOrderHedge(); err != nil {
					log.Errorf("doOrderHedge err: %s", err.Error())
					notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "doOrderHedge", err.Error())
				}
				log.Info("doOrderHedge end")
			case <-t.Ctx.Done():
				log.Info("RunDoOrderHedge done")
				t.Wg.Done()
				return
			}
		}
	}()
}

func (t *ToolUniPay) doOrderHedge() error {
	list, err := t.DbDao.GetNeedHedgeOrderList()
	if err != nil {
		return fmt.Errorf("GetNeedHedgeOrderList err: %s", err.Error())
	}
	for _, v := range list {
		switch v.PayTokenId {
		case tables.TokenCoupon, tables.TokenIdCkb,
			tables.TokenIdCkbInternal, tables.TokenIdDas:
			if err = t.DbDao.UpdateHedgeStatus(v.OrderId, tables.TxStatusSending, tables.TxStatusOk); err != nil {
				return fmt.Errorf("UpdateHedgeStatus err: %s", err.Error())
			}
			continue
		}
		// pay amount check
		payToken := timer.GetTokenInfo(v.PayTokenId)
		if payToken.Id <= 0 || payToken.Price.Cmp(decimal.Zero) != 1 {
			return fmt.Errorf("GetTokenInfo err: %s", v.PayTokenId)
		}
		payAmount := v.PayAmount.DivRound(decimal.New(1, payToken.Decimals), payToken.Decimals)
		req := ReqHedge{
			OrderId:    v.OrderId,
			PayTokenId: v.PayTokenId,
			PayAmount:  payAmount,
		}
		// update order
		if err := t.DbDao.UpdateHedgeStatus(v.OrderId, tables.TxStatusSending, tables.TxStatusOk); err != nil {
			return fmt.Errorf("UpdateHedgeStatus err: %s", err.Error())
		}
		//
		if err := t.doHedge(req); err != nil {
			log.Error("doHedge err: ", err.Error(), req.OrderId)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "do hedge", notify.GetLarkTextNotifyStr("doHedge", req.OrderId, err.Error()))
			continue
		}
	}
	return nil
}

type ReqHedge struct {
	OrderId    string            `json:"orderId"`
	PayTokenId tables.PayTokenId `json:"payTokenId"`
	PayAmount  decimal.Decimal   `json:"payAmount"`
}

type RespDeposit struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (t *ToolUniPay) doHedge(req ReqHedge) error {
	if req.PayTokenId == tables.TokenIdCkb || req.PayTokenId == tables.TokenIdDas || req.PayTokenId == tables.TokenIdCkbInternal {
		return nil
	}
	var res RespDeposit
	url := config.Cfg.Server.HedgeUrl
	if url == "" {
		return nil
	}
	resp, body, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&res)
	if len(errs) > 0 {
		return fmt.Errorf("doHedge errs:%+v %s", errs, string(body))
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("doHedge resp.StatusCode:%d", resp.StatusCode)
	} else if res.Code != 0 {
		return fmt.Errorf("doHedge res.Code:%d [%s]", res.Code, res.Message)
	}
	return nil
}
