package handle

import (
	"context"
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
	"time"
)

type ReqAccountMine struct {
	core.ChainTypeAddress
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Pagination
	Keyword  string          `json:"keyword"`
	Category tables.Category `json:"category"`
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

	if err = h.doAccountMine(h.ctx, &req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doAccountMine(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAccountMine err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountMine(ctx context.Context, req *ReqAccountMine, apiResp *api_code.ApiResp) error {
	var resp RespAccountMine
	resp.List = make([]AccountData, 0)

	req.Keyword = strings.ToLower(req.Keyword)
	action := "AccountMine"

	addressHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	if req.Keyword != "" {
		if err := h.rc.LockWithRedis(ctx, req.ChainType, req.Address, action, time.Millisecond*600); err != nil {
			if err == cache.ErrDistributedLockPreemption {
				apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "The operation is too frequent")
				return nil
			}
		}
	}

	list, err := h.dbDao.SearchAccountListWithPage(req.ChainType, req.Address, req.Keyword, req.GetLimit(), req.GetOffset(), req.Category)
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

	count, err := h.dbDao.GetAccountsCount(req.ChainType, req.Address, req.Keyword, req.Category)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "get account count err")
		return fmt.Errorf("GetAccountsCount err: %s", err.Error())
	}
	resp.Total = count

	apiResp.ApiRespOK(resp)
	return nil
}
