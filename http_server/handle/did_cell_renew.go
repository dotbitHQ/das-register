package handle

import (
	"context"
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
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type ReqDidCellRenew struct {
	core.ChainTypeAddress
	Account    string            `json:"account"`
	PayTokenId tables.PayTokenId `json:"pay_token_id"`
	RenewYears int               `json:"renew_years"`
}

type RespDidCellRenew struct {
	OrderId         string          `json:"order_id"`
	ReceiptAddress  string          `json:"receipt_address"`
	Amount          decimal.Decimal `json:"amount"`
	ContractAddress string          `json:"contract_address"`
	ClientSecret    string          `json:"client_secret"`
	SignInfo
}

func (h *HttpHandle) RpcDidCellRenew(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellRenew
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

	if err = h.doDidCellRenew(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellRenew err:", err.Error())
	}
}

func (h *HttpHandle) DidCellRenew(ctx *gin.Context) {
	var (
		funcName = "DidCellRenew"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellRenew
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

	if err = h.doDidCellRenew(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellRenew err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellRenew(ctx context.Context, req *ReqDidCellRenew, apiResp *http_api.ApiResp) error {
	var resp RespDidCellRenew

	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "FormatChainTypeAddress err")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	}

	//switch addrHex.DasAlgorithmId {
	//case common.DasAlgorithmIdAnyLock, common.DasAlgorithmIdBitcoin:
	//	apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address invalid")
	//	return nil
	//}

	if req.Account == "" || !strings.HasSuffix(req.Account, common.DasAccountSuffix) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	req.Account = strings.ToLower(req.Account)
	if yes := req.PayTokenId.IsTokenIdCkbInternal(); yes {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("pay token id [%s] invalid", req.PayTokenId))
		return nil
	}
	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	// check renew info
	if req.RenewYears < 1 || req.RenewYears > config.Cfg.Das.MaxRegisterYears {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("renew years[%d] invalid", req.RenewYears))
		return nil
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		log.Error(ctx, "GetAccountInfoByAccountId err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account fail")
		return nil
	} else if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "not support sub account")
		return nil
	} else if acc.Status == tables.AccountStatusOnCross {
		apiResp.ApiRespErr(api_code.ApiCodeOnCross, "account on cross")
		return nil
	}
	builder, err := h.dasCore.ConfigCellDataBuilderByTypeArgsList(
		common.ConfigCellTypeArgsAccount,
	)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error()))
		return nil
	}
	expirationGracePeriod, _ := builder.ExpirationGracePeriod()
	log.Info(ctx, "checkRenewOrder:", expirationGracePeriod)
	if int64(acc.ExpiredAt+uint64(expirationGracePeriod)) <= time.Now().Unix() {
		apiResp.ApiRespErr(api_code.ApiCodeAfterGracePeriod, "after the grace period")
		return nil
	}

	// did cell
	var txParams *txbuilder.BuildTransactionParams
	if acc.Status == tables.AccountStatusOnUpgrade {
		didAcc, err := h.dbDao.GetDidAccountByAccountIdWithoutArgs(accountId)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "Failed to get did account")
			return nil
		} else if didAcc.Id == 0 {
			apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "did account not exist")
			return nil
		}
		parseSvrAddr, err := address.Parse(config.Cfg.Server.PayServerAddress)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("address.Parse err: %s", err.Error())
		}
		txParams, err = txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
			DasCore:             h.dasCore,
			DasCache:            h.dasCache,
			Action:              common.DidCellActionRenew,
			DidCellOutPoint:     didAcc.GetOutpoint(),
			AccountCellOutPoint: acc.GetOutpoint(),
			EditRecords:         nil,
			EditOwnerLock:       nil,
			RenewYears:          req.RenewYears,
			NormalCellScript:    parseSvrAddr.Script,
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build tx")
			return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
		}
	}

	// renew price
	accOutpoint := common.String2OutPointStruct(acc.Outpoint)
	accTx, err := h.dasCore.Client().GetTransaction(h.ctx, accOutpoint.TxHash)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	mapAcc, err := witness.AccountIdCellDataBuilderFromTx(accTx.Transaction, common.DataTypeNew)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}
	accBuilder, ok := mapAcc[acc.AccountId]
	if !ok {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "mapAcc is nil")
		return fmt.Errorf("mapAcc is nil")
	}

	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(ctx, uint8(accBuilder.AccountChars.Len()), "", req.Account, "", req.RenewYears, true, req.PayTokenId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return fmt.Errorf("getOrderAmount err: %s", err.Error())
	}
	if amountTotalUSD.Cmp(decimal.Zero) != 1 || amountTotalCKB.Cmp(decimal.Zero) != 1 || amountTotalPayToken.Cmp(decimal.Zero) != 1 {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get order amount fail")
		return nil
	}

	// order
	orderContent := tables.TableOrderContent{
		AmountTotalUSD: amountTotalUSD,
		AmountTotalCKB: amountTotalCKB,
		RenewYears:     req.RenewYears,
	}
	contentDataStr, err := json.Marshal(&orderContent)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json marshal fail")
		return fmt.Errorf("json.Marshal err: %s", err.Error())
	}
	if req.PayTokenId == tables.TokenIdDas {
		if addrHex.DasAlgorithmId == common.DasAlgorithmIdAnyLock {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address invalid")
			return nil
		}
		dasLock, _, err := h.dasCore.Daf().HexToScript(*addrHex)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
			return fmt.Errorf("HexToScript err: %s", err.Error())
		}

		needCapacity := amountTotalPayToken.BigInt().Uint64()
		_, _, err = h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
			DasCache:          h.dasCache,
			LockScript:        dasLock,
			CapacityNeed:      needCapacity + common.OneCkb,
			CapacityForChange: common.DasLockWithBalanceTypeMinCkbCapacity,
			SearchOrder:       indexer.SearchOrderDesc,
		})
		if err != nil {
			checkBalanceErr(err, apiResp)
			return nil
		}
	}

	// unipay
	var order tables.TableDasOrderInfo
	var paymentInfo tables.TableDasOrderPayInfo

	if config.Cfg.Server.UniPayUrl == "" {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "UniPayUrl is nil")
		return fmt.Errorf("UniPayUrl is nil")
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
		ChainTypeAddress:  req.ChainTypeAddress,
		BusinessId:        unipay.BusinessIdDasRegisterSvr,
		Amount:            amountTotalPayToken,
		PayTokenId:        req.PayTokenId,
		PaymentAddress:    config.GetUnipayAddress(req.PayTokenId),
		PremiumPercentage: premiumPercentage,
		PremiumBase:       premiumBase,
		PremiumAmount:     premiumAmount,
		MetaData: map[string]string{
			"account":      req.Account,
			"algorithm_id": addrHex.ChainType.ToString(),
			"address":      req.KeyInfo.Key,
			"action":       "renew",
		},
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to create order by unipay")
		return nil
	}
	order = tables.TableDasOrderInfo{
		OrderType:         tables.OrderTypeSelf,
		OrderId:           res.OrderId,
		AccountId:         accountId,
		Account:           req.Account,
		Action:            common.DasActionRenewAccount,
		ChainType:         addrHex.ChainType,
		Address:           req.KeyInfo.Key,
		Timestamp:         time.Now().UnixMilli(),
		PayTokenId:        req.PayTokenId,
		PayAmount:         amountTotalPayToken,
		Content:           string(contentDataStr),
		PayStatus:         tables.TxStatusDefault,
		HedgeStatus:       tables.TxStatusDefault,
		PreRegisterStatus: tables.TxStatusDefault,
		RegisterStatus:    tables.RegisterStatusDefault,
		OrderStatus:       tables.OrderStatusDefault,
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

	resp.OrderId = order.OrderId
	resp.Amount = order.PayAmount
	addr := config.GetUnipayAddress(order.PayTokenId)
	if addr == "" {
		apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("not supported [%s]", order.PayTokenId))
		return nil
	} else {
		resp.ReceiptAddress = addr
	}

	//
	if txParams != nil {
		reqBuild := reqBuildTx{
			OrderId:    order.OrderId,
			Action:     common.DasActionRenewAccount,
			ChainType:  0,
			Address:    req.KeyInfo.Key,
			Account:    req.Account,
			EvmChainId: req.GetChainId(config.Cfg.Server.Net),
		}
		if didCellTx, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
			checkBuildTxErr(err, apiResp)
			return fmt.Errorf("buildTx: %s", err.Error())
		} else {
			resp.SignInfo = *si
			resp.SignInfo.CKBTx = didCellTx
		}
		order.IsDidCell = tables.IsDidCellYes
	}

	if err := h.dbDao.CreateOrderWithPayment(order, paymentInfo); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "create order fail")
		return fmt.Errorf("CreateOrderWithPayment err: %s", err.Error())
	}

	// notify
	go func() {
		notify.SendLarkOrderNotify(&notify.SendLarkOrderNotifyParam{
			Key:        config.Cfg.Notify.LarkRegisterKey,
			Action:     "renew account order",
			Account:    order.Account,
			OrderId:    order.OrderId,
			ChainType:  order.ChainType,
			Address:    order.Address,
			PayTokenId: order.PayTokenId,
			Amount:     order.PayAmount,
		})
	}()

	apiResp.ApiRespOK(resp)
	return nil
}
