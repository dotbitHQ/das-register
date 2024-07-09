package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqDidCellRecycle struct {
	core.ChainTypeAddress
	Account string `json:"account"`
}

type RespDidCellRecycle struct {
	SignInfo
}

func (h *HttpHandle) RpcDidCellRecycle(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellRecycle
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

	if err = h.doDidCellRecycle(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellRecycle err:", err.Error())
	}
}

func (h *HttpHandle) DidCellRecycle(ctx *gin.Context) {
	var (
		funcName = "DidCellRecycle"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellRecycle
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doDidCellRecycle(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellRecycle err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellRecycle(ctx context.Context, req *ReqDidCellRecycle, apiResp *http_api.ApiResp) error {
	var resp RespDidCellRecycle

	req.Account = strings.ToLower(req.Account)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	} else if addrHex.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}
	args := common.Bytes2Hex(addrHex.ParsedAddress.Script.Args)
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))

	didAccount, err := h.dbDao.GetDidAccountByAccountId(accountId, args)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account info")
		return fmt.Errorf("GetDidAccountByAccountId err: %s", err.Error())
	} else if didAccount.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "account not exist")
		return nil
	}

	expiredAt := tables.GetDidCellRecycleExpiredAt()
	if didAccount.ExpiredAt > expiredAt {
		apiResp.ApiRespErr(http_api.ApiCodeNotYetDueForRecycle, "not yet due for recycle")
		return nil
	}

	didCellOutpoint := common.String2OutPointStruct(didAccount.Outpoint)
	txParams, err := txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
		DasCore:         h.dasCore,
		DasCache:        h.dasCache,
		Action:          common.DidCellActionRecycle,
		DidCellOutPoint: didCellOutpoint,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build recycle tx")
		return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
	}

	reqBuild := reqBuildTx{
		Action:    common.DidCellActionRecycle,
		ChainType: addrHex.ChainType,
		Address:   addrHex.AddressHex,
		Account:   req.Account,
	}
	if didCellTx, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
		doBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
		resp.SignInfo.CKBTx = didCellTx
	}

	apiResp.ApiRespOK(resp)
	return nil
}
