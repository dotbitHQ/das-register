package handle

import (
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
)

type ReqWithdrawList struct {
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

	if err = h.doWithdrawList(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doWithdrawList(&req, &apiResp); err != nil {
		log.Error("doWithdrawList err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doWithdrawList(req *ReqWithdrawList, apiResp *api_code.ApiResp) error {
	var resp RespWithdrawList
	resp.List = make([]WithdrawListData, 0)

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
