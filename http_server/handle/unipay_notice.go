package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/notify"
	"das_register_server/tables"
	"das_register_server/unipay"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type EventType string

const (
	EventTypeOrderPay       EventType = "ORDER.PAY"
	EventTypeOrderRefund    EventType = "ORDER.REFUND"
	EventTypePaymentDispute EventType = "PAYMENT.DISPUTE"
)

type EventInfo struct {
	EventType    EventType                 `json:"event_type"`
	OrderId      string                    `json:"order_id"`
	PayStatus    tables.UniPayStatus       `json:"pay_status"`
	PayHash      string                    `json:"pay_hash"`
	PayAddress   string                    `json:"pay_address"`
	AlgorithmId  common.DasAlgorithmId     `json:"algorithm_id"`
	RefundStatus tables.UniPayRefundStatus `json:"refund_status"`
	RefundHash   string                    `json:"refund_hash"`
}

type ReqUniPayNotice struct {
	BusinessId string      `json:"business_id"`
	EventList  []EventInfo `json:"event_list"`
}

type RespUniPayNotice struct {
}

func (h *HttpHandle) UniPayNotice(ctx *gin.Context) {
	var (
		funcName = "UniPayNotice"
		clientIp = GetClientIp(ctx)
		req      ReqUniPayNotice
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doUniPayNotice(&req, &apiResp); err != nil {
		log.Error("doUniPayNotice err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doUniPayNotice(req *ReqUniPayNotice, apiResp *api_code.ApiResp) error {
	var resp RespUniPayNotice

	// check BusinessId
	if req.BusinessId != unipay.BusinessIdDasRegisterSvr {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("BusinessId[%s] invalid", req.BusinessId))
		return nil
	}
	// check order id
	for _, v := range req.EventList {
		switch v.EventType {
		case EventTypeOrderPay:
			if err := unipay.DoPaymentConfirm(h.dbDao, v.OrderId, v.PayHash, v.PayAddress, v.AlgorithmId); err != nil {
				log.Error("DoPaymentConfirm err: ", err.Error(), v.OrderId, v.PayHash)
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "DoPaymentConfirm", err.Error())
			}
		case EventTypeOrderRefund:
			if err := h.dbDao.UpdateUniPayRefundStatusToRefunded(v.PayHash, v.OrderId, v.RefundHash); err != nil {
				log.Error("UpdateUniPayRefundStatusToRefunded err: ", err.Error())
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "UpdateUniPayRefundStatusToRefunded", err.Error())
			}
		case EventTypePaymentDispute:
			if err := h.dbDao.UpdatePayHashStatusToFailByDispute(v.PayHash, v.OrderId); err != nil {
				log.Error("UpdatePayHashStatusToFailByDispute err: ", err.Error())
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "UpdatePayHashStatusToFailByDispute", err.Error())
			}
		default:
			log.Error("EventType invalid:", v.EventType)
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
