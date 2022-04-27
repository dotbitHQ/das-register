package handle

import (
	"das_register_server/cache"
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

type ReqAccountMine struct {
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Pagination
	Keyword string `json:"keyword"`
}

type RespAccountMine struct {
	Total int64         `json:"total"`
	List  []AccountData `json:"list"`
}

func (h *HttpHandle) RpcAccountMine(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountMine
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

	if err = h.doAccountMine(&req[0], apiResp); err != nil {
		log.Error("doAccountMine err:", err.Error())
	}
}

func (h *HttpHandle) AccountMine(ctx *gin.Context) {
	var (
		funcName = "AccountMine"
		clientIp = GetClientIp(ctx)
		req      ReqAccountMine
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

	if err = h.doAccountMine(&req, &apiResp); err != nil {
		log.Error("doAccountMine err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountMine(req *ReqAccountMine, apiResp *api_code.ApiResp) error {
	var resp RespAccountMine
	resp.List = make([]AccountData, 0)

	action := "AccountMine"
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

	if req.Keyword != "" {
		if err := h.rc.LockWithRedis(req.ChainType, req.Address, action, time.Millisecond*600); err != nil {
			if err == cache.ErrDistributedLockPreemption {
				apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "The operation is too frequent")
				return nil
			}
		}
	}

	list, err := h.dbDao.SearchAccountListWithPage(req.ChainType, req.Address, req.Keyword, req.GetLimit(), req.GetOffset())
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

	count, err := h.dbDao.GetAccountsCount(req.ChainType, req.Address, req.Keyword)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "get account count err")
		return fmt.Errorf("GetAccountsCount err: %s", err.Error())
	}
	resp.Total = count

	apiResp.ApiRespOK(resp)
	return nil
}
