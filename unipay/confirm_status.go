package unipay

import (
	"das_register_server/dao"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api"
	"time"
)

func (t *ToolUniPay) RunConfirmStatus() {
	tickerSearchStatus := time.NewTicker(time.Minute * 3)

	t.Wg.Add(1)
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerSearchStatus.C:
				log.Debug("doConfirmStatus start")
				if err := t.doConfirmStatus(); err != nil {
					log.Errorf("doConfirmStatus err: %s", err.Error())
					notify.SendLarkErrNotify("doConfirmStatus", err.Error())
				}
				log.Debug("doConfirmStatus end")
			case <-t.Ctx.Done():
				log.Debug("RunRefund done")
				t.Wg.Done()
				return
			}
		}
	}()
}

func (t *ToolUniPay) doConfirmStatus() error {
	// for check order pay status
	pendingList, err := t.DbDao.GetPayHashStatusPendingList()
	if err != nil {
		return fmt.Errorf("GetPayHashStatusPendingList err: %s", err.Error())
	}
	var orderIdList []string
	for _, v := range pendingList {
		orderIdList = append(orderIdList, v.OrderId)
	}

	// for check refund status
	refundingList, err := t.DbDao.GetRefundStatusRefundingList()
	if err != nil {
		return fmt.Errorf("GetRefundStatusRefundingList err: %s", err.Error())
	}
	var payHashList []string
	for _, v := range refundingList {
		payHashList = append(payHashList, v.Hash)
	}

	if len(orderIdList) == 0 && len(payHashList) == 0 {
		return nil
	}

	log.Info("doConfirmStatus:", len(orderIdList), len(payHashList))

	// call unipay
	resp, err := GetPaymentInfo(ReqPaymentInfo{
		BusinessId:  BusinessIdDasRegisterSvr,
		OrderIdList: orderIdList,
		PayHashList: payHashList,
	})
	if err != nil {
		return fmt.Errorf("OrderInfo err: %s", err.Error())
	}

	var orderIdMap = make(map[string][]PaymentInfo)
	var payHashMap = make(map[string]PaymentInfo)
	for i, v := range resp.PaymentList {
		orderIdMap[v.OrderId] = append(orderIdMap[v.OrderId], resp.PaymentList[i])
		payHashMap[v.PayHash] = resp.PaymentList[i]
	}

	// payment confirm
	for _, pending := range pendingList {
		paymentInfoList, ok := orderIdMap[pending.OrderId]
		if !ok {
			min := pending.PayHashUnconfirmedMin()
			log.Info("PayHashUnconfirmedMin:", pending.OrderId, min)
			//if min > 60 {
			//	notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Payment not completed", pending.OrderId)
			//	if err := t.DbDao.UpdateUniPayUnconfirmedToRejected(pending.OrderId, pending.Hash); err != nil {
			//		return fmt.Errorf("UpdateUniPayUnconfirmedToRejected err: %s", err.Error())
			//	}
			//}
			continue
		}
		for _, v := range paymentInfoList {
			if v.PayHashStatus != tables.PayHashStatusConfirmed {
				continue
			}
			if err = DoPaymentConfirm(t.DbDao, v.OrderId, v.PayHash, v.PayAddress, v.AlgorithmId); err != nil {
				log.Errorf("DoPaymentConfirm err: %s", err.Error())
				notify.SendLarkErrNotify("DoPaymentConfirm", err.Error())
			}
		}
	}

	// refund confirm
	for _, v := range payHashList {
		paymentInfo, ok := payHashMap[v]
		if !ok {
			continue
		}
		if paymentInfo.RefundStatus != tables.UniPayRefundStatusRefunded {
			continue
		}
		if err := t.DbDao.UpdateUniPayRefundStatusToRefunded(paymentInfo.PayHash, paymentInfo.OrderId, paymentInfo.RefundHash); err != nil {
			log.Error("UpdateUniPayRefundStatusToRefunded err: ", err.Error())
			notify.SendLarkErrNotify("UpdateUniPayRefundStatusToRefunded", err.Error())
		}
	}

	return nil
}

func DoPaymentConfirm(dbDao *dao.DbDao, orderId, payHash, payAddress string, algorithmId common.DasAlgorithmId) error {
	orderInfo, err := dbDao.GetOrderByOrderId(orderId)
	if err != nil {
		return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
	}
	paymentInfo := tables.TableDasOrderPayInfo{
		Id:                 0,
		Hash:               payHash,
		OrderId:            orderId,
		ChainType:          algorithmId.ToChainType(),
		Address:            payAddress,
		Status:             tables.OrderTxStatusConfirm,
		Timestamp:          time.Now().UnixNano() / 1e6,
		AccountId:          orderInfo.AccountId,
		RefundStatus:       tables.TxStatusDefault,
		UniPayRefundStatus: tables.UniPayRefundStatusDefault,
		RefundHash:         "",
	}
	if err = dbDao.UpdatePayment(paymentInfo); err != nil {
		return fmt.Errorf("UpdatePayment err: %s", err.Error())
	}
	return nil
}
