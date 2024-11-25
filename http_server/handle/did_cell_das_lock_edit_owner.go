package handle

import (
	"context"
	"encoding/json"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqDidCellDasLockEditOwner struct {
	core.ChainTypeAddress
	Account  string `json:"account"`
	RawParam struct {
		ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
		ReceiverAddress  string          `json:"receiver_address"`
	} `json:"raw_param"`
}

type RespDidCellDasLockEditOwner struct {
}

func (h *HttpHandle) RpcDidCellDasLockEditOwner(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqDidCellDasLockEditOwner
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doDidCellDasLockEditOwner(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellDasLockEditOwner err:", err.Error())
	}
}

func (h *HttpHandle) DidCellDasLockEditOwner(ctx *gin.Context) {
	var (
		funcName = "DidCellDasLockEditOwner"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellDasLockEditOwner
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doDidCellDasLockEditOwner(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellDasLockEditOwner err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellDasLockEditOwner(ctx context.Context, req *ReqDidCellDasLockEditOwner, apiResp *api_code.ApiResp) error {
	var resp RespDidCellDasLockEditOwner

	apiResp.ApiRespOK(resp)
	return nil
}
