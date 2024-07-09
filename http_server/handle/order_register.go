package handle

import (
	"context"
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/http_server/compatible"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
	"das_register_server/timer"
	"das_register_server/unipay"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ReqOrderRegister struct {
	ReqAccountSearch
	ReqOrderRegisterBase
	core.ChainTypeAddress
	PayChainType  common.ChainType  `json:"pay_chain_type"`
	PayAddress    string            `json:"pay_address"`
	PayTokenId    tables.PayTokenId `json:"pay_token_id"`
	PayType       tables.PayType    `json:"pay_type"`
	CoinType      string            `json:"coin_type"`
	CrossCoinType string            `json:"cross_coin_type"`
	GiftCard      string            `json:"gift_card"`
}

type ReqCheckCoupon struct {
	Coupon string `json:"coupon"`
}
type RespCouponInfo struct {
	CouponType   tables.CouponType   `json:"type"`
	CouponStatus tables.CouponStatus `json:"status"`
}
type ReqOrderRegisterBase struct {
	RegisterYears  int    `json:"register_years"`
	InviterAccount string `json:"inviter_account"`
	ChannelAccount string `json:"channel_account"`
}

type RespOrderRegister struct {
	OrderId         string            `json:"order_id"`
	TokenId         tables.PayTokenId `json:"token_id"`
	ReceiptAddress  string            `json:"receipt_address"`
	Amount          decimal.Decimal   `json:"amount"`
	ContractAddress string            `json:"contract_address"`
	ClientSecret    string            `json:"client_secret"`
}

type AccountAttr struct {
	Length uint8           `json:"length"`
	Amount decimal.Decimal `json:"amount"`
}

func (h *HttpHandle) RpcCheckCouponr(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqCheckCoupon
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

	if err = h.doCheckCoupon(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doOrderRegister err:", err.Error())
	}
}
func (h *HttpHandle) CheckCoupon(ctx *gin.Context) {
	var (
		funcName = "CheckCoupon"
		clientIp = GetClientIp(ctx)
		req      ReqCheckCoupon
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

	if err = h.doCheckCoupon(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doCheckCoupon err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCheckCoupon(ctx context.Context, req *ReqCheckCoupon, apiResp *api_code.ApiResp) error {
	if req.Coupon == "" {
		apiResp.ApiRespErr(api_code.ApiCodeCouponInvalid, "params invalid")
		return nil
	}
	err, respResult := h.getCouponInfo(ctx, req.Coupon)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "getCouponInfo err")
		return err
	}
	apiResp.ApiRespOK(respResult)
	return nil
}

func (h *HttpHandle) RpcOrderRegister(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqOrderRegister
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

	if err = h.doOrderRegister(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doOrderRegister err:", err.Error())
	}
}

func (h *HttpHandle) OrderRegister(ctx *gin.Context) {
	var (
		funcName = "OrderRegister"
		clientIp = GetClientIp(ctx)
		req      ReqOrderRegister
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

	if err = h.doOrderRegister(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doOrderRegister err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderRegister(ctx context.Context, req *ReqOrderRegister, apiResp *api_code.ApiResp) error {
	var resp RespOrderRegister

	req.CrossCoinType = "" // closed cross nft
	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	if yes := req.PayTokenId.IsTokenIdCkbInternal(); yes {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("pay token id [%s] invalid", req.PayTokenId))
		return nil
	}

	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid")
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	if !checkChainType(req.ChainType) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("chain type [%d] invalid", req.ChainType))
		return nil
	}

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if err := h.rc.RegisterLimitLockWithRedis(req.ChainType, req.Address, "register", req.Account, time.Second*10); err != nil {
		if err == cache.ErrDistributedLockPreemption {
			apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
			return nil
		}
	}

	// check un pay
	maxUnPayCount := int64(200)
	if config.Cfg.Server.Net != common.DasNetTypeMainNet {
		maxUnPayCount = 200
	}
	if unPayCount, err := h.dbDao.GetUnPayOrderCount(req.ChainType, req.Address); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "failed to check order count")
		return nil
	} else if unPayCount > maxUnPayCount {
		log.Info(ctx, "GetUnPayOrderCount:", req.ChainType, req.Address, unPayCount)
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
		return nil
	}

	//if exi := h.rc.RegisterLimitExist(req.ChainType, req.Address, req.Account, "1"); exi {
	//	apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
	//	return fmt.Errorf("AccountActionLimitExist: %d %s %s", req.ChainType, req.Address, req.Account)
	//}

	// order check
	if err := h.checkOrderInfo(req.CoinType, req.CrossCoinType, &req.ReqOrderRegisterBase, apiResp); err != nil {
		return fmt.Errorf("checkOrderInfo err: %s", err.Error())
	}
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	// account check
	h.checkAccountCharSet(&req.ReqAccountSearch, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}
	// base check
	_, status, _, _ := h.checkAccountBase(ctx, &req.ReqAccountSearch, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}
	if status != tables.SearchStatusRegisterAble {
		switch status {
		case tables.SearchStatusUnAvailableAccount:
			apiResp.ApiRespErr(api_code.ApiCodeUnAvailableAccount, "unavailable account")
		case tables.SearchStatusReservedAccount:
			apiResp.ApiRespErr(api_code.ApiCodeReservedAccount, "reserved account")
		case tables.SearchStatusRegisterNotOpen:
			apiResp.ApiRespErr(api_code.ApiCodeNotOpenForRegistration, "registration is not open")
		default:
			apiResp.ApiRespErr(api_code.ApiCodeAccountAlreadyRegister, "account already register")
		}
		return nil
	}
	// self order
	status, _ = h.checkAddressOrder(ctx, &req.ReqAccountSearch, apiResp, false)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if status != tables.SearchStatusRegisterAble {
		apiResp.ApiRespErr(api_code.ApiCodeAccountAlreadyRegister, "account registering")
		return nil
	}
	// registering check
	status = h.checkOtherAddressOrder(ctx, &req.ReqAccountSearch, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if status >= tables.SearchStatusRegistering {
		apiResp.ApiRespErr(api_code.ApiCodeAccountAlreadyRegister, "account registering")
		return nil
	}

	// create order
	if req.GiftCard != "" {
		h.doRegisterCouponOrder(ctx, req, apiResp, &resp)
	} else {
		h.doRegisterOrder(ctx, req, apiResp, &resp)
	}

	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}
	// cache
	// _ = h.rc.SetRegisterLimit(req.ChainType, req.Address, req.Account, "1", time.Second*30)
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) checkOrderInfo(coinType, crossCoinType string, req *ReqOrderRegisterBase, apiResp *api_code.ApiResp) error {
	if req.RegisterYears <= 0 || req.RegisterYears > config.Cfg.Das.MaxRegisterYears {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("register years[%d] invalid", req.RegisterYears))
		return nil
	}
	if req.InviterAccount != "" {
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.InviterAccount))
		acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search inviter account fail")
			return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
		} else if acc.Id == 0 {
			apiResp.ApiRespErr(api_code.ApiCodeInviterAccountNotExist, "inviter account not exist")
			return nil
		} else if acc.Status == tables.AccountStatusOnCross {
			apiResp.ApiRespErr(api_code.ApiCodeOnCross, "account on cross")
			return nil
		} else if strings.EqualFold(acc.Owner, common.BlackHoleAddress) {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "inviter account owner is 0x0")
			return nil
		}
	}

	if req.ChannelAccount != "" {
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.ChannelAccount))
		acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search channel account fail")
			return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
		} else if acc.Id == 0 || acc.Status == tables.AccountStatusOnCross || acc.IsExpired() {
			//apiResp.ApiRespErr(api_code.ApiCodeChannelAccountNotExist, "channel account not exist")
			//return nil
			req.ChannelAccount = ""
		}
	}
	if coinType != "" {
		if ok, _ := regexp.MatchString("^(0|[1-9][0-9]*)$", coinType); !ok {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("CoinType [%s] is invalid", coinType))
			return nil
		}
	}
	if crossCoinType != "" {
		if crossCoinType != string(common.CoinTypeEth) {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("CrossCoinType [%s] is invalid", coinType))
			return nil
		}
		//if ok, _ := regexp.MatchString("^(0|[1-9][0-9]*)$", crossCoinType); !ok {
		//	apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("CrossCoinType [%s] is invalid", coinType))
		//	return nil
		//}
	}
	return nil
}

func (h *HttpHandle) doRegisterOrder(ctx context.Context, req *ReqOrderRegister, apiResp *api_code.ApiResp, resp *RespOrderRegister) {
	// pay amount
	addrHex := core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	}
	args, err := h.dasCore.Daf().HexToArgs(addrHex, addrHex)
	if err != nil {
		log.Error(ctx, "HexToArgs err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
		return
	}
	accLen := uint8(len(req.AccountCharStr))
	if tables.EndWithDotBitChar(req.AccountCharStr) {
		accLen -= 4
	}

	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(ctx, accLen, common.Bytes2Hex(args), req.Account, req.InviterAccount, req.RegisterYears, false, req.PayTokenId)
	if err != nil {
		log.Error(ctx, "getOrderAmount err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return
	}

	if amountTotalUSD.Cmp(decimal.Zero) != 1 || amountTotalCKB.Cmp(decimal.Zero) != 1 || amountTotalPayToken.Cmp(decimal.Zero) != 1 {
		log.Error(ctx, "order amount err:", amountTotalUSD, amountTotalCKB, amountTotalPayToken)
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return
	}

	inviterAccountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.InviterAccount))
	if _, ok := config.Cfg.InviterWhitelist[inviterAccountId]; ok {
		req.ChannelAccount = req.InviterAccount
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	orderContent := tables.TableOrderContent{
		AccountCharStr: req.AccountCharStr,
		InviterAccount: req.InviterAccount,
		ChannelAccount: req.ChannelAccount,
		RegisterYears:  req.RegisterYears,
		AmountTotalUSD: amountTotalUSD,
		AmountTotalCKB: amountTotalCKB,
	}

	contentDataStr, err := json.Marshal(&orderContent)
	if err != nil {
		log.Error(ctx, "json marshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json marshal fail")
		return
	}

	// check balance
	if req.PayTokenId == tables.TokenIdDas {
		dasLock, _, err := h.dasCore.Daf().HexToScript(addrHex)
		if err != nil {
			log.Error(ctx, "HexToArgs err: ", err.Error())
			apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
			return
		}

		fee := common.OneCkb
		needCapacity := amountTotalPayToken.BigInt().Uint64()
		_, _, err = h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
			DasCache:          h.dasCache,
			LockScript:        dasLock,
			CapacityNeed:      needCapacity + fee,
			CapacityForChange: common.DasLockWithBalanceTypeMinCkbCapacity,
			SearchOrder:       indexer.SearchOrderDesc,
		})
		if err != nil {
			checkBalanceErr(err, apiResp)
			return
		}
	}

	var order tables.TableDasOrderInfo
	var paymentInfo tables.TableDasOrderPayInfo
	// unipay
	if config.Cfg.Server.UniPayUrl != "" {
		addrNormal, err := h.dasCore.Daf().HexToNormal(core.DasAddressHex{
			DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
			AddressHex:     req.Address,
			AddressPayload: nil,
			IsMulti:        false,
			ChainType:      req.ChainType,
		})
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("HexToNormal err"))
			return
		}
		premiumPercentage := decimal.Zero
		premiumBase := decimal.Zero
		premiumAmount := decimal.Zero
		if req.PayTokenId == tables.TokenIdStripeUSD {
			premiumPercentage = config.Cfg.Stripe.PremiumPercentage
			premiumBase = config.Cfg.Stripe.PremiumBase
			premiumAmount = amountTotalPayToken
			amountTotalPayToken = amountTotalPayToken.Mul(premiumPercentage.Add(decimal.NewFromInt(1))).Add(premiumBase.Mul(decimal.NewFromInt(100)))
			amountTotalPayToken = decimal.NewFromInt(amountTotalPayToken.Ceil().IntPart())
			premiumAmount = amountTotalPayToken.Sub(premiumAmount)
		}
		res, err := unipay.CreateOrder(unipay.ReqOrderCreate{
			ChainTypeAddress: core.ChainTypeAddress{
				Type: "blockchain",
				KeyInfo: core.KeyInfo{
					CoinType: addrNormal.ChainType.ToDasAlgorithmId(true).ToCoinType(),
					Key:      addrNormal.AddressNormal,
				},
			},
			BusinessId:        unipay.BusinessIdDasRegisterSvr,
			Amount:            amountTotalPayToken,
			PayTokenId:        req.PayTokenId,
			PaymentAddress:    config.GetUnipayAddress(req.PayTokenId),
			PremiumPercentage: premiumPercentage,
			PremiumBase:       premiumBase,
			PremiumAmount:     premiumAmount,
			MetaData: map[string]string{
				"account":      req.Account,
				"algorithm_id": req.ChainType.ToString(),
				"address":      addrNormal.AddressNormal,
				"action":       "register",
			},
		})
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to create order by unipay")
			return
		}
		order = tables.TableDasOrderInfo{
			OrderType:         tables.OrderTypeSelf,
			OrderId:           res.OrderId,
			AccountId:         accountId,
			Account:           req.Account,
			Action:            common.DasActionApplyRegister,
			ChainType:         req.ChainType,
			Address:           req.Address,
			Timestamp:         time.Now().UnixNano() / 1e6,
			PayTokenId:        req.PayTokenId,
			PayType:           req.PayType,
			PayAmount:         amountTotalPayToken,
			Content:           string(contentDataStr),
			PayStatus:         tables.TxStatusDefault,
			HedgeStatus:       tables.TxStatusDefault,
			PreRegisterStatus: tables.TxStatusDefault,
			OrderStatus:       tables.OrderStatusDefault,
			RegisterStatus:    tables.RegisterStatusConfirmPayment,
			CoinType:          req.CoinType,
			CrossCoinType:     req.CrossCoinType,
			IsUniPay:          tables.IsUniPayTrue,
			PremiumPercentage: premiumPercentage,
			PremiumBase:       premiumBase,
			PremiumAmount:     premiumAmount,
		}
		if req.PayTokenId == tables.TokenIdStripeUSD && res.StripePaymentIntentId != "" {
			paymentInfo = tables.TableDasOrderPayInfo{
				Hash:      res.StripePaymentIntentId,
				OrderId:   res.OrderId,
				ChainType: order.ChainType,
				Address:   order.Address,
				Status:    tables.OrderTxStatusDefault,
				Timestamp: time.Now().UnixMilli(),
				AccountId: order.AccountId,
			}
		}
		resp.ContractAddress = res.ContractAddress
		resp.ClientSecret = res.ClientSecret
	} else {
		order = tables.TableDasOrderInfo{
			OrderType:         tables.OrderTypeSelf,
			OrderId:           "",
			AccountId:         accountId,
			Account:           req.Account,
			Action:            common.DasActionApplyRegister,
			ChainType:         req.ChainType,
			Address:           req.Address,
			Timestamp:         time.Now().UnixNano() / 1e6,
			PayTokenId:        req.PayTokenId,
			PayType:           req.PayType,
			PayAmount:         amountTotalPayToken,
			Content:           string(contentDataStr),
			PayStatus:         tables.TxStatusDefault,
			HedgeStatus:       tables.TxStatusDefault,
			PreRegisterStatus: tables.TxStatusDefault,
			OrderStatus:       tables.OrderStatusDefault,
			RegisterStatus:    tables.RegisterStatusConfirmPayment,
			CoinType:          req.CoinType,
			CrossCoinType:     req.CrossCoinType,
		}
		order.CreateOrderId()
	}

	resp.OrderId = order.OrderId
	resp.TokenId = req.PayTokenId
	resp.Amount = order.PayAmount

	addr := config.GetUnipayAddress(order.PayTokenId)
	if addr == "" {
		apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("not supported [%s]", order.PayTokenId))
		return
	} else {
		resp.ReceiptAddress = addr
	}

	if err := h.dbDao.CreateOrderWithPayment(order, paymentInfo); err != nil {
		log.Error(ctx, "CreateOrder err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "create order fail")
		return
	}

	// notify
	go func() {
		notify.SendLarkOrderNotify(&notify.SendLarkOrderNotifyParam{
			Key:        config.Cfg.Notify.LarkRegisterKey,
			Action:     "new register order",
			Account:    order.Account,
			OrderId:    order.OrderId,
			ChainType:  order.ChainType,
			Address:    order.Address,
			PayTokenId: order.PayTokenId,
			Amount:     order.PayAmount,
		})
	}()
	return
}

func (h *HttpHandle) doRegisterCouponOrder(ctx context.Context, req *ReqOrderRegister, apiResp *api_code.ApiResp, resp *RespOrderRegister) {
	req.PayTokenId = tables.TokenCoupon
	// pay amount
	addrHex := core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	}
	args, err := h.dasCore.Daf().HexToArgs(addrHex, addrHex)
	if err != nil {
		log.Error(ctx, "HexToArgs err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
		return
	}
	accLen := uint8(len(req.AccountCharStr))
	if tables.EndWithDotBitChar(req.AccountCharStr) {
		accLen -= 4
	}

	var coupon *tables.TableCoupon
	if req.RegisterYears != 1 {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}

	coupon = h.checkCoupon(ctx, req.GiftCard, apiResp)
	if coupon == nil {
		return
	}

	accountAttr := AccountAttr{
		Length: accLen,
	}
	if res := h.checkCouponType(accountAttr, coupon); !res {
		apiResp.ApiRespErr(api_code.ApiCodeCouponInvalid, "gift card type err")
		return
	}

	req.InviterAccount = ""
	req.ChannelAccount = ""

	if err := h.rc.GetCouponLockWithRedis(coupon.Code, time.Minute*10); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the gift card operation is too frequent")
		return
	}

	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(ctx, accLen, common.Bytes2Hex(args), req.Account, req.InviterAccount, req.RegisterYears, false, req.PayTokenId)
	if err != nil {
		log.Error(ctx, "getOrderAmount err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return
	}

	if amountTotalUSD.Cmp(decimal.Zero) != 0 || amountTotalCKB.Cmp(decimal.Zero) != 0 || amountTotalPayToken.Cmp(decimal.Zero) != 0 {
		log.Error(ctx, "order amount err:", amountTotalUSD, amountTotalCKB, amountTotalPayToken)
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	orderContent := tables.TableOrderContent{
		AccountCharStr: req.AccountCharStr,
		InviterAccount: req.InviterAccount,
		ChannelAccount: req.ChannelAccount,
		RegisterYears:  req.RegisterYears,
		AmountTotalUSD: amountTotalUSD,
		AmountTotalCKB: amountTotalCKB,
	}

	contentDataStr, err := json.Marshal(&orderContent)
	if err != nil {
		log.Error(ctx, "json marshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json marshal fail")
		return
	}

	order := tables.TableDasOrderInfo{
		Id:                0,
		OrderType:         tables.OrderTypeSelf,
		OrderId:           "",
		AccountId:         accountId,
		Account:           req.Account,
		Action:            common.DasActionApplyRegister,
		ChainType:         req.ChainType,
		Address:           req.Address,
		Timestamp:         time.Now().UnixNano() / 1e6,
		PayTokenId:        req.PayTokenId,
		PayType:           req.PayType,
		PayAmount:         amountTotalPayToken,
		Content:           string(contentDataStr),
		PayStatus:         tables.TxStatusSending,
		HedgeStatus:       tables.TxStatusDefault,
		PreRegisterStatus: tables.TxStatusDefault,
		OrderStatus:       tables.OrderStatusDefault,
		RegisterStatus:    tables.RegisterStatusApplyRegister,
		CoinType:          req.CoinType,
		CrossCoinType:     req.CrossCoinType,
	}
	order.CreateOrderId()

	resp.OrderId = order.OrderId
	resp.TokenId = req.PayTokenId
	//resp.PayType = req.PayType
	resp.Amount = order.PayAmount
	//resp.CodeUrl = ""

	err = h.dbDao.CreateCouponOrder(&order, coupon.Code)
	if redisErr := h.rc.DeleteCouponLockWithRedis(coupon.Code); redisErr != nil {
		log.Error(ctx, "delete coupon redis lock error : ", redisErr.Error())
	}
	if err != nil {
		log.Error(ctx, "CreateOrder err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "create order fail")
		return
	}

	// notify
	go func() {
		notify.SendLarkOrderNotify(&notify.SendLarkOrderNotifyParam{
			Key:        config.Cfg.Notify.LarkRegisterKey,
			Action:     "new register coupon order",
			Account:    order.Account,
			OrderId:    order.OrderId,
			ChainType:  order.ChainType,
			Address:    order.Address,
			PayTokenId: order.PayTokenId,
			Amount:     order.PayAmount,
		})
	}()
	return
}

func (h *HttpHandle) getOrderAmount(ctx context.Context, accLen uint8, args, account, inviterAccount string, years int, isRenew bool, payTokenId tables.PayTokenId) (amountTotalUSD decimal.Decimal, amountTotalCKB decimal.Decimal, amountTotalPayToken decimal.Decimal, e error) {
	// pay token
	if payTokenId == tables.TokenCoupon {
		amountTotalUSD = decimal.Zero
		amountTotalCKB = decimal.Zero
		amountTotalPayToken = decimal.Zero
		return
	}
	payToken := timer.GetTokenInfo(payTokenId)
	if payToken.TokenId == "" {
		e = fmt.Errorf("not supported [%s]", payTokenId)
		return
	}
	//
	quoteCell, err := h.dasCore.GetQuoteCell()
	if err != nil {
		e = fmt.Errorf("GetQuoteCell err: %s", err.Error())
		return
	}
	quote := quoteCell.Quote()
	decQuote := decimal.NewFromInt(int64(quote)).Div(decimal.NewFromInt(common.UsdRateBase))
	// base price
	baseAmount, accountPrice, err := h.getAccountPrice(ctx, accLen, args, account, isRenew)
	if err != nil {
		e = fmt.Errorf("getAccountPrice err: %s", err.Error())
		return
	}
	if isRenew {
		baseAmount = decimal.Zero
	}
	accountPrice = accountPrice.Mul(decimal.NewFromInt(int64(years)))
	if inviterAccount != "" {
		builder, err := h.dasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsPrice)
		if err != nil {
			e = fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
			return
		}
		discount, _ := builder.PriceInvitedDiscount()
		decDiscount := decimal.NewFromInt(int64(discount)).Div(decimal.NewFromInt(common.PercentRateBase))
		accountPrice = accountPrice.Mul(decimal.NewFromInt(1).Sub(decDiscount))
	}
	amountTotalUSD = accountPrice

	log.Info(ctx, "before Premium:", account, isRenew, amountTotalUSD, baseAmount, accountPrice)
	if config.Cfg.Das.Premium.Cmp(decimal.Zero) == 1 {
		amountTotalUSD = amountTotalUSD.Mul(config.Cfg.Das.Premium.Add(decimal.NewFromInt(1)))
	}
	if config.Cfg.Das.Discount.Cmp(decimal.Zero) == 1 {
		amountTotalUSD = amountTotalUSD.Mul(config.Cfg.Das.Discount)
	}
	amountTotalUSD = amountTotalUSD.Add(baseAmount)
	log.Info(ctx, "after Premium:", account, isRenew, amountTotalUSD, baseAmount, accountPrice)

	amountTotalUSD = amountTotalUSD.Mul(decimal.NewFromInt(100)).Ceil().DivRound(decimal.NewFromInt(100), 2)
	amountTotalCKB = amountTotalUSD.Div(decQuote).Mul(decimal.NewFromInt(int64(common.OneCkb))).Ceil()
	amountTotalPayToken = amountTotalUSD.Div(payToken.Price).Mul(decimal.New(1, payToken.Decimals)).Ceil()

	log.Info(ctx, "getOrderAmount:", amountTotalUSD, amountTotalCKB, amountTotalPayToken)
	if payToken.TokenId == tables.TokenIdCkb {
		amountTotalPayToken = amountTotalCKB
	}
	//if payToken.TokenId == tables.TokenIdMatic || payToken.TokenId == tables.TokenIdBnb || payToken.TokenId == tables.TokenIdEth {
	//	log.Info("amountTotalPayToken:", amountTotalPayToken.String())
	//	decCeil := decimal.NewFromInt(1e6)
	//	amountTotalPayToken = amountTotalPayToken.DivRound(decCeil, 6).Ceil().Mul(decCeil)
	//	log.Info("amountTotalPayToken:", amountTotalPayToken.String())
	//}
	if payToken.TokenId == tables.TokenIdDoge && h.dasCore.NetType() != common.DasNetTypeMainNet {
		amountTotalPayToken = decimal.NewFromInt(rand.Int63n(10000000) + 100000000)
	}
	amountTotalPayToken = unipay.RoundAmount(amountTotalPayToken, payToken.TokenId)
	return
}
func (h *HttpHandle) getCouponInfo(ctx context.Context, code string) (err error, info *RespCouponInfo) {
	info = new(RespCouponInfo)
	salt := config.Cfg.Server.CouponEncrySalt
	if salt == "" {
		log.Error(ctx, "GetCoupon err: config coupon_encry_salt is empty")
		return fmt.Errorf("system setting error"), info
	}
	code = couponEncry(code, salt)
	res, err := h.dbDao.GetCouponByCode(code)
	if err != nil {
		log.Error(ctx, "GetCoupon err:", err.Error())
		return fmt.Errorf("get gift card error"), info
	}
	if res.Id == 0 {
		info.CouponStatus = tables.CouponStatusNotfound
		return nil, info
	}
	info.CouponType = res.CouponType
	if res.OrderId != "" {
		info.CouponStatus = tables.CouponStatusUsed
		return nil, info
	}
	nowTime := time.Now().Unix()
	if nowTime < res.StartAt.Unix() || nowTime > res.ExpiredAt.Unix() {
		info.CouponStatus = tables.CouponStatusExpired
		return nil, info
	}

	info.CouponStatus = tables.CouponStatusAvailable
	return nil, info
}

func (h *HttpHandle) checkCoupon(ctx context.Context, code string, apiResp *api_code.ApiResp) (coupon *tables.TableCoupon) {
	salt := config.Cfg.Server.CouponEncrySalt
	if salt == "" {
		log.Error(ctx, "GetCoupon err: config coupon_encry_salt is empty")
		apiResp.ApiRespErr(api_code.ApiCodeError500, "system setting error")
		return nil
	}
	code = couponEncry(code, salt)
	res, err := h.dbDao.GetCouponByCode(code)
	if err != nil {
		log.Error(ctx, "GetCoupon err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get gift card error")
		return nil
	}
	if res.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeCouponInvalid, "gift card not found")
		return nil
	}
	if res.OrderId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeCouponUsed, "gift card has been used")
		return nil
	}
	nowTime := time.Now().Unix()
	if nowTime < res.StartAt.Unix() || nowTime > res.ExpiredAt.Unix() {
		apiResp.ApiRespErr(api_code.ApiCodeCouponUnopen, "gift card time has not arrived or expired")
		return nil
	}

	return &res
}

func (h *HttpHandle) checkCouponType(accountAttr AccountAttr, coupon *tables.TableCoupon) bool {
	if coupon.CouponType == tables.CouponType4byte && accountAttr.Length == 4 {
		return true
	}
	if coupon.CouponType == tables.CouponType5byte && accountAttr.Length >= 5 {
		return true
	}
	return false
}
