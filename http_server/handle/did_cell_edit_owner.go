package handle

import (
	"bytes"
	"context"
	"das_register_server/config"
	"das_register_server/internal"
	"das_register_server/tables"
	"das_register_server/timer"
	"das_register_server/unipay"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type ReqDidCellEditOwner struct {
	core.ChainTypeAddress
	Account  string `json:"account"`
	RawParam struct {
		ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
		ReceiverAddress  string          `json:"receiver_address"`
	} `json:"raw_param"`
	PayTokenId tables.PayTokenId `json:"pay_token_id"`
	IsUpgrade  bool              `json:"is_upgrade"`
}

type RespDidCellEditOwner struct {
	OrderId         string          `json:"order_id"`
	ReceiptAddress  string          `json:"receipt_address"`
	Amount          decimal.Decimal `json:"amount"`
	ContractAddress string          `json:"contract_address"`
	ClientSecret    string          `json:"client_secret"`
	SignInfo
}

func (h *HttpHandle) RpcDidCellEditOwner(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellEditOwner
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

	if err = h.doDidCellEditOwner(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellEditOwner err:", err.Error())
	}
}

func (h *HttpHandle) DidCellEditOwner(ctx *gin.Context) {
	var (
		funcName = "DidCellEditOwner"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellEditOwner
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

	if err = h.doDidCellEditOwner(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellEditOwner err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellEditOwner(ctx context.Context, req *ReqDidCellEditOwner, apiResp *http_api.ApiResp) error {
	var resp RespDidCellEditOwner

	req.Account = strings.ToLower(req.Account)
	addrHexFrom, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	}

	switch addrHexFrom.DasAlgorithmId {
	case common.DasAlgorithmIdBitcoin:
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}
	toCTA := core.ChainTypeAddress{
		Type: "blockchain",
		KeyInfo: core.KeyInfo{
			CoinType: req.RawParam.ReceiverCoinType,
			Key:      req.RawParam.ReceiverAddress,
		},
	}
	addrHexTo, err := toCTA.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeInvalidTargetAddress, "receiver address is invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	}

	if req.IsUpgrade {
		if addrHexTo.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
			apiResp.ApiRespErr(http_api.ApiCodeAnyLockAddressInvalid, "address invalid")
			return nil
		}
		if addrHexTo.DasAlgorithmId == common.DasAlgorithmIdAnyLock &&
			addrHexTo.ParsedAddress.Script.CodeHash.Hex() == transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
			apiResp.ApiRespErr(http_api.ApiCodeAnyLockAddressInvalid, "address invalid")
			return nil
		}
	} else {
		if addrHexFrom.DasAlgorithmId == common.DasAlgorithmIdAnyLock && addrHexTo.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
			apiResp.ApiRespErr(http_api.ApiCodeAnyLockAddressInvalid, "address invalid")
			return nil
		}
		if addrHexFrom.DasAlgorithmId == common.DasAlgorithmIdAnyLock &&
			addrHexTo.DasAlgorithmId == common.DasAlgorithmIdAnyLock &&
			addrHexTo.ParsedAddress.Script.CodeHash.Hex() == transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
			apiResp.ApiRespErr(http_api.ApiCodeAnyLockAddressInvalid, "address invalid")
			return nil
		}
		if addrHexFrom.DasAlgorithmId != common.DasAlgorithmIdAnyLock &&
			addrHexTo.DasAlgorithmId == common.DasAlgorithmIdAnyLock {
			apiResp.ApiRespErr(http_api.ApiCodeInvalidTargetAddress, "address invalid")
			return nil
		}
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	if strings.EqualFold(req.KeyInfo.Key, req.RawParam.ReceiverAddress) {
		apiResp.ApiRespErr(http_api.ApiCodeSameLock, "same address")
		return nil
	}
	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(http_api.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}
	var editOwnerCapacity uint64
	var didCellOutPoint, accountCellOutPoint *types.OutPoint
	var editOwnerLock, normalCellScript *types.Script

	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get account info")
		return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
	} else if acc.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.IsExpired() {
		apiResp.ApiRespErr(http_api.ApiCodeAccountIsExpired, "account expired")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "not support sub account")
		return nil
	}
	if acc.Status == tables.AccountStatusNormal {
		accountCellOutPoint = acc.GetOutpoint()
		if addrHexFrom.ChainType != acc.OwnerChainType || !strings.EqualFold(addrHexFrom.AddressHex, acc.Owner) {
			apiResp.ApiRespErr(http_api.ApiCodePermissionDenied, "transfer owner permission denied")
			return nil
		}
	} else if acc.Status == tables.AccountStatusOnUpgrade {
		didAccount, err := h.dbDao.GetDidAccountByAccountIdWithoutArgs(accountId)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did cell info")
			return fmt.Errorf("GetDidAccountByAccountId err: %s", err.Error())
		} else if didAccount.Id == 0 {
			apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "did cell not exist")
			return nil
		} else if addrHexFrom.ParsedAddress == nil || bytes.Compare(common.Hex2Bytes(didAccount.Args), addrHexFrom.ParsedAddress.Script.Args) != 0 {
			apiResp.ApiRespErr(http_api.ApiCodeNoAccountPermissions, "transfer account permission denied")
			return nil
		}
		didCellOutPoint = didAccount.GetOutpoint()
	} else {
		apiResp.ApiRespErr(http_api.ApiCodeAccountStatusNotNormal, "account status is not normal")
		return nil
	}

	if addrHexFrom.DasAlgorithmId == common.DasAlgorithmIdAnyLock && addrHexTo.DasAlgorithmId == common.DasAlgorithmIdAnyLock {
		// did cell -> did cell
		editOwnerLock = addrHexTo.ParsedAddress.Script
	} else if addrHexFrom.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		if addrHexTo.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
			// account cell -> account cell
			editOwnerLock, _, err = h.dasCore.Daf().HexToScript(*addrHexTo)
			if err != nil {
				apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "receiver address is invalid")
				return nil
			}
		} else {
			// account cell -> did cell
			editOwnerLock = addrHexTo.ParsedAddress.Script
			editOwnerCapacity, err = h.dasCore.GetDidCellOccupiedCapacity(editOwnerLock, req.Account)
			if err != nil {
				apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to get did cell capacity")
				return fmt.Errorf("GetDidCellOccupiedCapacity err: %s", err.Error())
			}
			log.Info(ctx, "GetDidCellOccupiedCapacity:", editOwnerCapacity)
		}
	} else {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid")
		return nil
	}

	parseSvrAddr, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	normalCellScript = parseSvrAddr.Script
	txParams, err := txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
		DasCore:             h.dasCore,
		DasCache:            h.dasCache,
		Action:              common.DidCellActionEditOwner,
		DidCellOutPoint:     didCellOutPoint,
		AccountCellOutPoint: accountCellOutPoint,
		EditOwnerLock:       editOwnerLock,
		NormalCellScript:    normalCellScript,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build tx")
		return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
	}

	if editOwnerCapacity > 0 {
		if req.PayTokenId == "" {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "pay token id is nil")
			return nil
		}
		var order tables.TableDasOrderInfo
		var paymentInfo tables.TableDasOrderPayInfo
		unipayAddr := config.GetUnipayAddress(req.PayTokenId)
		if unipayAddr == "" {
			apiResp.ApiRespErr(http_api.ApiCodeError500, fmt.Sprintf("not supported [%s]", req.PayTokenId))
			return nil
		}
		payToken := timer.GetTokenInfo(req.PayTokenId)
		if payToken.TokenId == "" {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid")
			return fmt.Errorf("timer.GetTokenInfo is nil [%s]", req.PayTokenId)
		}
		quoteCell, err := h.dasCore.GetQuoteCell()
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to get quote cell")
			return fmt.Errorf("GetQuoteCell err: %s", err.Error())
		}
		quote := quoteCell.Quote()

		editOwnerAmountUSD, _ := decimal.NewFromString(fmt.Sprintf("%d", editOwnerCapacity/common.OneCkb))
		decQuote, _ := decimal.NewFromString(fmt.Sprintf("%d", quote))
		decUsdRateBase := decimal.NewFromInt(common.UsdRateBase)
		editOwnerAmountUSD = editOwnerAmountUSD.Mul(decQuote).DivRound(decUsdRateBase, 6)
		amountTotalPayToken := editOwnerAmountUSD.Div(payToken.Price).Mul(decimal.New(1, payToken.Decimals)).Ceil()
		if payToken.TokenId == tables.TokenIdCkb {
			amountTotalPayToken, _ = decimal.NewFromString(fmt.Sprintf("%d", editOwnerCapacity))
		}
		log.Info(ctx, "edit owner amountTotalPayToken:", amountTotalPayToken, editOwnerCapacity)

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
			PaymentAddress:    unipayAddr,
			PremiumPercentage: premiumPercentage,
			PremiumBase:       premiumBase,
			PremiumAmount:     premiumAmount,
			MetaData: map[string]string{
				"account":      req.Account,
				"algorithm_id": addrHexFrom.ChainType.ToString(),
				"address":      req.ChainTypeAddress.KeyInfo.Key,
				"action":       "edit_owner",
			},
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to create order by unipay")
			return fmt.Errorf("unipay.CreateOrder err: %s", err.Error())
		}

		order = tables.TableDasOrderInfo{
			OrderType:         tables.OrderTypeSelf,
			OrderId:           res.OrderId,
			AccountId:         accountId,
			Account:           req.Account,
			Action:            common.DasActionTransferAccount,
			ChainType:         addrHexFrom.ChainType,
			Address:           addrHexFrom.AddressHex,
			Timestamp:         time.Now().UnixMilli(),
			PayTokenId:        req.PayTokenId,
			PayAmount:         amountTotalPayToken,
			PayStatus:         tables.TxStatusDefault,
			HedgeStatus:       tables.TxStatusDefault,
			PreRegisterStatus: tables.TxStatusDefault,
			RegisterStatus:    tables.RegisterStatusDefault,
			OrderStatus:       tables.OrderStatusDefault,
			IsUniPay:          tables.IsUniPayTrue,
			PremiumPercentage: premiumPercentage,
			PremiumBase:       premiumBase,
			PremiumAmount:     premiumAmount,
			IsDidCell:         tables.IsDidCellYes,
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

		resp.OrderId = order.OrderId
		resp.ReceiptAddress = unipayAddr
		resp.Amount = order.PayAmount
		resp.ContractAddress = res.ContractAddress
		resp.ClientSecret = res.ClientSecret

		if err := h.dbDao.CreateOrderWithPayment(order, paymentInfo); err != nil {
			log.Error(ctx, "CreateOrder err:", err.Error())
			apiResp.ApiRespErr(http_api.ApiCodeError500, "create order fail")
			return fmt.Errorf("CreateOrderWithPayment err: %s", err.Error())
		}
	}

	reqBuild := reqBuildTx{
		OrderId:    resp.OrderId,
		Action:     common.DasActionTransferAccount,
		ChainType:  addrHexFrom.ChainType,
		Address:    addrHexFrom.AddressHex,
		Account:    req.Account,
		EvmChainId: req.GetChainId(config.Cfg.Server.Net),
	}
	if didCellTx, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
		checkBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
		if acc.Status == tables.AccountStatusOnUpgrade {
			resp.SignInfo.CKBTx = didCellTx
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) isAnyLock(cta core.ChainTypeAddress, apiResp *http_api.ApiResp) (bool, *address.ParsedAddress, error) {
	if cta.KeyInfo.CoinType == common.CoinTypeCKB {
		addrParse, err := address.Parse(cta.KeyInfo.Key)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
			return false, nil, fmt.Errorf("address.Parse err: %s", err.Error())
		}
		contractDispatch, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to get dispatch contract")
			return false, nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		} else if !contractDispatch.IsSameTypeId(addrParse.Script.CodeHash) {
			return true, addrParse, nil
		}
	}
	return false, nil, nil
}
