package handle

import (
	api_code_local "das_register_server/http_server/api_code"
	"fmt"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

func (h *HttpHandle) Operate(ctx *gin.Context) {
	var (
		req       api_code.JsonRequest
		resp      api_code.JsonResponse
		apiResp   api_code.ApiResp
		clientIp  = GetClientIp(ctx)
		startTime = time.Now()
	)
	resp.Result = &apiResp

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Error("ShouldBindJSON err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, resp)
		return
	}

	resp.ID, resp.JsonRpc = req.ID, req.JsonRpc
	log.Info("Operate:", req.Method, clientIp, toolib.JsonString(req))

	switch req.Method {
	//case api_code_local.MethodReverseDeclare:
	//	h.RpcReverseDeclare(req.Params, &apiResp)
	//case api_code_local.MethodReverseRedeclare:
	//	h.RpcReverseRedeclare(req.Params, &apiResp)
	//case api_code_local.MethodReverseRetract:
	//	h.RpcReverseRetract(req.Params, &apiResp)
	case api_code_local.MethodTransactionSend:
		h.RpcTransactionSend(req.Params, &apiResp)
	case api_code_local.MethodBalancePay:
		h.RpcBalancePay(req.Params, &apiResp)
	case api_code_local.MethodBalanceWithdraw:
		h.RpcBalanceWithdraw(req.Params, &apiResp)
	case api_code_local.MethodBalanceTransfer:
		h.RpcBalanceTransfer(req.Params, &apiResp)
	case api_code_local.MethodEditManager:
		h.RpcEditManager(req.Params, &apiResp)
	case api_code_local.MethodEditOwner:
		h.RpcEditOwner(req.Params, &apiResp)
	case api_code_local.MethodEditRecords:
		h.RpcEditRecords(req.Params, &apiResp)
	case api_code_local.MethodOrderRenew:
		h.RpcOrderRenew(req.Params, &apiResp)
	case api_code_local.MethodOrderRegister:
		h.RpcOrderRegister(req.Params, &apiResp)
	case api_code_local.MethodOrderChange:
		h.RpcOrderChange(req.Params, &apiResp)
	case api_code_local.MethodOrderPayHash:
		h.RpcOrderPayHash(req.Params, &apiResp)
	case api_code_local.MethodOrderCheckCoupon:
		h.RpcCheckCouponr(req.Params, &apiResp)
	default:
		log.Error("method not exist:", req.Method)
		apiResp.ApiRespErr(api_code.ApiCodeMethodNotExist, fmt.Sprintf("method [%s] not exits", req.Method))
	}

	api_code_local.DoMonitorLogRpc(&apiResp, req.Method, clientIp, startTime)

	ctx.JSON(http.StatusOK, resp)
	return
}
