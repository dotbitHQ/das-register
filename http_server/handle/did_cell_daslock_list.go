package handle

import (
	"context"
	"das_register_server/config"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqDidCellDasLockList struct {
	core.ChainTypeAddress
	Pagination
}

type RespDidCellDasLockList struct {
	Total int64        `json:"total"`
	List  []DidAccount `json:"list"`
}

func (h *HttpHandle) RpcDidCellDasLockList(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellDasLockList
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

	if err = h.doDidCellDasLockList(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellDasLockList err:", err.Error())
	}
}

func (h *HttpHandle) DidCellDasLockList(ctx *gin.Context) {
	var (
		funcName = "DidCellDasLockList"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellDasLockList
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doDidCellDasLockList(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellDasLockList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellDasLockList(ctx context.Context, req *ReqDidCellDasLockList, apiResp *http_api.ApiResp) error {
	var resp RespDidCellDasLockList
	resp.List = make([]DidAccount, 0)

	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	} else if addrHex.DasAlgorithmId == common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}

	dasLock, _, err := h.dasCore.Daf().HexToScript(*addrHex)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}
	args := common.Bytes2Hex(dasLock.Args)
	codeHash := dasLock.CodeHash.String()
	list, err := h.dbDao.GetDasLockDidCellList(args, codeHash, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account list")
		return fmt.Errorf("GetDasLockDidCellList err: %s", err.Error())
	}

	for _, v := range list {
		didAcc := DidAccount{
			Account:   v.Account,
			ExpiredAt: v.ExpiredAt * 1000,
		}
		resp.List = append(resp.List, didAcc)
	}

	count, err := h.dbDao.GetDasLockDidCellListTotal(args, codeHash)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account count")
		return fmt.Errorf("GetDasLockDidCellListTotal err: %s", err.Error())
	}
	resp.Total = count

	apiResp.ApiRespOK(resp)
	return nil
}
