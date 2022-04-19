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
	"net/http"
)

type ReqTransactionList struct {
	Pagination
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
}
type RespTransactionList struct {
	Total int64             `json:"total"`
	List  []DataTransaction `json:"list"`
}

type DataTransaction struct {
	Hash        string          `json:"hash"`
	BlockNumber uint64          `json:"block_number"`
	Action      tables.TxAction `json:"action"`
	Account     string          `json:"account"`
	Capacity    uint64          `json:"capacity"`
	Timestamp   uint64          `json:"timestamp"`
}

func (h *HttpHandle) RpcTransactionList(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqTransactionList
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

	if err = h.doTransactionList(&req[0], apiResp); err != nil {
		log.Error("doTransactionList err:", err.Error())
	}
}

func (h *HttpHandle) TransactionList(ctx *gin.Context) {
	var (
		funcName = "TransactionList"
		clientIp = GetClientIp(ctx)
		req      ReqTransactionList
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

	if err = h.doTransactionList(&req, &apiResp); err != nil {
		log.Error("doTransactionList err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTransactionList(req *ReqTransactionList, apiResp *api_code.ApiResp) error {
	req.Address = core.FormatAddressToHex(req.ChainType, req.Address)
	var resp RespTransactionList
	resp.List = make([]DataTransaction, 0)

	list, err := h.dbDao.GetTransactionList(req.ChainType, req.Address, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search tx list err")
		return fmt.Errorf("GetTransactionList err: %s", err.Error())
	}
	for _, v := range list {
		hash, _ := common.String2OutPoint(v.Outpoint)
		resp.List = append(resp.List, DataTransaction{
			Hash:        hash,
			BlockNumber: v.BlockNumber,
			Action:      tables.FormatTxAction(v.Action),
			Account:     v.Account,
			Capacity:    v.Capacity,
			Timestamp:   v.BlockTimestamp,
		})
	}
	// 总记录数
	count, err := h.dbDao.GetTransactionListTotal(req.ChainType, req.Address)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search tx count err")
		return fmt.Errorf("GetTransactionListTotal err: %s", err.Error())
	}
	resp.Total = count

	apiResp.ApiRespOK(resp)
	return nil
}
