package handle

import (
	"das_register_server/config"
	"das_register_server/tables"
	"das_register_server/txtool"
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
	UpgradeHash   string        `json:"upgrade_hash"`
}

type UpgradeStatus int

const (
	UpgradeStatusDefault        UpgradeStatus = 0
	UpgradeStatusConfirmPayment UpgradeStatus = 1
	UpgradeStatusIng            UpgradeStatus = 2
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
	var upgradingMap = make(map[string]UpgradeStatus)
	var upgradingTxMap = make(map[string]string)
	for _, v := range orders {
		switch v.PayStatus {
		case tables.TxStatusDefault:
			upgradingMap[v.AccountId] = UpgradeStatusConfirmPayment
		case tables.TxStatusSending, tables.TxStatusOk:
			upgradingMap[v.AccountId] = UpgradeStatusIng
		}

		didCellTxStr, err := h.rc.GetCache(v.OrderId)
		if err != nil {
			log.Error("GetCache err: %s", err.Error())
			continue
		}
		var txCache txtool.DidCellTxCache
		if err := json.Unmarshal([]byte(didCellTxStr), &txCache); err != nil {
			log.Error("txtool.DidCellTxCache json.Unmarshal err: %s", err.Error())
			continue
		}
		txHash, err := txCache.BuilderTx.Transaction.ComputeHash()
		if err != nil {
			log.Error("txCache.BuilderTx.Transaction.ComputeHash err: %s", err.Error())
			continue
		}
		upgradingTxMap[v.AccountId] = txHash.Hex()
	}

	for i, v := range resp.List {
		if s, ok := upgradingMap[v.AccountId]; ok {
			resp.List[i].UpgradeStatus = s
		}
		if txHash, ok := upgradingTxMap[v.AccountId]; ok {
			resp.List[i].UpgradeHash = txHash
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
