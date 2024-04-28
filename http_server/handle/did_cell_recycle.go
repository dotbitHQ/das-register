package handle

import (
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/scorpiotzh/toolib"
	"net/http"
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

	if err = h.doDidCellRecycle(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doDidCellRecycle(&req, &apiResp); err != nil {
		log.Error("doDidCellRecycle err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellRecycle(req *ReqDidCellRecycle, apiResp *http_api.ApiResp) error {
	var resp RespDidCellRecycle

	addr, err := address.Parse(req.ChainTypeAddress.KeyInfo.Key)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	args := common.Bytes2Hex(addr.Script.Args)
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
		DasCore:             h.dasCore,
		Action:              common.DidCellActionRecycle,
		DidCellOutPoint:     didCellOutpoint,
		AccountCellOutPoint: nil,
		EditRecords:         nil,
		EditOwnerLock:       nil,
		NormalCkbLiveCell:   nil,
		RenewYears:          0,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build recycle tx")
		return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
	}

	// todo
	reqBuild := reqBuildTx{
		Action:      common.DidCellActionRecycle,
		ChainType:   0,
		Address:     req.KeyInfo.Key,
		Account:     req.Account,
		Capacity:    0,
		EvmChainId:  0,
		AuctionInfo: AuctionInfo{},
	}
	if si, err := h.buildTx(&reqBuild, txParams); err != nil {
		doBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}
