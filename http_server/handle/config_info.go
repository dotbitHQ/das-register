package handle

import (
	"context"
	"das_register_server/config"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
)

type RespConfigInfo struct {
	ReverseRecordCapacity uint64 `json:"reverse_record_capacity"` // 反向解析需要的 capacity
	MinChangeCapacity     uint64 `json:"min_change_capacity"`     // 最小找零 capacity
	SaleCellCapacity      uint64 `json:"sale_cell_capacity"`      // 上架质押
	MinSellPrice          uint64 `json:"min_sell_price"`          // 最小出售金额

	AccountExpirationGracePeriod uint32          `json:"account_expiration_grace_period"`
	MinTtl                       uint32          `json:"min_ttl"`
	ProfitRateOfInviter          decimal.Decimal `json:"profit_rate_of_inviter"`
	InviterDiscount              decimal.Decimal `json:"inviter_discount"`
	MinAccountLen                uint32          `json:"min_account_len"`
	MaxAccountLen                uint32          `json:"max_account_len"`
	EditRecordsThrottle          uint32          `json:"edit_records_throttle"`
	EditManagerThrottle          uint32          `json:"edit_manager_throttle"`
	TransferThrottle             uint32          `json:"transfer_throttle"`
	IncomeCellMinTransferValue   uint64          `json:"income_cell_min_transfer_value"`
	Premium                      decimal.Decimal `json:"premium" yaml:"premium"`
	TimestampOnChain             int64           `json:"timestamp_on_chain"`
	PremiumPercentage            decimal.Decimal `json:"premium_percentage"`
	PremiumBase                  decimal.Decimal `json:"premium_base"`
}

func (h *HttpHandle) RpcConfigInfo(p json.RawMessage, apiResp *api_code.ApiResp) {
	if err := h.doConfigInfo(h.ctx, apiResp); err != nil {
		log.Error("doConfigInfo err:", err.Error())
	}
}

func (h *HttpHandle) ConfigInfo(ctx *gin.Context) {
	var (
		funcName = "ConfigInfo"
		clientIp = GetClientIp(ctx)
		apiResp  api_code.ApiResp
		err      error
	)

	log.Info("ApiReq:", funcName, clientIp, ctx)

	if err = h.doConfigInfo(ctx.Request.Context(), &apiResp); err != nil {
		log.Error("doConfigInfo err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doConfigInfo(ctx context.Context, apiResp *api_code.ApiResp) error {
	var resp RespConfigInfo
	resp.PremiumPercentage = config.Cfg.Stripe.PremiumPercentage
	resp.PremiumBase = config.Cfg.Stripe.PremiumBase

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	//
	builder, err := h.dasCore.ConfigCellDataBuilderByTypeArgsList(
		common.ConfigCellTypeArgsIncome,
		common.ConfigCellTypeArgsAccount,
		common.ConfigCellTypeArgsPrice,
		common.ConfigCellTypeArgsProfitRate,
		common.ConfigCellTypeArgsSecondaryMarket,
		common.ConfigCellTypeArgsReverseRecord,
	)
	var preparedFee, baseCapacity uint64
	var saleCellBasicCapacity, saleCellPreparedFeeCapacity, saleMinPrice uint64
	var inviteDiscount, profitRateInviter uint32
	var incomeCellMinTransferValue uint64
	var transferThrottle, editManagerThrottle, editRecordsThrottle, maxAccountLen, minTtl, accountExpirationGracePeriod uint32
	err = errors.New("test config cell cache")
	if err != nil {
		log.Error(err.Error())
		var cacheBuilder core.CacheConfigCellBase
		strCache, errCache := h.dasCore.GetConfigCellByCache(core.CacheConfigCellKeyBase)
		if errCache != nil {
			log.Error("GetConfigCellByCache err: %s", errCache.Error())
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("ConfigCellDataBuilderByTypeArgsList1 err: %s", err.Error())
		} else if strCache == "" {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("ConfigCellDataBuilderByTypeArgsList2 err: %s", err.Error())
		} else if errCache = json.Unmarshal([]byte(strCache), &cacheBuilder); errCache != nil {
			log.Error("json.Unmarshal err: %s", errCache.Error())
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("ConfigCellDataBuilderByTypeArgsList3 err: %s", err.Error())
		}
		preparedFee = cacheBuilder.RecordPreparedFeeCapacity
		baseCapacity = cacheBuilder.RecordBasicCapacity
		saleCellBasicCapacity = cacheBuilder.SaleCellBasicCapacity
		saleCellPreparedFeeCapacity = cacheBuilder.SaleCellPreparedFeeCapacity
		saleMinPrice = cacheBuilder.SaleMinPrice
		inviteDiscount = cacheBuilder.PriceInvitedDiscount
		profitRateInviter = cacheBuilder.ProfitRateInviter

		incomeCellMinTransferValue = cacheBuilder.IncomeMinTransferCapacity
		transferThrottle = cacheBuilder.TransferAccountThrottle
		editManagerThrottle = cacheBuilder.EditManagerThrottle
		editRecordsThrottle = cacheBuilder.EditRecordsThrottle
		maxAccountLen = cacheBuilder.MaxLength
		minTtl = cacheBuilder.RecordMinTtl
		accountExpirationGracePeriod = cacheBuilder.ExpirationGracePeriod
	} else {
		preparedFee, _ = builder.RecordPreparedFeeCapacity()
		baseCapacity, _ = builder.RecordBasicCapacity()
		saleCellBasicCapacity, _ = builder.SaleCellBasicCapacity()
		saleCellPreparedFeeCapacity, _ = builder.SaleCellPreparedFeeCapacity()
		saleMinPrice, _ = builder.SaleMinPrice()
		inviteDiscount, _ = builder.PriceInvitedDiscount()
		profitRateInviter, _ = builder.ProfitRateInviter()

		incomeCellMinTransferValue, _ = builder.IncomeMinTransferCapacity()
		transferThrottle, _ = builder.TransferAccountThrottle()
		editManagerThrottle, _ = builder.EditManagerThrottle()
		editRecordsThrottle, _ = builder.EditRecordsThrottle()
		maxAccountLen, _ = builder.MaxLength()
		minTtl, _ = builder.RecordMinTtl()
		accountExpirationGracePeriod, _ = builder.ExpirationGracePeriod()
	}
	//
	resp.ReverseRecordCapacity = preparedFee + baseCapacity
	resp.MinChangeCapacity = common.DasLockWithBalanceTypeMinCkbCapacity
	//
	resp.SaleCellCapacity = saleCellBasicCapacity + saleCellPreparedFeeCapacity
	resp.MinSellPrice = saleMinPrice
	//
	decInviteDiscount := decimal.NewFromFloat(float64(inviteDiscount) / common.PercentRateBase)
	decProfitRateInviter := decimal.NewFromFloat(float64(profitRateInviter) / common.PercentRateBase)
	//
	resp.IncomeCellMinTransferValue = incomeCellMinTransferValue
	resp.TransferThrottle = transferThrottle
	resp.EditManagerThrottle = editManagerThrottle
	resp.EditRecordsThrottle = editRecordsThrottle
	resp.MaxAccountLen = maxAccountLen
	resp.MinTtl = minTtl
	resp.AccountExpirationGracePeriod = accountExpirationGracePeriod

	resp.Premium = config.Cfg.Das.Premium
	resp.MinAccountLen = uint32(config.Cfg.Das.AccountMinLength)
	resp.InviterDiscount = decInviteDiscount
	resp.ProfitRateOfInviter = decProfitRateInviter

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	resp.TimestampOnChain = timeCell.Timestamp()

	apiResp.ApiRespOK(resp)
	return nil
}
