package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type ReqAccountRegister struct {
	ReqAccountSearch
	ReqOrderRegisterBase

	//PayChainType common.ChainType  `json:"pay_chain_type"`
	//PayAddress   string            `json:"pay_address"`
	//PayTokenId   tables.PayTokenId `json:"pay_token_id"`
	//PayType      tables.PayType    `json:"pay_type"`
}

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

	if err = h.doAccountRegister(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doAccountRegister(&req, &apiResp); err != nil {
		log.Error("doAccountRegister err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountRegister(req *ReqAccountRegister, apiResp *api_code.ApiResp) error {
	var resp RespAccountRegister

	if req.Address == "" || req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}

	if len(req.AccountCharStr) == 0 {
		req.AccountCharStr = AccountToCharSet(req.Account)
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
	h.doInternalRegisterOrder(req, apiResp, &resp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func AccountToCharSet(account string) (accountChars []tables.AccountCharSet) {
	chars := []rune(account)
	for _, v := range chars {
		char := string(v)
		charSetName := tables.AccountCharTypeEmoji
		if strings.Contains(config.AccountCharSetEn, char) {
			charSetName = tables.AccountCharTypeEn
		} else if strings.Contains(config.AccountCharSetNumber, char) {
			charSetName = tables.AccountCharTypeNumber
		}
		accountChars = append(accountChars, tables.AccountCharSet{
			CharSetName: charSetName,
			Char:        char,
		})
	}
	return
}

func (h *HttpHandle) doInternalRegisterOrder(req *ReqAccountRegister, apiResp *api_code.ApiResp, resp *RespAccountRegister) {
	payTokenId := tables.TokenIdCkbInternal
	// pay amount
	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(req.Account, req.InviterAccount, req.RegisterYears, false, payTokenId)
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
		PayTokenId:        payTokenId,
		PayType:           "",
		PayAmount:         amountTotalPayToken,
		Content:           string(contentDataStr),
		PayStatus:         tables.TxStatusSending,
		HedgeStatus:       tables.TxStatusDefault,
		PreRegisterStatus: tables.TxStatusDefault,
		OrderStatus:       tables.OrderStatusDefault,
		RegisterStatus:    tables.RegisterStatusConfirmPayment,
	}
	order.CreateOrderId()
	resp.OrderId = order.OrderId

	if err := h.dbDao.CreateOrder(&order); err != nil {
		log.Error("CreateOrder err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "create order fail")
		return
	}
	// notify
	go func() {
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
