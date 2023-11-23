package handle

import (
	"das_register_server/config"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqGetPendingAuctionOrder struct {
	core.ChainTypeAddress
	address   string
	chainType common.ChainType
}

func (h *HttpHandle) GetPendingAuctionOrder(ctx *gin.Context) {
	var (
		funcName = "GetPendingAuctionOrder"
		clientIp = GetClientIp(ctx)
		req      ReqGetPendingAuctionOrder
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}

	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doGetPendingAuctionOrder(&req, &apiResp); err != nil {
		log.Error("doGetPendingAuctionOrder err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetPendingAuctionOrder(req *ReqGetPendingAuctionOrder, apiResp *http_api.ApiResp) (err error) {
	resp := make([]RepReqGetAuctionOrder, 0)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.address, req.chainType = addrHex.AddressHex, addrHex.ChainType
	list, err := h.dbDao.GetPendingAuctionOrder(addrHex.ChainType, addrHex.AddressHex)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "db error")
		return
	}
	for _, v := range list {
		hash, _ := common.String2OutPoint(v.Outpoint)
		resp = append(resp, RepReqGetAuctionOrder{
			Account:      v.Account,
			PremiumPrice: v.PremiumPrice,
			BasicPrice:   v.BasicPrice,
			Hash:         hash,
			Status:       v.Status,
		})
	}
	apiResp.ApiRespOK(resp)
	return
}
