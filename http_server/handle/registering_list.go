package handle

import (
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqRegisteringList struct {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doRegisteringList(&req, &apiResp); err != nil {
		log.Error("doRegisteringList err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doRegisteringList(req *ReqRegisteringList, apiResp *api_code.ApiResp) error {
	var resp RespRegisteringList
	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.ChainType,
		AddressNormal: req.Address,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address NormalToHex err")
		return fmt.Errorf("NormalToHex err: %s", err.Error())
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
	for _, v := range list {
		resp.RegisteringAccounts = append(resp.RegisteringAccounts, RespRegisteringData{
			Account:       v.Account,
			Status:        v.RegisterStatus,
			CrossCoinType: v.CrossCoinType,
		})
	}

	apiResp.ApiRespOK(resp)
	return nil
}
