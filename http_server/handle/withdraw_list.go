package handle

import (
	"context"
	"das_register_server/http_server/compatible"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
)

type ReqWithdrawList struct {
	core.ChainTypeAddress
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Pagination
}

type RespWithdrawList struct {
	Count int64              `json:"count"`
	Total decimal.Decimal    `json:"total"`
	List  []WithdrawListData `json:"list"`
}

type WithdrawListData struct {
	Hash        string          `json:"hash"`
	BlockNumber uint64          `json:"block_number"`
	Amount      decimal.Decimal `json:"amount"`
}

func (h *HttpHandle) RpcWithdrawList(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqWithdrawList
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

	if err = h.doWithdrawList(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doWithdrawList err:", err.Error())
	}
}

func (h *HttpHandle) WithdrawList(ctx *gin.Context) {
	var (
		funcName = "WithdrawList"
		clientIp = GetClientIp(ctx)
		req      ReqWithdrawList
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

	if err = h.doWithdrawList(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doWithdrawList err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doWithdrawList(ctx context.Context, req *ReqWithdrawList, apiResp *api_code.ApiResp) error {
	var resp RespWithdrawList
	resp.List = make([]WithdrawListData, 0)

	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	list, err := h.dbDao.GetTransactionListByAction(req.ChainType, req.Address, common.DasActionWithdrawFromWallet, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search withdraw list err")
		return fmt.Errorf("GetTransactionListByAction err: %s", err.Error())
	}

	for _, v := range list {
		hash, _ := common.String2OutPoint(v.Outpoint)
		amount, _ := decimal.NewFromString(fmt.Sprintf("%d", v.Capacity))
		resp.List = append(resp.List, WithdrawListData{
			Hash:        hash,
			BlockNumber: v.BlockNumber,
			Amount:      amount,
		})
	}

	tt, err := h.dbDao.GetTransactionTotalCapacityByAction(req.ChainType, req.Address, common.DasActionWithdrawFromWallet)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search withdraw list err")
		return fmt.Errorf("GetTransactionTotalCapacityByAction err: %s", err.Error())
	}
	resp.Total = tt.TotalCapacity
	resp.Count = tt.CountNumber

	apiResp.ApiRespOK(resp)
	return nil
}
