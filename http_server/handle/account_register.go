package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

type ReqAccountRegister struct {
	ReqAccountSearch
	ReqOrderRegisterBase
	CoinType string   `json:"coin_type"`
	MintFrom MintFrom `json:"mint_from"`
}

type MintFrom string

const (
	MintFromDefault MintFrom = ""
	MintFromPadge   MintFrom = "padge"
)

type RespAccountRegister struct {
	OrderId string `json:"order_id"`
}

func (h *HttpHandle) RpcAccountRegister(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountRegister
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

	if err = h.doAccountRegister(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doAccountRegister err:", err.Error())
	}
}

func (h *HttpHandle) AccountRegister(ctx *gin.Context) {
	var (
		funcName = "AccountRegister"
		clientIp = GetClientIp(ctx)
		req      ReqAccountRegister
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

	if err = h.doAccountRegister(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAccountRegister err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountRegister(ctx context.Context, req *ReqAccountRegister, apiResp *api_code.ApiResp) error {
	var resp RespAccountRegister

	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}

	if len(req.AccountCharStr) == 0 {
		accountChars, err := h.dasCore.GetAccountCharSetList(req.Account)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
			return nil
		}
		req.AccountCharStr = accountChars
		log.Info(ctx, "AccountToAccountChars:", toolib.JsonString(req.AccountCharStr))
	}

	var err error
	req.addressHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}

	if !checkChainType(req.addressHex.ChainType) {
		if req.addressHex.ChainType != common.ChainTypeAnyLock {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("chain type [%d] invalid", req.addressHex.ChainType))
			return nil
		}
	}

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	// order check
	if err := h.checkOrderInfo("", &req.ReqOrderRegisterBase, apiResp); err != nil {
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
	h.doInternalRegisterOrder(ctx, req, apiResp, &resp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) doInternalRegisterOrder(ctx context.Context, req *ReqAccountRegister, apiResp *api_code.ApiResp, resp *RespAccountRegister) {
	payTokenId := tables.TokenIdCkbInternal
	if req.MintFrom == MintFromPadge {
		payTokenId = tables.TokenIdPadgeInternal
	}
	// pay amount
	accLen := uint8(len(req.ReqAccountSearch.AccountCharStr))
	if tables.EndWithDotBitChar(req.ReqAccountSearch.AccountCharStr) {
		accLen -= 4
	}
	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(ctx, req.addressHex, accLen, req.Account, req.InviterAccount, req.RegisterYears, false, payTokenId)
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
		log.Error(ctx, "json marshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json marshal fail")
		return
	}
	order := tables.TableDasOrderInfo{
		OrderType:         tables.OrderTypeSelf,
		AccountId:         accountId,
		Account:           req.Account,
		Action:            common.DasActionApplyRegister,
		ChainType:         req.addressHex.ChainType,
		Address:           req.addressHex.AddressHex,
		Timestamp:         time.Now().UnixMilli(),
		PayTokenId:        payTokenId,
		PayAmount:         amountTotalPayToken,
		Content:           string(contentDataStr),
		PayStatus:         tables.TxStatusSending,
		HedgeStatus:       tables.TxStatusDefault,
		PreRegisterStatus: tables.TxStatusDefault,
		OrderStatus:       tables.OrderStatusDefault,
		RegisterStatus:    tables.RegisterStatusConfirmPayment,
		CoinType:          req.CoinType,
	}
	order.CreateOrderId()
	resp.OrderId = order.OrderId

	if err := h.dbDao.CreateOrder(&order); err != nil {
		log.Error(ctx, "CreateOrder err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "create order fail")
		return
	}
	// notify
	go func() {
		defer api_code.RecoverPanic()
		notify.SendLarkOrderNotify(&notify.SendLarkOrderNotifyParam{
			Key:        config.Cfg.Notify.LarkRegisterKey,
			Action:     "internal register order",
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
