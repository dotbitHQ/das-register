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

func (h *HttpHandle) Query(ctx *gin.Context) {
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
		log.Error("ShouldBindJSON err:", err.Error(), ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, resp)
		return
	}

	resp.ID, resp.JsonRpc = req.ID, req.JsonRpc
	log.Info("Query:", req.Method, clientIp, toolib.JsonString(req), ctx)

	switch req.Method {
	case api_code_local.MethodTokenList:
		h.RpcTokenList(req.Params, &apiResp)
	case api_code_local.MethodConfigInfo:
		h.RpcConfigInfo(req.Params, &apiResp)
	case api_code_local.MethodAccountList:
		h.RpcAccountList(req.Params, &apiResp)
	case api_code_local.MethodAccountMine:
		h.RpcAccountMine(req.Params, &apiResp)
	case api_code_local.MethodAccountDetail:
		h.RpcAccountDetail(req.Params, &apiResp)
	case api_code_local.MethodAccountRecords:
		h.RpcAccountRecords(req.Params, &apiResp)
	case api_code_local.MethodReverseLatest:
		h.RpcReverseLatest(req.Params, &apiResp)
	case api_code_local.MethodReverseList:
		h.RpcReverseList(req.Params, &apiResp)
	case api_code_local.MethodTransactionStatus:
		h.RpcTransactionStatus(req.Params, &apiResp)
	case api_code_local.MethodBalanceInfo:
		h.RpcBalanceInfo(req.Params, &apiResp)
	case api_code_local.MethodTransactionList:
		h.RpcTransactionList(req.Params, &apiResp)
	case api_code_local.MethodRewardsMine:
		h.RpcRewardsMine(req.Params, &apiResp)
	case api_code_local.MethodWithdrawList:
		h.RpcWithdrawList(req.Params, &apiResp)
	case api_code_local.MethodAccountSearch:
		h.RpcAccountSearch(req.Params, &apiResp)
	case api_code_local.MethodRegisteringList:
		h.RpcRegisteringList(req.Params, &apiResp)
	case api_code_local.MethodOrderDetail:
		h.RpcOrderDetail(req.Params, &apiResp)
	default:
		log.Error("method not exist:", req.Method, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeMethodNotExist, fmt.Sprintf("method [%s] not exits", req.Method))
	}

	api_code_local.DoMonitorLogRpc(&apiResp, req.Method, clientIp, startTime)

	ctx.JSON(http.StatusOK, resp)
	return
}
