package handle

import (
	"das_register_server/http_server/compatible"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
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
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Account   string           `json:"account"`
	OrderId   string           `json:"order_id"`
	PayHash   string           `json:"pay_hash"`
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

	if err = h.doOrderPayHash(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doOrderPayHash(&req, &apiResp); err != nil {
		log.Error("doOrderPayHash err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderPayHash(req *ReqOrderPayHash, apiResp *api_code.ApiResp) error {
	var resp RespOrderPayHash
	if req.Address == "" || req.Account == "" || req.OrderId == "" || req.PayHash == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	addressHex, err := compatible.ChaintyeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

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
	} else if req.ChainType != order.ChainType || !strings.EqualFold(req.Address, order.Address) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "order's owner does not match")
		return nil
	}
	payInfo := tables.TableDasOrderPayInfo{
		Id:           0,
		Hash:         req.PayHash,
		OrderId:      req.OrderId,
		ChainType:    req.ChainType,
		Address:      req.Address,
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
