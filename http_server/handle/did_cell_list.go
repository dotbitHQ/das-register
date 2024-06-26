package handle

import (
	"das_register_server/config"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqDidCellList struct {
	core.ChainTypeAddress
	Pagination
	Keyword string `json:"keyword"`
}

type RespDidCellList struct {
	Total int64        `json:"total"`
	List  []DidAccount `json:"list"`
}

type DidAccount struct {
	Account   string `json:"account"`
	ExpiredAt uint64 `json:"expired_at"`
}

func (h *HttpHandle) RpcDidCellList(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellList
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doDidCellList(&req[0], apiResp); err != nil {
		log.Error("doDidCellList err:", err.Error())
	}
}

func (h *HttpHandle) DidCellList(ctx *gin.Context) {
	var (
		funcName = "DidCellList"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellList
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doDidCellList(&req, &apiResp); err != nil {
		log.Error("doDidCellList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellList(req *ReqDidCellList, apiResp *http_api.ApiResp) error {
	var resp RespDidCellList
	resp.List = make([]DidAccount, 0)

	req.Keyword = strings.ToLower(req.Keyword)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	} else if addrHex.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}
	args := common.Bytes2Hex(addrHex.ParsedAddress.Script.Args)
	list, err := h.dbDao.GetDidAccountList(args, req.Keyword, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account list")
		return fmt.Errorf("GetDidAccountList err: %s", err.Error())
	}

	for _, v := range list {
		didAcc := DidAccount{
			Account:   v.Account,
			ExpiredAt: v.ExpiredAt,
		}
		resp.List = append(resp.List, didAcc)
	}

	count, err := h.dbDao.GetDidAccountListTotal(args, req.Keyword)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account count")
		return fmt.Errorf("GetDidAccountListTotal err: %s", err.Error())
	}
	resp.Total = count

	apiResp.ApiRespOK(resp)
	return nil
}
