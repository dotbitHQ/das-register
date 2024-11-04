package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqRegisteringList struct {
	core.ChainTypeAddress
	addressHex *core.DasAddressHex
}

type RespRegisteringList struct {
	RegisteringAccounts []RespRegisteringData `json:"registering_accounts"`
}

type RespRegisteringData struct {
	Account string                `json:"account"`
	Status  tables.RegisterStatus `json:"status"`
}

func (h *HttpHandle) RpcRegisteringList(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqRegisteringList
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

	if err = h.doRegisteringList(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doRegisteringList err:", err.Error())
	}
}

func (h *HttpHandle) RegisteringList(ctx *gin.Context) {
	var (
		funcName = "RegisteringList"
		clientIp = GetClientIp(ctx)
		req      ReqRegisteringList
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doRegisteringList(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doRegisteringList err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doRegisteringList(ctx context.Context, req *ReqRegisteringList, apiResp *api_code.ApiResp) error {
	var resp RespRegisteringList

	var err error
	req.addressHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}

	resp.RegisteringAccounts = make([]RespRegisteringData, 0)

	list, err := h.dbDao.GetRegisteringOrders(req.addressHex.ChainType, req.addressHex.AddressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "get registering account fail")
		return fmt.Errorf("GetRegisteringOrders err: %s", err.Error())
	}
	var accMap = make(map[string]int)
	for _, v := range list {
		if item, ok := accMap[v.AccountId]; ok {
			if v.RegisterStatus > resp.RegisteringAccounts[item].Status {
				resp.RegisteringAccounts[item].Status = v.RegisterStatus
			}
		} else {
			resp.RegisteringAccounts = append(resp.RegisteringAccounts, RespRegisteringData{
				Account: v.Account,
				Status:  v.RegisterStatus,
			})
			accMap[v.AccountId] = len(resp.RegisteringAccounts) - 1
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
