package handle

import (
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
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

	PayChainType common.ChainType  `json:"pay_chain_type"`
	PayTokenId   tables.PayTokenId `json:"pay_token_id"`
	PayAddress   string            `json:"pay_address"`
	PayType      tables.PayType    `json:"pay_type"`

	ReqOrderRegisterBase
}

type RespOrderChange struct {
	OrderId        string            `json:"order_id"`
	TokenId        tables.PayTokenId `json:"token_id"`
	ReceiptAddress string            `json:"receipt_address"`
	Amount         decimal.Decimal   `json:"amount"`
	CodeUrl        string            `json:"code_url"`
	PayType        tables.PayType    `json:"pay_type"`
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
	if err := h.checkOrderInfo(&req.ReqOrderRegisterBase, apiResp); err != nil {
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
	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(uint8(len(oldOrderContent.AccountCharStr)-4), common.Bytes2Hex(args), req.Account, req.InviterAccount, req.RegisterYears, false, req.PayTokenId)
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
		RegisterStatus:    tables.RegisterStatusConfirmPayment,
		OrderStatus:       tables.OrderStatusDefault,
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
