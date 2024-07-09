package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/http_server/compatible"
	"das_register_server/tables"
	"das_register_server/unipay"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
)

type ReqOrderDetail struct {
	core.ChainTypeAddress
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Account   string           `json:"account"`
	Action    common.DasAction `json:"action"`
}

type RespOrderDetail struct {
	OrderId        string            `json:"order_id"`
	Account        string            `json:"account"`
	Action         common.DasAction  `json:"action"`
	Status         tables.TxStatus   `json:"status"`
	Timestamp      int64             `json:"timestamp"`
	PayTokenId     tables.PayTokenId `json:"pay_token_id"`
	PayAmount      decimal.Decimal   `json:"pay_amount"`
	PayType        tables.PayType    `json:"pay_type"`
	ReceiptAddress string            `json:"receipt_address"`
	InviterAccount string            `json:"inviter_account"`
	ChannelAccount string            `json:"channel_account"`
	RegisterYears  int               `json:"register_years"`
	//CodeUrl        string            `json:"code_url"` // wx pay code
	CoinType        string `json:"coin_type"`
	CrossCoinType   string `json:"cross_coin_type"`
	ContractAddress string `json:"contract_address"`
	ClientSecret    string `json:"client_secret"`
}

func (h *HttpHandle) RpcOrderDetail(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqOrderDetail
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

	if err = h.doOrderDetail(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doOrderDetail err:", err.Error())
	}
}

func (h *HttpHandle) OrderDetail(ctx *gin.Context) {
	var (
		funcName = "OrderDetail"
		clientIp = GetClientIp(ctx)
		req      ReqOrderDetail
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

	if err = h.doOrderDetail(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doOrderDetail err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderDetail(ctx context.Context, req *ReqOrderDetail, apiResp *api_code.ApiResp) error {
	var resp RespOrderDetail
	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	req.Account = strings.ToLower(req.Account)
	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	order, err := h.dbDao.GetLatestRegisterOrderBySelf(req.ChainType, req.Address, accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order fail")
		return fmt.Errorf("GetLatestOrderBySelf err: %s", err.Error())
	} else if order.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeOrderNotExist, "order not exist")
		return nil
	}

	resp.OrderId = order.OrderId
	resp.Account = order.Account
	resp.Action = order.Action
	resp.PayTokenId = order.PayTokenId
	resp.PayType = order.PayType
	resp.PayAmount = order.PayAmount
	resp.Timestamp = order.Timestamp
	resp.Status = order.PayStatus
	resp.CoinType = order.CoinType
	resp.CrossCoinType = order.CrossCoinType

	switch order.PayTokenId {
	case tables.TokenIdStripeUSD, tables.TokenIdTrc20USDT,
		tables.TokenIdBep20USDT, tables.TokenIdErc20USDT:
		unipayRes, err := unipay.GetOrderInfo(unipay.ReqOrderInfo{
			BusinessId: unipay.BusinessIdDasRegisterSvr,
			OrderId:    order.OrderId,
		})
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to call unipay.GetOrderInfo")
			return fmt.Errorf("unipay.GetOrderInfo err: %s[%s]", err.Error(), order.OrderId)
		}
		resp.ContractAddress = unipayRes.ContractAddress
		resp.ClientSecret = unipayRes.ClientSecret
	}

	if req.Action == common.DasActionApplyRegister {
		var contentData tables.TableOrderContent
		if err := json.Unmarshal([]byte(order.Content), &contentData); err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "json.Unmarshal fail")
			return fmt.Errorf("json.Unmarshal err: %s [%s]", err.Error(), order.Content)
		}
		resp.InviterAccount = contentData.InviterAccount
		resp.RegisterYears = contentData.RegisterYears
		resp.ChannelAccount = contentData.ChannelAccount
	}
	addr := config.GetUnipayAddress(order.PayTokenId)
	if addr == "" {
		apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("not supported [%s]", order.PayTokenId))
		return nil
	} else {
		resp.ReceiptAddress = addr
	}
	//resp.CodeUrl = ""

	apiResp.ApiRespOK(resp)
	return nil
}
