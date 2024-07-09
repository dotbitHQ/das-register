package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
)

type ReqDidCellUpgradePrice struct {
	core.ChainTypeAddress
	Account string `json:"account"`
}

type RespDidCellUpgradePrice struct {
	PriceUSD decimal.Decimal `json:"price_usd"`
	PriceCKB decimal.Decimal `json:"price_ckb"`
}

func (h *HttpHandle) RpcDidCellUpgradePrice(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellUpgradePrice
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doDidCellUpgradePrice(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellUpgradePrice err:", err.Error())
	}
}

func (h *HttpHandle) DidCellUpgradePrice(ctx *gin.Context) {
	var (
		funcName = "DidCellUpgradePrice"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellUpgradePrice
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doDidCellUpgradePrice(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellUpgradePrice err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellUpgradePrice(ctx context.Context, req *ReqDidCellUpgradePrice, apiResp *http_api.ApiResp) error {
	var resp RespDidCellUpgradePrice

	req.Account = strings.ToLower(req.Account)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address is invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	} else if addrHex.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}

	quoteCell, err := h.dasCore.GetQuoteCell()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to get quote cell")
		return fmt.Errorf("GetQuoteCell err: %s", err.Error())
	}
	quote := quoteCell.Quote()

	editOwnerLock := addrHex.ParsedAddress.Script
	editOwnerCapacity, err := h.dasCore.GetDidCellOccupiedCapacity(editOwnerLock, req.Account)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to get did cell capacity")
		return fmt.Errorf("GetDidCellOccupiedCapacity err: %s", err.Error())
	}

	editOwnerAmountUSD, _ := decimal.NewFromString(fmt.Sprintf("%d", editOwnerCapacity/common.OneCkb))
	decQuote, _ := decimal.NewFromString(fmt.Sprintf("%d", quote))
	decUsdRateBase := decimal.NewFromInt(common.UsdRateBase)

	resp.PriceUSD = editOwnerAmountUSD.Mul(decQuote).DivRound(decUsdRateBase, 6)
	resp.PriceCKB = decimal.NewFromInt(int64(editOwnerCapacity))

	apiResp.ApiRespOK(resp)
	return nil
}
