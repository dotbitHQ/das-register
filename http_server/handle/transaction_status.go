package handle

import (
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
)

type ReqTransactionStatus struct {
	ChainType common.ChainType  `json:"chain_type"`
	Address   string            `json:"address"`
	Actions   []tables.TxAction `json:"actions"`
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

	if err = h.doTransactionStatus(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doTransactionStatus(&req, &apiResp); err != nil {
		log.Error("doTransactionStatus err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTransactionStatus(req *ReqTransactionStatus, apiResp *api_code.ApiResp) error {
	req.Address = core.FormatAddressToHex(req.ChainType, req.Address)

	var resp RespTransactionStatus
	actionList := make([]common.DasAction, 0)
	for _, v := range req.Actions {
		actionList = append(actionList, tables.FormatActionType(v))
	}

	tx, err := h.dbDao.GetPendingStatus(req.ChainType, req.Address, actionList)
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
