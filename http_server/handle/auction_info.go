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
	"time"
)

type ReqAccountAuctionInfo struct {
	Account string `json:"account"  binding:"required"`
	core.ChainTypeAddress
	address   string
	chainType common.ChainType
}

type RespAccountAuctionInfo struct {
	AccountId     string           `json:"account_id"`
	Account       string           `json:"account"`
	BidStatus     tables.BidStatus `json:"bid_status"`
	Hash          string           `json:"hash"`
	StartsaleTime uint64           `json:"start_auction_time"`
	EndSaleTime   uint64           `json:"end_auction_time"`
	ExipiredTime  uint64           `json:"expired_at"`
	AccountPrice  decimal.Decimal  `json:"account_price"`
	BaseAmount    decimal.Decimal  `json:"base_amount"`
}

func (h *HttpHandle) GetAccountAuctionInfo(ctx *gin.Context) {
	var (
		funcName = "GetAccountAuctionInfo"
		clientIp = GetClientIp(ctx)
		req      ReqAccountAuctionInfo
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

	if err = h.doGetAccountAuctionInfo(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("GetAccountAuctionInfo err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetAccountAuctionInfo(ctx context.Context, req *ReqAccountAuctionInfo, apiResp *http_api.ApiResp) (err error) {
	var resp RespAccountAuctionInfo
	var addrHex *core.DasAddressHex
	if req.KeyInfo.Key != "" {
		addrHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
			return nil
		}
		req.address, req.chainType = addrHex.AddressHex, addrHex.ChainType
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	if acc.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, fmt.Sprintf("account [%s] not exist", req.Account))
		return
	}

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "GetTimeCell err")
		return fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	nowTime := timeCell.Timestamp()
	if status, _, err := h.checkDutchAuction(ctx, acc.ExpiredAt, uint64(nowTime)); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "checkDutchAuction err")
		return fmt.Errorf("checkDutchAuction err: %s", err.Error())
	} else if status != tables.SearchStatusOnDutchAuction {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return nil
	}

	//search bid status of a account
	createTime := time.Now().Unix() - 30*86400
	list, err := h.dbDao.GetAuctionOrderByAccount(req.Account, createTime)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "db error")
		return
	}

	if addrHex != nil {
		if len(list) == 0 {
			resp.BidStatus = tables.BidStatusNoOne
		} else {
			resp.BidStatus = tables.BidStatusByOthers
			for _, v := range list {
				if v.ChainType == addrHex.ChainType && v.Address == addrHex.AddressHex {
					resp.BidStatus = tables.BidStatusByMe
					resp.Hash, _ = common.String2OutPoint(v.Outpoint)
				}
			}
			apiResp.ApiRespOK(resp)
			return
		}
	}

	_, accLen, err := common.GetDotBitAccountLength(req.Account)
	if err != nil {
		return
	}
	if accLen == 0 {
		err = fmt.Errorf("accLen is 0")
		return
	}
	baseAmount, accountPrice, err := h.getAccountPrice(ctx, uint8(accLen), "", req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	auctionConfig, err := h.GetAuctionConfig(h.dasCore)
	if err != nil {
		err = fmt.Errorf("GetAuctionConfig err: %s", err.Error())
		return
	}
	gracePeriodTime := auctionConfig.GracePeriodTime
	auctionPeriodTime := auctionConfig.AuctionPeriodTime

	resp.AccountId = acc.AccountId
	resp.Account = req.Account
	resp.StartsaleTime = acc.ExpiredAt + uint64(gracePeriodTime)
	resp.EndSaleTime = acc.ExpiredAt + uint64(gracePeriodTime+auctionPeriodTime)
	resp.AccountPrice = accountPrice
	resp.BaseAmount = baseAmount
	resp.ExipiredTime = acc.ExpiredAt
	apiResp.ApiRespOK(resp)
	return
}
