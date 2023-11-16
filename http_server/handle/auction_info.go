package handle

import (
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

type ReqAuctionPrice struct {
	Account string `json:"account"`
}

type RespAuctionPrice struct {
	//BasicPrice   decimal.Decimal `json:"basic_price"`
	AccountPrice decimal.Decimal `json:"account_price"`
	BaseAmount   decimal.Decimal `json:"base_amount"`
	PremiumPrice int64           `json:"premium_price"`
}

//查询价格
func (h *HttpHandle) GetAccountAuctionPrice(ctx *gin.Context) {
	var (
		funcName = "GetAccountAuctionPrice"
		clientIp = GetClientIp(ctx)
		req      ReqAuctionPrice
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

	if err = h.doGetAccountAuctionPrice(&req, &apiResp); err != nil {
		log.Error("GetAccountAuctionInfo err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doGetAccountAuctionPrice(req *ReqAuctionPrice, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuctionPrice
	if req.Account == "" {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid: account is empty")
		return
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	nowTime := uint64(time.Now().Unix())
	//exp + 90 + 27 +3
	//now > exp+117 exp< now - 117
	//now< exp+90 exp>now -90
	if acc.ExpiredAt > nowTime-90*24*3600 || acc.ExpiredAt < nowTime-117*24*3600 {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return
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
	baseAmount, accountPrice, err := h.getAccountPrice(uint8(accLen), "", req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	resp.BaseAmount = baseAmount
	resp.AccountPrice = accountPrice
	resp.PremiumPrice = common.Premium(int64(acc.ExpiredAt), int64(nowTime))
	apiResp.ApiRespOK(resp)
	return
}

type ReqAccountAuctionInfo struct {
	Account string `json:"account"`
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doGetAccountAuctionInfo(&req, &apiResp); err != nil {
		log.Error("GetAccountAuctionInfo err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetAccountAuctionInfo(req *ReqAccountAuctionInfo, apiResp *http_api.ApiResp) (err error) {
	var resp RespAccountAuctionInfo
	if req.Account == "" {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid: account is empty")
		return
	}
	var addrHex *core.DasAddressHex
	if req.KeyInfo.Key != "" {
		addrHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
			return nil
		}
		fmt.Println(addrHex.DasAlgorithmId, addrHex.DasSubAlgorithmId, addrHex.AddressHex)
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

	if status, _, err := h.checkDutchAuction(acc.ExpiredAt); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "checkDutchAuction err")
		return fmt.Errorf("checkDutchAuction err: %s", err.Error())
	} else if status != tables.SearchStatusOnDutchAuction {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return nil
	}
	//if acc.ExpiredAt > nowTime-90*24*3600 || acc.ExpiredAt < nowTime-117*24*3600 {
	//	apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
	//	return
	//}

	//查询账号的竞拍状态
	list, err := h.dbDao.GetAuctionOrderByAccount(req.Account)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "db error")
		return
	}

	//无人竞拍（有发送中和上链成功的订单）
	if addrHex != nil {
		if len(list) == 0 {
			resp.BidStatus = tables.BidStatusNoOne
		} else {
			resp.BidStatus = tables.BidStatusByOthers //被其他人竞拍
			for _, v := range list {
				//被自己竞拍
				if v.AlgorithmId == addrHex.DasAlgorithmId && v.SubAlgorithmId == addrHex.DasSubAlgorithmId && v.Address == addrHex.AddressHex {
					resp.BidStatus = tables.BidStatusByMe
					resp.Hash, _ = common.String2OutPoint(v.Outpoint)
				}
			}
			apiResp.ApiRespOK(resp)
			return
		}
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
	baseAmount, accountPrice, err := h.getAccountPrice(uint8(accLen), "", req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	resp.AccountId = acc.AccountId
	resp.Account = req.Account
	resp.StartsaleTime = acc.ExpiredAt + 90*86400
	resp.EndSaleTime = acc.ExpiredAt + 117*86400
	resp.AccountPrice = accountPrice
	resp.BaseAmount = baseAmount
	resp.ExipiredTime = acc.ExpiredAt
	apiResp.ApiRespOK(resp)
	return
}

type ReqGetAuctionOrder struct {
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
		funcName = "GetAuctionOrder"
		clientIp = GetClientIp(ctx)
		req      ReqGetAuctionOrder
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

	if err = h.doGetAuctionOrderStatus(&req, &apiResp); err != nil {
		log.Error("GetBidStatus err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetAuctionOrderStatus(req *ReqGetAuctionOrder, apiResp *http_api.ApiResp) (err error) {
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

type ReqGetGetPendingAuctionOrder struct {
	core.ChainTypeAddress
	address   string
	chainType common.ChainType
}

func (h *HttpHandle) GetPendingAuctionOrder(ctx *gin.Context) {
	var (
		funcName = "GetPendingAuctionOrder"
		clientIp = GetClientIp(ctx)
		req      ReqGetGetPendingAuctionOrder
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
		log.Error("GetBidStatus err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doGetPendingAuctionOrder(req *ReqGetGetPendingAuctionOrder, apiResp *http_api.ApiResp) (err error) {
	resp := make([]RepReqGetAuctionOrder, 0)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.address, req.chainType = addrHex.AddressHex, addrHex.ChainType
	if req.chainType != 1 {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
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
