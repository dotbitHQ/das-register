package handle

import (
	"context"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqAccountRecords struct {
	Account string `json:"account"`
}

type RespAccountRecords struct {
	Records []RespAccountRecordsData `json:"records"`
}

type RespAccountRecordsData struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	Ttl   string `json:"ttl"`
}

func (h *HttpHandle) RpcAccountRecords(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountRecords
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

	if err = h.doAccountRecords(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doVersion err:", err.Error())
	}
}

func (h *HttpHandle) AccountRecords(ctx *gin.Context) {
	var (
		funcName = "AccountRecords"
		clientIp = GetClientIp(ctx)
		req      ReqAccountRecords
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

	if err = h.doAccountRecords(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAccountRecords err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountRecords(ctx context.Context, req *ReqAccountRecords, apiResp *api_code.ApiResp) error {
	var resp RespAccountRecords
	resp.Records = make([]RespAccountRecordsData, 0)

	// account
	req.Account = strings.ToLower(req.Account)
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	if acc.Id == 0 {
		apiResp.ApiRespOK(resp)
		return nil
	} else if acc.Status == tables.AccountStatusOnCross {
		apiResp.ApiRespOK(resp)
		return nil
	}

	list, err := h.dbDao.SearchRecordsByAccount(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search records err")
		return fmt.Errorf("SearchRecordsByAccount err: %s", err.Error())
	}
	for _, v := range list {
		resp.Records = append(resp.Records, RespAccountRecordsData{
			Key:   v.Key,
			Type:  v.Type,
			Label: v.Label,
			Value: v.Value,
			Ttl:   v.Ttl,
		})
	}

	apiResp.ApiRespOK(resp)
	return nil
}
