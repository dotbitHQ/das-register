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

type ReqDidCellUpgradableList struct {
	core.ChainTypeAddress
	Pagination
	Keyword string `json:"keyword"`
}

type RespDidCellUpgradableList struct {
	Total int64               `json:"total"`
	List  []UpgradableAccount `json:"list"`
}

type UpgradableAccount struct {
	AccountId     string        `json:"account_id"`
	Account       string        `json:"account"`
	ExpiredAt     uint64        `json:"expired_at"`
	UpgradeStatus UpgradeStatus `json:"upgrade_status"`
}

type UpgradeStatus int

const (
	UpgradeStatusDefault UpgradeStatus = 0
	UpgradeStatusIng     UpgradeStatus = 1
)

func (h *HttpHandle) RpcDidCellUpgradableList(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellUpgradableList
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

	if err = h.doDidCellUpgradableList(&req[0], apiResp); err != nil {
		log.Error("doDidCellUpgradableList err:", err.Error())
	}
}

func (h *HttpHandle) DidCellUpgradableList(ctx *gin.Context) {
	var (
		funcName = "DidCellUpgradableList"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellUpgradableList
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

	if err = h.doDidCellUpgradableList(&req, &apiResp); err != nil {
		log.Error("doDidCellUpgradableList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellUpgradableList(req *ReqDidCellUpgradableList, apiResp *http_api.ApiResp) error {
	var resp RespDidCellUpgradableList
	resp.List = make([]UpgradableAccount, 0)
	req.Keyword = strings.ToLower(req.Keyword)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	} else if addrHex.DasAlgorithmId == common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}

	list, err := h.dbDao.GetAccountUpgradableList(addrHex.ChainType, addrHex.AddressHex, req.Keyword, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get account list")
		return fmt.Errorf("GetAccountUpgradableList err: %s", err.Error())
	}

	var accountIds []string
	for _, v := range list {
		acc := UpgradableAccount{
			AccountId:     v.AccountId,
			Account:       v.Account,
			ExpiredAt:     v.ExpiredAt,
			UpgradeStatus: UpgradeStatusDefault,
		}
		resp.List = append(resp.List, acc)
		accountIds = append(accountIds, v.AccountId)
	}

	count, err := h.dbDao.GetAccountUpgradableListTotal(addrHex.ChainType, addrHex.AddressHex, req.Keyword)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get account count")
		return fmt.Errorf("GetAccountUpgradableListTotal err: %s", err.Error())
	}
	resp.Total = count

	// status
	orders, err := h.dbDao.GetUpgradeOrder(accountIds)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get upgrade order")
		return fmt.Errorf("GetUpgradeOrder err: %s", err.Error())
	}
	var upgradingMap = make(map[string]struct{})
	for _, v := range orders {
		upgradingMap[v.AccountId] = struct{}{}
	}

	for i, v := range resp.List {
		if _, ok := upgradingMap[v.AccountId]; ok {
			resp.List[i].UpgradeStatus = UpgradeStatusIng
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
