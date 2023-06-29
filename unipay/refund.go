package unipay

import (
	"das_register_server/config"
	"das_register_server/notify"
	"fmt"
	"time"
)

func (t *ToolUniPay) RunOrderRefund() {
	tickerRefund := time.NewTicker(time.Minute * 10)

	t.Wg.Add(1)
	go func() {
		for {
			select {
			case <-tickerRefund.C:
				log.Info("doRefund start")
				if err := t.doRefund(); err != nil {
					log.Errorf("doRefund err: %s", err.Error())
					notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "doRefund", err.Error())
				}
				log.Info("doRefund end")
			case <-t.Ctx.Done():
				log.Info("RunRefund done")
				t.Wg.Done()
				return
			}
		}
	}()
}

func (t *ToolUniPay) doRefund() error {
	if !config.Cfg.Server.RefundSwitch {
		return nil
	}
	//get payment list
	list, err := t.DbDao.GetUnRefundList()
	if err != nil {
		return fmt.Errorf("GetUnRefundList err: %s", err.Error())
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
