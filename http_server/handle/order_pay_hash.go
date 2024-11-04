package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
	"time"
)

type ReqOrderPayHash struct {
	core.ChainTypeAddress
	Account    string `json:"account"`
	OrderId    string `json:"order_id"`
	PayHash    string `json:"pay_hash"`
	addressHex *core.DasAddressHex
}

type RespOrderPayHash struct {
}

func (h *HttpHandle) RpcOrderPayHash(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqOrderPayHash
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

	if err = h.doOrderPayHash(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doOrderPayHash err:", err.Error())
	}
}

func (h *HttpHandle) OrderPayHash(ctx *gin.Context) {
	var (
		funcName = "OrderPayHash"
		clientIp = GetClientIp(ctx)
		req      ReqOrderPayHash
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

	if err = h.doOrderPayHash(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doOrderPayHash err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderPayHash(ctx context.Context, req *ReqOrderPayHash, apiResp *api_code.ApiResp) error {
	var resp RespOrderPayHash
	if req.Account == "" || req.OrderId == "" || req.PayHash == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	req.Account = strings.ToLower(req.Account)

	var err error
	req.addressHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}

	order, err := h.dbDao.GetOrderByOrderId(req.OrderId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order fail")
		return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
	} else if order.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeOrderNotExist, "order not exist")
		return nil
	} else if !strings.EqualFold(order.Account, req.Account) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("account[%s] does not match the order", req.Account))
		return nil
	} else if req.addressHex.ChainType != order.ChainType || !strings.EqualFold(req.addressHex.AddressHex, order.Address) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "order's owner does not match")
		return nil
	}
	payInfo := tables.TableDasOrderPayInfo{
		Id:           0,
		Hash:         req.PayHash,
		OrderId:      req.OrderId,
		ChainType:    req.addressHex.ChainType,
		Address:      req.addressHex.AddressHex,
		Status:       tables.OrderTxStatusDefault,
		RefundStatus: tables.TxStatusDefault,
		RefundHash:   "",
		AccountId:    order.AccountId,
		Timestamp:    time.Now().UnixNano() / 1e6,
	}

	if err := h.dbDao.CreateOrderPayInfo(&payInfo); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "update hash fail")
		return fmt.Errorf("CreateOrderPayInfo err: %s", err.Error())
	}

	apiResp.ApiRespOK(resp)
	return nil
}
