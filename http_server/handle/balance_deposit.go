package handle

import (
	"das_register_server/http_server/api_code"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqBalanceDeposit struct {
	FromCkbAddress string `json:"from_ckb_address"`
	ToCkbAddress   string `json:"to_ckb_address"`
	Amount         uint64 `json:"amount"`
}

type RespBalanceDeposit struct {
}

func (h *HttpHandle) RpcBalanceDeposit(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqBalanceDeposit
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

	if err = h.doBalanceDeposit(&req[0], apiResp); err != nil {
		log.Error("doBalanceDeposit err:", err.Error())
	}
}

func (h *HttpHandle) BalanceDeposit(ctx *gin.Context) {
	var (
		funcName = "BalanceDeposit"
		clientIp = GetClientIp(ctx)
		req      ReqBalanceDeposit
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

	if err = h.doBalanceDeposit(&req, &apiResp); err != nil {
		log.Error("doBalanceDeposit err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalanceDeposit(req *ReqBalanceDeposit, apiResp *api_code.ApiResp) error {
	var resp RespBalanceDeposit

	apiResp.ApiRespOK(resp)
	return nil
}
