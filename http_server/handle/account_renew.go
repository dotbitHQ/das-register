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
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type ReqAccountRenew struct {
	core.ChainTypeAddress
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account"`
	RenewYears int              `json:"renew_years"`
}

type RespAccountRenew struct {
	OrderId string `json:"order_id"`
}

func (h *HttpHandle) RpcAccountRenew(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountRenew
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

	if err = h.doAccountRenew(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doAccountRenew err:", err.Error())
	}
}

func (h *HttpHandle) AccountRenew(ctx *gin.Context) {
	var (
		funcName = "AccountRenew"
		clientIp = GetClientIp(ctx)
		req      ReqAccountRenew
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

	if err = h.doAccountRenew(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAccountRenew err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountRenew(ctx context.Context, req *ReqAccountRenew, apiResp *api_code.ApiResp) error {
	var resp RespAccountRenew

	req.Account = strings.ToLower(req.Account)
	if req.Address == "" || req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	} else if !strings.HasSuffix(req.Account, common.DasAccountSuffix) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}

	addressHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	// check renew
	if req.RenewYears < 1 || req.RenewYears > config.Cfg.Das.MaxRegisterYears {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("renew years[%d] invalid", req.RenewYears))
		return nil
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account fail")
		return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
	} else if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("account is invalid"))
		return nil
	} else if acc.Status == tables.AccountStatusOnCross {
		apiResp.ApiRespErr(api_code.ApiCodeOnCross, "account on cross")
		return nil
	}

	// renew account
	h.doInternalRenewOrder(ctx, acc, req, apiResp, &resp)

	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) doInternalRenewOrder(ctx context.Context, acc tables.TableAccountInfo, req *ReqAccountRenew, apiResp *api_code.ApiResp, resp *RespAccountRenew) {
	accOutpoint := common.String2OutPointStruct(acc.Outpoint)
	accTx, err := h.dasCore.Client().GetTransaction(h.ctx, accOutpoint.TxHash)
	if err != nil {
		log.Error(ctx, "GetTransaction err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return
	}
	mapAcc, err := witness.AccountIdCellDataBuilderFromTx(accTx.Transaction, common.DataTypeNew)
	if err != nil {
		log.Error(ctx, "GetTransaction err: ", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return
	}
	accBuilder, ok := mapAcc[acc.AccountId]
	if !ok {
		log.Error(ctx, "mapAcc is nil")
		apiResp.ApiRespErr(api_code.ApiCodeError500, "mapAcc is nil")
		return
	}
	//
	payTokenId := tables.TokenIdCkbInternal
	// pay amount
	amountTotalUSD, amountTotalCKB, amountTotalPayToken, err := h.getOrderAmount(ctx, uint8(accBuilder.AccountChars.Len()), "", req.Account, "", req.RenewYears, true, payTokenId)
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
		AmountTotalUSD: amountTotalUSD,
		AmountTotalCKB: amountTotalCKB,
		RenewYears:     req.RenewYears,
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
		Action:            common.DasActionRenewAccount,
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
		RegisterStatus:    tables.RegisterStatusConfirmPayment,
		OrderStatus:       tables.OrderStatusDefault,
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
			Action:     "internal renew order",
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
