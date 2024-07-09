package handle

import (
	"context"
	"das_register_server/config"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
)

type ReqAuctionOrderStatus struct {
	Hash string `json:"hash" binding:"required"`
	core.ChainTypeAddress
	address   string
	chainType common.ChainType
}
type RepReqGetAuctionOrder struct {
	Account      string          `json:"account"`
	Hash         string          `json:"hash"`
	Status       int             `json:"status"`
	BasicPrice   decimal.Decimal `json:"basic_price" gorm:"column:basic_price;type:decimal(60,0) NOT NULL DEFAULT '0' COMMENT ''"`
	PremiumPrice decimal.Decimal `json:"premium_price" gorm:"column:premium_price;type:decimal(60,0) NOT NULL DEFAULT '0' COMMENT ''"`
}

func (h *HttpHandle) GetAuctionOrderStatus(ctx *gin.Context) {
	var (
		funcName = "GetAuctionOrderStatus"
		clientIp = GetClientIp(ctx)
		req      ReqAuctionOrderStatus
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

	if err = h.doGetAuctionOrderStatus(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doGetAuctionOrderStatus err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetAuctionOrderStatus(ctx context.Context, req *ReqAuctionOrderStatus, apiResp *http_api.ApiResp) (err error) {
	var resp RepReqGetAuctionOrder

	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.address, req.chainType = addrHex.AddressHex, addrHex.ChainType
	order, err := h.dbDao.GetAuctionOrderStatus(addrHex.ChainType, addrHex.AddressHex, req.Hash)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "db error")
		return
	}
	if order.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionOrderNotFound, "order not found")
		return
	}

	resp.Account = order.Account
	resp.PremiumPrice = order.PremiumPrice
	resp.BasicPrice = order.BasicPrice
	resp.Hash, _ = common.String2OutPoint(order.Outpoint)
	resp.Status = order.Status
	apiResp.ApiRespOK(resp)
	return
}
