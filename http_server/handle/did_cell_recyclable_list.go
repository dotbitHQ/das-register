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

type ReqDidCellRecyclableList struct {
	core.ChainTypeAddress
	Pagination
	Keyword string `json:"keyword"`
}

type RespDidCellRecyclableList struct {
	Total int64               `json:"total"`
	List  []DidCellRecyclable `json:"list"`
}

type DidCellRecyclable struct {
	Account       string        `json:"account"`
	ExpiredAt     uint64        `json:"expired_at"`
	RecycleStatus RecycleStatus `json:"recycle_status"`
	RecycleHash   string        `json:"recycle_hash"`
}

type RecycleStatus int

const (
	RecycleStatusDefault RecycleStatus = 0
	RecycleStatusIng     RecycleStatus = 1
)

func (h *HttpHandle) RpcDidCellRecyclableList(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellRecyclableList
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

	if err = h.doDidCellRecyclableList(&req[0], apiResp); err != nil {
		log.Error("doDidCellRecyclableList err:", err.Error())
	}
}

func (h *HttpHandle) DidCellRecyclableList(ctx *gin.Context) {
	var (
		funcName = "DidCellRecyclableList"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellRecyclableList
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

	if err = h.doDidCellRecyclableList(&req, &apiResp); err != nil {
		log.Error("doDidCellRecyclableList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellRecyclableList(req *ReqDidCellRecyclableList, apiResp *http_api.ApiResp) error {
	var resp RespDidCellRecyclableList
	resp.List = make([]DidCellRecyclable, 0)

	req.Keyword = strings.ToLower(req.Keyword)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	} else if addrHex.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return nil
	}
	// expireAt
	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "failed to get time cell")
		return fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	builderConfigCell, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsAccount)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "failed to get config cell")
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	expirationGracePeriod, err := builderConfigCell.ExpirationGracePeriod()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "ExpirationGracePeriod err")
		return fmt.Errorf("ExpirationGracePeriod err: %s", err.Error())
	}
	expiredAt := uint64(timeCell.Timestamp()) - uint64(expirationGracePeriod)
	log.Info("doDidCellRecyclableList:", expiredAt, timeCell.Timestamp(), expirationGracePeriod)

	//
	args := common.Bytes2Hex(addrHex.ParsedAddress.Script.Args)
	list, err := h.dbDao.GetDidCellRecyclableList(args, req.Keyword, req.GetLimit(), req.GetOffset(), expiredAt)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account list")
		return fmt.Errorf("GetDidCellRecyclableList err: %s", err.Error())
	}

	var accounts []string
	for _, v := range list {
		resp.List = append(resp.List, DidCellRecyclable{
			Account:   v.Account,
			ExpiredAt: v.ExpiredAt,
		})
		accounts = append(accounts, v.Account)
	}

	count, err := h.dbDao.GetDidCellRecyclableListTotal(args, req.Keyword, expiredAt)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did account count")
		return fmt.Errorf("GetDidAccountListTotal err: %s", err.Error())
	}
	resp.Total = count

	// recycle ing
	pendingList, err := h.dbDao.GetRecyclingByAddr(addrHex.ChainType, addrHex.AddressHex, accounts)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get recycling list")
		return fmt.Errorf("GetRecyclingByAddr err: %s", err.Error())
	}
	var pendingMap = make(map[string]string)
	for _, v := range pendingList {
		outpoint := common.String2OutPointStruct(v.Outpoint)
		pendingMap[v.Account] = outpoint.TxHash.Hex()
	}

	for i, v := range resp.List {
		if txHash, ok := pendingMap[v.Account]; ok {
			resp.List[i].RecycleHash = txHash
			resp.List[i].RecycleStatus = RecycleStatusIng
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
