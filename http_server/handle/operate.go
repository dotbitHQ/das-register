package handle

import (
	"das_register_server/http_server/api_code"
	"fmt"
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
	case api_code.MethodReverseDeclare:
		h.RpcReverseDeclare(req.Params, &apiResp)
	case api_code.MethodReverseRedeclare:
		h.RpcReverseRedeclare(req.Params, &apiResp)
	case api_code.MethodReverseRetract:
		h.RpcReverseRetract(req.Params, &apiResp)
	case api_code.MethodTransactionSend:
		h.RpcTransactionSend(req.Params, &apiResp)
	case api_code.MethodBalancePay:
		h.RpcBalancePay(req.Params, &apiResp)
	case api_code.MethodBalanceWithdraw:
		h.RpcBalanceWithdraw(req.Params, &apiResp)
	case api_code.MethodBalanceTransfer:
		h.RpcBalanceTransfer(req.Params, &apiResp)
	case api_code.MethodEditManager:
		h.RpcEditManager(req.Params, &apiResp)
	case api_code.MethodEditOwner:
		h.RpcEditOwner(req.Params, &apiResp)
	case api_code.MethodEditRecords:
		h.RpcEditRecords(req.Params, &apiResp)
	case api_code.MethodOrderRenew:
		h.RpcOrderRenew(req.Params, &apiResp)
	case api_code.MethodOrderRegister:
		h.RpcOrderRegister(req.Params, &apiResp)
	case api_code.MethodOrderChange:
		h.RpcOrderChange(req.Params, &apiResp)
	case api_code.MethodOrderPayHash:
		h.RpcOrderPayHash(req.Params, &apiResp)
	default:
		log.Error("method not exist:", req.Method)
		apiResp.ApiRespErr(api_code.ApiCodeMethodNotExist, fmt.Sprintf("method [%s] not exits", req.Method))
	}

	api_code.DoMonitorLogRpc(&apiResp, req.Method, clientIp, startTime)

	ctx.JSON(http.StatusOK, resp)
	return
}
