package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
)

type ReqAuctionPrice struct {
	Account               string `json:"account"  binding:"required"`
	core.ChainTypeAddress        // ccc address
	addressHex            *core.DasAddressHex
}

type RespAuctionPrice struct {
	//BasicPrice   decimal.Decimal `json:"basic_price"`
	AccountPrice decimal.Decimal `json:"account_price"`
	DidCellPrice decimal.Decimal `json:"did_cell_amount"`
	BaseAmount   decimal.Decimal `json:"base_amount"`
	PremiumPrice decimal.Decimal `json:"premium_price"`
}

// 查询价格
func (h *HttpHandle) GetAccountAuctionPrice(ctx *gin.Context) {
	var (
		funcName = "GetAccountAuctionPrice"
		clientIp = GetClientIp(ctx)
		req      ReqAuctionPrice
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

	if err = h.doGetAccountAuctionPrice(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doGetAccountAuctionPrice err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}
	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doGetAccountAuctionPrice(ctx context.Context, req *ReqAuctionPrice, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuctionPrice
	req.addressHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	//nowTime := uint64(time.Now().Unix())
	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "GetTimeCell err")

		return fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	nowTime := timeCell.Timestamp()
	//exp + 90 + 27 +3
	//now > exp+117 exp< now - 117
	//now< exp+90 exp>now -90
	if status, _, err := h.checkDutchAuction(ctx, acc.ExpiredAt, uint64(nowTime)); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "checkDutchAuction err")
		return fmt.Errorf("checkDutchAuction err: %s", err.Error())
	} else if status != tables.SearchStatusOnDutchAuction {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return nil
	}

	//计算长度
	_, accLen, err := common.GetDotBitAccountLength(req.Account)
	if err != nil {
		return
	}
	if accLen == 0 {
		err = fmt.Errorf("accLen is 0")
		return
	}
	didCellAmount, baseAmount, accountPrice, err := h.getAccountPrice(ctx, req.addressHex, uint8(accLen), req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	auctionConfig, err := h.GetAuctionConfig(h.dasCore)
	if err != nil {
		err = fmt.Errorf("GetAuctionConfig err: %s", err.Error())
		return
	}

	resp.BaseAmount = baseAmount
	resp.AccountPrice = accountPrice
	resp.DidCellPrice = didCellAmount
	resp.PremiumPrice = decimal.NewFromFloat(common.Premium(int64(acc.ExpiredAt+uint64(auctionConfig.GracePeriodTime)), int64(nowTime)))
	apiResp.ApiRespOK(resp)
	return
}
