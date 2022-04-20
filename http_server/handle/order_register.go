package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
	"das_register_server/timer"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ReqOrderRegister struct {
	ReqAccountSearch
	ReqOrderRegisterBase

	PayChainType common.ChainType  `json:"pay_chain_type"`
	PayAddress   string            `json:"pay_address"`
	PayTokenId   tables.PayTokenId `json:"pay_token_id"`
	PayType      tables.PayType    `json:"pay_type"`
}

type ReqOrderRegisterBase struct {
	RegisterYears  int    `json:"register_years"`
	InviterAccount string `json:"inviter_account"`
	ChannelAccount string `json:"channel_account"`
}

type RespOrderRegister struct {
	OrderId        string            `json:"order_id"`
	TokenId        tables.PayTokenId `json:"token_id"`
	ReceiptAddress string            `json:"receipt_address"`
	Amount         decimal.Decimal   `json:"amount"`
	CodeUrl        string            `json:"code_url"`
	PayType        tables.PayType    `json:"pay_type"`
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

	if err = h.doOrderRegister(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doOrderRegister(&req, &apiResp); err != nil {
		log.Error("doOrderRegister err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderRegister(req *ReqOrderRegister, apiResp *api_code.ApiResp) error {
	var resp RespOrderRegister

	if req.Address == "" || req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	if yes := req.PayTokenId.IsTokenIdCkbInternal(); yes {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("pay token id [%s] invalid", req.PayTokenId))
		return nil
	}
	if ok := checkRegisterChainTypeAndAddress(req.ChainType, req.Address); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("chain type and address [%s-%s] invalid", req.ChainType.String(), req.Address))
		return nil
	}

	req.Address = core.FormatAddressToHex(req.ChainType, req.Address)

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if exi := h.rc.RegisterLimitExist(req.ChainType, req.Address, req.Account, "1"); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
		return fmt.Errorf("AccountActionLimitExist: %d %s %s", req.ChainType, req.Address, req.Account)
	}

	// order check
	if err := h.checkOrderInfo(&req.ReqOrderRegisterBase, apiResp); err != nil {
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
	status, _ := h.checkAccountBase(&req.ReqAccountSearch, apiResp)
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
	status, _ = h.checkAddressOrder(&req.ReqAccountSearch, apiResp, false)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if status != tables.SearchStatusRegisterAble {
		apiResp.ApiRespErr(api_code.ApiCodeAccountAlreadyRegister, "account registering")
		return nil
	}
	// registering check
	status = h.checkOtherAddressOrder(&req.ReqAccountSearch, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if status >= tables.SearchStatusRegistering {
		apiResp.ApiRespErr(api_code.ApiCodeAccountAlreadyRegister, "account registering")
		return nil
	}

	// create order
	h.doRegisterOrder(req, apiResp, &resp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}
	// cache
	_ = h.rc.SetRegisterLimit(req.ChainType, req.Address, req.Account, "1", time.Second*30)
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) checkOrderInfo(req *ReqOrderRegisterBase, apiResp *api_code.ApiResp) error {
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
		}
	}

	if req.ChannelAccount != "" {
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.ChannelAccount))
		acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search channel account fail")
			return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
		} else if acc.Id == 0 {
			apiResp.ApiRespErr(api_code.ApiCodeChannelAccountNotExist, "channel account not exist")
			return nil
		}
	}
	return nil
}

func (h *HttpHandle) doRegisterOrder(req *ReqOrderRegister, apiResp *api_code.ApiResp, resp *RespOrderRegister) {
	// pay amount
	args := common.Bytes2Hex(core.FormatOwnerManagerAddressToArgs(req.ChainType, req.ChainType, req.Address, req.Address))
	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(args, req.Account, req.InviterAccount, req.RegisterYears, false, req.PayTokenId)
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
		log.Error("json marshal err:", err.Error())
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
		PayStatus:         tables.TxStatusDefault,
		HedgeStatus:       tables.TxStatusDefault,
		PreRegisterStatus: tables.TxStatusDefault,
		OrderStatus:       tables.OrderStatusDefault,
		RegisterStatus:    tables.RegisterStatusConfirmPayment,
	}
	order.CreateOrderId()
	resp.OrderId = order.OrderId
	resp.TokenId = req.PayTokenId
	resp.PayType = req.PayType
	resp.Amount = order.PayAmount
	resp.CodeUrl = ""
	if addr, ok := config.Cfg.PayAddressMap[order.PayTokenId.ToChainString()]; !ok {
		apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("not supported [%s]", order.PayTokenId))
		return
	} else {
		resp.ReceiptAddress = addr
	}

	if err := h.dbDao.CreateOrder(&order); err != nil {
		log.Error("CreateOrder err:", err.Error())
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

func (h *HttpHandle) getOrderAmount(args, account, inviterAccount string, years int, isRenew bool, payTokenId tables.PayTokenId) (amountTotalUSD decimal.Decimal, amountTotalCKB decimal.Decimal, amountTotalPayToken decimal.Decimal, e error) {
	// pay token
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
	baseAmount, accountPrice, err := h.getAccountPrice(args, account, isRenew)
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
	amountTotalUSD = baseAmount.Add(accountPrice)

	log.Info("before Premium:", account, isRenew, amountTotalUSD, baseAmount, accountPrice)
	if config.Cfg.Das.Premium.Cmp(decimal.Zero) == 1 {
		amountTotalUSD = amountTotalUSD.Mul(config.Cfg.Das.Premium.Add(decimal.NewFromInt(1)))
	}
	log.Info("after Premium:", account, isRenew, amountTotalUSD, baseAmount, accountPrice)

	amountTotalUSD = amountTotalUSD.Mul(decimal.NewFromInt(100)).Ceil().DivRound(decimal.NewFromInt(100), 2)
	amountTotalCKB = amountTotalUSD.Div(decQuote).Mul(decimal.NewFromInt(int64(common.OneCkb))).Ceil()
	amountTotalPayToken = amountTotalUSD.Div(payToken.Price).Mul(decimal.New(1, payToken.Decimals)).Ceil()

	log.Info("getOrderAmount:", amountTotalUSD, amountTotalCKB, amountTotalPayToken)
	if payToken.TokenId == tables.TokenIdCkb {
		amountTotalPayToken = amountTotalCKB
	} else if payToken.TokenId == tables.TokenIdMatic || payToken.TokenId == tables.TokenIdBnb || payToken.TokenId == tables.TokenIdEth {
		log.Info("amountTotalPayToken:", amountTotalPayToken.String())
		decCeil := decimal.NewFromInt(1e6)
		amountTotalPayToken = amountTotalPayToken.DivRound(decCeil, 6).Ceil().Mul(decCeil)
		log.Info("amountTotalPayToken:", amountTotalPayToken.String())
	}
	return
}

func checkRegisterChainTypeAndAddress(chainType common.ChainType, address string) bool {
	switch chainType {
	case common.ChainTypeTron:
		if strings.HasPrefix(address, common.TronPreFix) {
			if _, err := common.TronHexToBase58(address); err != nil {
				log.Error("TronHexToBase58 err:", err.Error(), address)
				return false
			}
			return true
		} else if strings.HasPrefix(address, common.TronBase58PreFix) {
			if _, err := common.TronBase58ToHex(address); err != nil {
				log.Error("TronBase58ToHex err:", err.Error(), address)
				return false
			}
			return true
		}
	case common.ChainTypeEth:
		if ok, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address); ok {
			return true
		}
	case common.ChainTypeMixin:
		if ok, _ := regexp.MatchString("^0x[0-9a-fA-F]{64}$", address); ok {
			return true
		}
	}
	return false
	//switch chainType {
	//case common.ChainTypeTron:
	//	if strings.HasPrefix(address, common.TronPreFix) || strings.HasPrefix(address, common.TronBase58PreFix) {
	//		return true
	//	}
	//case common.ChainTypeEth:
	//	if strings.HasPrefix(address, common.HexPreFix) && len(address) == 42 {
	//		return true
	//	}
	//case common.ChainTypeMixin:
	//	if strings.HasPrefix(address, common.HexPreFix) && len(address) == 66 {
	//		return true
	//	}
	//}
	//return false
}
