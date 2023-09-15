package unipay

import (
	"das_register_server/config"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/http_api"
	"time"
)

func (t *ToolUniPay) RunOrderRefund() {
	tickerRefund := time.NewTicker(time.Minute * 10)

	t.Wg.Add(1)
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerRefund.C:
				log.Debug("doRefund start")
				if err := t.doRefund(); err != nil {
					log.Errorf("doRefund err: %s", err.Error())
					notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "doRefund", err.Error())
				}
				log.Debug("doRefund end")
			case <-t.Ctx.Done():
				log.Debug("RunRefund done")
				t.Wg.Done()
				return
			}
		}
	}()
}

func (t *ToolUniPay) doRefund() error {
	if !config.Cfg.Server.UniPayRefundSwitch {
		return nil
	}
	//get payment list
	list, err := t.DbDao.GetUnRefundList()
	if err != nil {
		return fmt.Errorf("GetUnRefundList err: %s", err.Error())
	}

	// check is unipay order
	var orderIds []string
	for _, v := range list {
		orderIds = append(orderIds, v.OrderId)
	}
	orders, err := t.DbDao.GetIsUniPayInfoByOrderIds(orderIds)
	if err != nil {
		return fmt.Errorf("GetOrderListByOrderIds err: %s", err.Error())
	}
	var isUniPayMap = make(map[string]tables.IsUniPay)
	for _, v := range orders {
		isUniPayMap[v.OrderId] = v.IsUniPay
	}
	for _, v := range list {
		if isUniPay := isUniPayMap[v.OrderId]; isUniPay == tables.IsUniPayFalse {
			if err := t.DbDao.UpdateUniPayRefundStatusToDefaultForNotUniPayOrder(v.Hash, v.OrderId); err != nil {
				return fmt.Errorf("UpdateUniPayRefundStatusToDefaultForNotUniPayOrder err: %s", err.Error())
			}
		}
	}

	//call unipay to refund
	var req ReqOrderRefund
	req.BusinessId = BusinessIdDasRegisterSvr
	var ids []uint64
	for _, v := range list {
		ids = append(ids, v.Id)
		req.RefundList = append(req.RefundList, RefundInfo{
			OrderId: v.OrderId,
			PayHash: v.Hash,
		})
	}

	_, err = RefundOrder(req)
	if err != nil {
		return fmt.Errorf("RefundOrder err: %s", err.Error())
	}

	if err = t.DbDao.UpdateRefundStatusToRefundIng(ids); err != nil {
		return fmt.Errorf("UpdateRefundStatusToRefundIng err: %s", err.Error())
	}

	return nil
}
