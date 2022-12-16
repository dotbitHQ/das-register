package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/notify"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type ReqOrderRenew struct {
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Account   string           `json:"account"`

	PayChainType common.ChainType  `json:"pay_chain_type"`
	PayTokenId   tables.PayTokenId `json:"pay_token_id"`
	PayAddress   string            `json:"pay_address"`
	PayType      tables.PayType    `json:"pay_type"`

	RenewYears int `json:"renew_years"`
}

type RespOrderRenew struct {
	OrderId        string            `json:"order_id"`
	TokenId        tables.PayTokenId `json:"token_id"`
	ReceiptAddress string            `json:"receipt_address"`
	Amount         decimal.Decimal   `json:"amount"`
	CodeUrl        string            `json:"code_url"`
	PayType        tables.PayType    `json:"pay_type"`
}

func (h *HttpHandle) RpcOrderRenew(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqOrderRenew
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

	if err = h.doOrderRenew(&req[0], apiResp); err != nil {
		log.Error("doOrderRenew err:", err.Error())
	}
}

func (h *HttpHandle) OrderRenew(ctx *gin.Context) {
	var (
		funcName = "OrderRenew"
		clientIp = GetClientIp(ctx)
		req      ReqOrderRenew
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

	if err = h.doOrderRenew(&req, &apiResp); err != nil {
		log.Error("doOrderRenew err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderRenew(req *ReqOrderRenew, apiResp *api_code.ApiResp) error {
	var resp RespOrderRenew

	if req.Account == "" || req.Address == "" || !strings.HasSuffix(req.Account, common.DasAccountSuffix) {
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

	if exi := h.rc.AccountLimitExist(req.Account); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
		return fmt.Errorf("AccountActionLimitExist: %d %s %s", req.ChainType, req.Address, req.Account)
	}

	acc := h.checkRenewOrder(req, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	h.doRenewOrder(acc, req, apiResp, &resp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	// account limit
	_ = h.rc.SetAccountLimit(req.Account, time.Second*30)
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) doRenewOrder(acc *tables.TableAccountInfo, req *ReqOrderRenew, apiResp *api_code.ApiResp, resp *RespOrderRenew) {
	if acc == nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "acc is nil")
		return
	}
	// pay amount
	addrHex := core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	}
	args, err := h.dasCore.Daf().HexToArgs(addrHex, addrHex)

	//
	accOutpoint := common.String2OutPointStruct(acc.Outpoint)
	accTx, err := h.dasCore.Client().GetTransaction(h.ctx, accOutpoint.TxHash)
	if err != nil {
		log.Error("GetTransaction err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return
	}
	mapAcc, err := witness.AccountIdCellDataBuilderFromTx(accTx.Transaction, common.DataTypeNew)
	if err != nil {
		log.Error("GetTransaction err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return
	}
	accBuilder, ok := mapAcc[acc.AccountId]
	if !ok {
		log.Error("mapAcc is nil")
		apiResp.ApiRespErr(api_code.ApiCodeError500, "mapAcc is nil")
		return
	}

	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(uint8(accBuilder.AccountChars.Len()), common.Bytes2Hex(args), req.Account, "", req.RenewYears, true, req.PayTokenId)
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
		AmountTotalUSD: amountTotalUSD,
		AmountTotalCKB: amountTotalCKB,
		RenewYears:     req.RenewYears,
	}
	contentDataStr, err := json.Marshal(&orderContent)
	if err != nil {
		log.Error("json marshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json marshal fail")
		return
	}
	if req.PayTokenId == tables.TokenIdDas {
		dasLock, _, err := h.dasCore.Daf().HexToScript(addrHex)
		if err != nil {
			log.Error("HexToArgs err: ", err.Error())
			apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
			return
		}

		fee := common.OneCkb
		needCapacity := amountTotalPayToken.BigInt().Uint64()
		_, _, err = h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
			DasCache:          h.dasCache,
			LockScript:        dasLock,
			CapacityNeed:      needCapacity + fee,
			CapacityForChange: common.DasLockWithBalanceTypeOccupiedCkb,
			SearchOrder:       indexer.SearchOrderDesc,
		})
		if err != nil {
			checkBalanceErr(err, apiResp)
			return
		}
	}
	order := tables.TableDasOrderInfo{
		Id:                0,
		OrderType:         tables.OrderTypeSelf,
		OrderId:           "",
		AccountId:         accountId,
		Account:           req.Account,
		Action:            common.DasActionRenewAccount,
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
		RegisterStatus:    tables.RegisterStatusDefault,
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
			Action:     "renew account order",
			Account:    order.Account,
			OrderId:    order.OrderId,
			ChainType:  order.ChainType,
			Address:    order.Address,
			PayTokenId: order.PayTokenId,
			Amount:     order.PayAmount,
		})
	}()
}

func (h *HttpHandle) checkRenewOrder(req *ReqOrderRenew, apiResp *api_code.ApiResp) *tables.TableAccountInfo {
	if req.RenewYears < 1 || req.RenewYears > config.Cfg.Das.MaxRegisterYears {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("renew years[%d] invalid", req.RenewYears))
		return nil
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		log.Error("GetAccountInfoByAccountId err: ", err.Error())
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
	log.Info("checkRenewOrder:", expirationGracePeriod)
	expirationGracePeriod -= 2 * 60 * 60
	log.Info("checkRenewOrder:", expirationGracePeriod)
	if int64(acc.ExpiredAt+uint64(expirationGracePeriod)) <= time.Now().Unix() {
		apiResp.ApiRespErr(api_code.ApiCodeAfterGracePeriod, "after the grace period")
		return nil
	}
	return &acc
}
