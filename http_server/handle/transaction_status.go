package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
)

type ReqTransactionStatus struct {
	core.ChainTypeAddress
	Actions    []tables.TxAction `json:"actions"`
	addressHex *core.DasAddressHex
}

type RespTransactionStatus struct {
	BlockNumber uint64          `json:"block_number"`
	Hash        string          `json:"hash"`
	Action      tables.TxAction `json:"action"`
	Status      int             `json:"status"`
}

func (h *HttpHandle) RpcTransactionStatus(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqTransactionStatus
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

	if err = h.doTransactionStatus(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doTransactionStatus err:", err.Error())
	}
}

func (h *HttpHandle) TransactionStatus(ctx *gin.Context) {
	var (
		funcName = "TransactionStatus"
		clientIp = GetClientIp(ctx)
		req      ReqTransactionStatus
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

	if err = h.doTransactionStatus(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doTransactionStatus err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTransactionStatus(ctx context.Context, req *ReqTransactionStatus, apiResp *api_code.ApiResp) error {
	var err error
	req.addressHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}

	var resp RespTransactionStatus
	actionList := make([]common.DasAction, 0)
	for _, v := range req.Actions {
		actionList = append(actionList, tables.FormatActionType(v))
	}

	tx, err := h.dbDao.GetPendingStatus(req.addressHex.ChainType, req.addressHex.AddressHex, actionList)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search tx status err")
		return fmt.Errorf("GetTransactionStatus err: %s", err.Error())
	}
	if tx.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeTransactionNotExist, "not exits tx")
		return nil
	}
	resp.BlockNumber = tx.BlockNumber
	resp.Hash, _ = common.String2OutPoint(tx.Outpoint)
	resp.Action = tables.FormatTxAction(tx.Action)
	resp.Status = tx.Status

	apiResp.ApiRespOK(resp)
	return nil
}
