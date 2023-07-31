package handle

import (
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
	"das_register_server/unipay"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

type ReqOrderChange struct {
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Account   string           `json:"account"`

	PayChainType  common.ChainType  `json:"pay_chain_type"`
	PayTokenId    tables.PayTokenId `json:"pay_token_id"`
	PayAddress    string            `json:"pay_address"`
	PayType       tables.PayType    `json:"pay_type"`
	CoinType      string            `json:"coin_type"`
	CrossCoinType string            `json:"cross_coin_type"`

	ReqOrderRegisterBase
}

type RespOrderChange struct {
	OrderId         string            `json:"order_id"`
	TokenId         tables.PayTokenId `json:"token_id"`
	ReceiptAddress  string            `json:"receipt_address"`
	Amount          decimal.Decimal   `json:"amount"`
	ContractAddress string            `json:"contract_address"`
	ClientSecret    string            `json:"client_secret"`
}

func (h *HttpHandle) RpcOrderChange(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqOrderChange
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

	if err = h.doOrderChange(&req[0], apiResp); err != nil {
		log.Error("doOrderChange err:", err.Error())
	}
}

func (h *HttpHandle) OrderChange(ctx *gin.Context) {
	var (
		funcName = "OrderChange"
		clientIp = GetClientIp(ctx)
		req      ReqOrderChange
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

	if err = h.doOrderChange(&req, &apiResp); err != nil {
		log.Error("doOrderChange err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderChange(req *ReqOrderChange, apiResp *api_code.ApiResp) error {
	var resp RespOrderChange
	if req.Address == "" || req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	if yes := req.PayTokenId.IsTokenIdCkbInternal(); yes {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("pay token id [%s] invalid", req.PayTokenId))
		return nil
	}
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
	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if err := h.rc.RegisterLimitLockWithRedis(req.ChainType, req.Address, "change", req.Account, time.Second*10); err != nil {
		if err == cache.ErrDistributedLockPreemption {
			apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
			return nil
		}
	}
	//if exi := h.rc.RegisterLimitExist(req.ChainType, req.Address, req.Account, "2"); exi {
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

	// old Order
	oldOrderContent := h.oldOrderCheck(req, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	// new Older
	h.doNewOrder(req, apiResp, &resp, oldOrderContent)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	// cache
	//_ = h.rc.SetRegisterLimit(req.ChainType, req.Address, req.Account, "2", time.Second*30)
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) doNewOrder(req *ReqOrderChange, apiResp *api_code.ApiResp, resp *RespOrderChange, oldOrderContent *tables.TableOrderContent) {
	// pay amount
	hexAddress := core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	}
	args, err := h.dasCore.Daf().HexToArgs(hexAddress, hexAddress)
	if err != nil {
		log.Error("HexToArgs err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
		return
	}
	accLen := uint8(len(oldOrderContent.AccountCharStr))
	if tables.EndWithDotBitChar(oldOrderContent.AccountCharStr) {
		accLen -= 4
	}
	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(accLen, common.Bytes2Hex(args), req.Account, req.InviterAccount, req.RegisterYears, false, req.PayTokenId)
	if err != nil {
		log.Error("getOrderAmount err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return
	}
	if amountTotalUSD.Cmp(decimal.Zero) != 1 || amountTotalCKB.Cmp(decimal.Zero) != 1 || amountTotalPayToken.Cmp(decimal.Zero) != 1 {
		log.Error("order amount err:", amountTotalUSD, amountTotalCKB, amountTotalPayToken)
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return
	}
	//
	inviterAccountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.InviterAccount))
	if _, ok := config.Cfg.InviterWhitelist[inviterAccountId]; ok {
		req.ChannelAccount = req.InviterAccount
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	orderContent := tables.TableOrderContent{
		AccountCharStr: oldOrderContent.AccountCharStr,
		InviterAccount: req.InviterAccount,
		ChannelAccount: req.ChannelAccount,
		RegisterYears:  req.RegisterYears,
		AmountTotalUSD: amountTotalUSD,
		AmountTotalCKB: amountTotalCKB,
	}
	contentDataStr, err := json.Marshal(&orderContent)
	if err != nil {
		log.Error("json marshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json marshal fail")
		return
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
			apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("HexToNormal err: %s", err.Error()))
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
			PayStatus:         tables.TxStatusDefault,
			HedgeStatus:       tables.TxStatusDefault,
			PreRegisterStatus: tables.TxStatusDefault,
			RegisterStatus:    tables.RegisterStatusConfirmPayment,
			OrderStatus:       tables.OrderStatusDefault,
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
		log.Error("CreateOrder err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "create order fail")
		return
	}
	// notify
	go func() {
		notify.SendLarkOrderNotify(&notify.SendLarkOrderNotifyParam{
			Key:        config.Cfg.Notify.LarkRegisterKey,
			Action:     "change register order",
			Account:    order.Account,
			OrderId:    order.OrderId,
			ChainType:  order.ChainType,
			Address:    order.Address,
			PayTokenId: order.PayTokenId,
			Amount:     order.PayAmount,
		})
	}()
}

func (h *HttpHandle) oldOrderCheck(req *ReqOrderChange, apiResp *api_code.ApiResp) (oldOrderContent *tables.TableOrderContent) {
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	order, err := h.dbDao.GetLatestRegisterOrderBySelf(req.ChainType, req.Address, accountId)
	if err != nil {
		log.Error("GetLatestRegisterOrderBySelf err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order fail")
		return
	} else if order.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeOrderNotExist, "order not exist")
		return
	} else if order.RegisterStatus != tables.RegisterStatusConfirmPayment {
		apiResp.ApiRespErr(api_code.ApiCodeOrderPaid, "order paid")
		return
	} else if order.OrderStatus != tables.OrderStatusDefault {
		apiResp.ApiRespErr(api_code.ApiCodeOrderNotExist, "order closed")
		return
	}

	var contentData tables.TableOrderContent
	if err := json.Unmarshal([]byte(order.Content), &contentData); err != nil {
		log.Error("json.Unmarshal err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json.Unmarshal fail")
		return
	}

	//if strings.EqualFold(contentData.InviterAccount, req.InviterAccount) &&
	//	strings.EqualFold(contentData.ChannelAccount, req.ChannelAccount) &&
	//	req.RegisterYears == contentData.RegisterYears &&
	//	req.PayTokenId == order.PayTokenId {
	//	apiResp.ApiRespErr(api_code.ApiCodeSameOrderInfo, "order info not change")
	//	return
	//}
	oldOrderContent = &contentData
	return
}
