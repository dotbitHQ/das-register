package handle

import (
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqReverseList struct {
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
}

type RespReverseList struct {
	List []ReverseListData `json:"list"`
}

type ReverseListData struct {
	Account     string `json:"account"`
	BlockNumber uint64 `json:"block_number"`
	Hash        string `json:"hash"`
	Index       uint   `json:"index"`
}

func (h *HttpHandle) RpcReverseList(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqReverseList
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

	if err = h.doReverseList(&req[0], apiResp); err != nil {
		log.Error("doReverseList err:", err.Error())
	}
}

func (h *HttpHandle) ReverseList(ctx *gin.Context) {
	var (
		funcName = "ReverseList"
		clientIp = GetClientIp(ctx)
		req      ReqReverseList
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

	if err = h.doReverseList(&req, &apiResp); err != nil {
		log.Error("doReverseList err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseList(req *ReqReverseList, apiResp *api_code.ApiResp) error {
	req.Address = core.FormatAddressToHex(req.ChainType, req.Address)

	var resp RespReverseList
	resp.List = make([]ReverseListData, 0)

	list, err := h.dbDao.SearchReverseList(req.ChainType, req.Address)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search reverse list err")
		return fmt.Errorf("SearchLatestReverse err: %s", err.Error())
	}
	for _, v := range list {
		hash, index := common.String2OutPoint(v.Outpoint)
		resp.List = append(resp.List, ReverseListData{
			Account:     v.Account,
			BlockNumber: v.BlockNumber,
			Hash:        hash,
			Index:       index,
		})
	}

	apiResp.ApiRespOK(resp)
	return nil
}