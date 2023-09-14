package handle

import (
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

type ReqAccountList struct {
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
}

type RespAccountList struct {
	List []AccountData `json:"list"`
}

type AccountData struct {
	Account              string                  `json:"account"`
	Status               tables.SearchStatus     `json:"status"`
	ExpiredAt            int64                   `json:"expired_at"`
	RegisteredAt         int64                   `json:"registered_at"`
	EnableSubAccount     tables.EnableSubAccount `json:"enable_sub_account"`
	RenewSubAccountPrice uint64                  `json:"renew_sub_account_price"`
	Nonce                uint64                  `json:"nonce"`
}

func (h *HttpHandle) RpcAccountList(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountList
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

	if err = h.doAccountList(&req[0], apiResp); err != nil {
		log.Error("doAccountList err:", err.Error())
	}
}

func (h *HttpHandle) AccountList(ctx *gin.Context) {
	var (
		funcName = "AccountList"
		clientIp = GetClientIp(ctx)
		req      ReqAccountList
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

	if err = h.doAccountList(&req, &apiResp); err != nil {
		log.Error("doAccountList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountList(req *ReqAccountList, apiResp *api_code.ApiResp) error {
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

	var resp RespAccountList
	resp.List = make([]AccountData, 0)

	list, err := h.dbDao.SearchAccountList(req.ChainType, req.Address)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account list err")
		return fmt.Errorf("SearchAccountList err: %s", err.Error())
	}
	for _, v := range list {
		resp.List = append(resp.List, AccountData{
			Account:              v.Account,
			Status:               v.FormatAccountStatus(),
			ExpiredAt:            int64(v.ExpiredAt) * 1e3,
			RegisteredAt:         int64(v.RegisteredAt) * 1e3,
			EnableSubAccount:     v.EnableSubAccount,
			RenewSubAccountPrice: v.RenewSubAccountPrice,
			Nonce:                v.Nonce,
		})
	}

	apiResp.ApiRespOK(resp)
	return nil
}
