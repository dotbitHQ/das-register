package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
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
	if err := h.doConfigInfo(apiResp); err != nil {
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

	log.Info("ApiReq:", funcName, clientIp)

	if err = h.doConfigInfo(&apiResp); err != nil {
		log.Error("doConfigInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doConfigInfo(apiResp *api_code.ApiResp) error {
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
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
	}
	//
	preparedFee, _ := builder.RecordPreparedFeeCapacity()
	baseCapacity, _ := builder.RecordBasicCapacity()

	resp.ReverseRecordCapacity = preparedFee + baseCapacity
	resp.MinChangeCapacity = common.DasLockWithBalanceTypeOccupiedCkb
	//
	saleCellBasicCapacity, _ := builder.SaleCellBasicCapacity()
	saleCellPreparedFeeCapacity, _ := builder.SaleCellPreparedFeeCapacity()
	resp.SaleCellCapacity = saleCellBasicCapacity + saleCellPreparedFeeCapacity
	resp.MinSellPrice, _ = builder.SaleMinPrice()
	//
	inviteDiscount, _ := builder.PriceInvitedDiscount()
	decInviteDiscount := decimal.NewFromFloat(float64(inviteDiscount) / common.PercentRateBase)

	profitRateInviter, _ := builder.ProfitRateInviter()
	decProfitRateInviter := decimal.NewFromFloat(float64(profitRateInviter) / common.PercentRateBase)
	//
	resp.Premium = config.Cfg.Das.Premium
	resp.IncomeCellMinTransferValue, _ = builder.IncomeMinTransferCapacity()
	resp.TransferThrottle, _ = builder.TransferAccountThrottle()
	resp.EditManagerThrottle, _ = builder.EditManagerThrottle()
	resp.EditRecordsThrottle, _ = builder.EditRecordsThrottle()
	resp.MaxAccountLen, _ = builder.MaxLength()
	resp.MinAccountLen = uint32(config.Cfg.Das.AccountMinLength)
	resp.InviterDiscount = decInviteDiscount
	resp.ProfitRateOfInviter = decProfitRateInviter
	resp.MinTtl, _ = builder.RecordMinTtl()
	resp.AccountExpirationGracePeriod, _ = builder.ExpirationGracePeriod()

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	resp.TimestampOnChain = timeCell.Timestamp()

	apiResp.ApiRespOK(resp)
	return nil
}
