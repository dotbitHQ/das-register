package handle

import (
	"das_register_server/http_server/compatible"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqRegisteringList struct {
	core.ChainTypeAddress
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
}

type RespRegisteringList struct {
	RegisteringAccounts []RespRegisteringData `json:"registering_accounts"`
}

type RespRegisteringData struct {
	Account       string                `json:"account"`
	Status        tables.RegisterStatus `json:"status"`
	CrossCoinType string                `json:"cross_coin_type"`
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

	if err = h.doRegisteringList(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doRegisteringList(&req, &apiResp); err != nil {
		log.Error("doRegisteringList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doRegisteringList(req *ReqRegisteringList, apiResp *api_code.ApiResp) error {
	var resp RespRegisteringList
	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex
	resp.RegisteringAccounts = make([]RespRegisteringData, 0)

	if req.Address == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}

	list, err := h.dbDao.GetRegisteringOrders(req.ChainType, req.Address)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "get registering account fail")
		return fmt.Errorf("GetRegisteringOrders err: %s", err.Error())
	}
	var accMap = make(map[string]int)
	for _, v := range list {
		if item, ok := accMap[v.AccountId]; ok {
			if v.RegisterStatus > resp.RegisteringAccounts[item].Status {
				resp.RegisteringAccounts[item].Status = v.RegisterStatus
				resp.RegisteringAccounts[item].CrossCoinType = v.CrossCoinType
			}
		} else {
			resp.RegisteringAccounts = append(resp.RegisteringAccounts, RespRegisteringData{
				Account:       v.Account,
				Status:        v.RegisterStatus,
				CrossCoinType: v.CrossCoinType,
			})
			accMap[v.AccountId] = len(resp.RegisteringAccounts) - 1
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
