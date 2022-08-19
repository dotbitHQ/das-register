package handle

import (
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqDasOrderDetail struct {
	OrderId string `json:"order_id"`
}

type RespDasOrderDetail struct {
	OrderId        string                `json:"order_id"`
	OrderStatus    tables.OrderStatus    `json:"order_status"`
	Action         common.DasAction      `json:"action"`
	Account        string                `json:"account"`
	AccountId      string                `json:"account_id"`
	PayStatus      tables.TxStatus       `json:"pay_status"`
	RegisterStatus tables.RegisterStatus `json:"register_status"`
}

func (h *HttpHandle) DasOrderDetail(ctx *gin.Context) {
	var (
		funcName = "DasOrderDetail"
		clientIp = GetClientIp(ctx)
		req      ReqDasOrderDetail
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

	if err = h.doDasOrderDetail(&req, &apiResp); err != nil {
		log.Error("doDasOrderDetail err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDasOrderDetail(req *ReqDasOrderDetail, apiResp *api_code.ApiResp) error {
	var resp RespDasOrderDetail

	order, err := h.dbDao.GetOrderByOrderId(req.OrderId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order fail")
		return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
	}
	resp.OrderId = order.OrderId
	resp.OrderStatus = order.OrderStatus
	resp.Action = order.Action
	resp.Account = order.Account
	resp.AccountId = order.AccountId
	resp.PayStatus = order.PayStatus
	resp.RegisterStatus = order.RegisterStatus

	apiResp.ApiRespOK(resp)
	return nil
}
