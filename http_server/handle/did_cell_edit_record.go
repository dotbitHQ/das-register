package handle

import (
	"bytes"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type ReqDidCellEditRecord struct {
	core.ChainTypeAddress
	Account  string `json:"account"`
	RawParam struct {
		Records []ReqRecord `json:"records"`
	} `json:"raw_param"`
}

type RespDidCellEditRecord struct {
	SignInfo
}

func (h *HttpHandle) RpcDidCellEditRecord(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellEditRecord
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

	if err = h.doDidCellEditRecord(&req[0], apiResp); err != nil {
		log.Error("doDidCellEditRecord err:", err.Error())
	}
}

func (h *HttpHandle) DidCellEditRecord(ctx *gin.Context) {
	var (
		funcName = "DidCellEditRecord"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellEditRecord
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

	if err = h.doDidCellEditRecord(&req, &apiResp); err != nil {
		log.Error("doDidCellEditRecord err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellEditRecord(req *ReqDidCellEditRecord, apiResp *http_api.ApiResp) error {
	var resp RespDidCellEditRecord

	req.Account = strings.ToLower(req.Account)
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	}
	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "account is invalid")
		return nil
	}
	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	var didCellOutPoint *types.OutPoint
	var accountCellOutPoint *types.OutPoint
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))

	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.IsExpired() {
		apiResp.ApiRespErr(api_code.ApiCodeAccountIsExpired, "account is expired")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "not support sub account")
		return nil
	}
	if acc.Status == tables.AccountStatusNormal {
		accountCellOutPoint = acc.GetOutpoint()
		if addrHex.ChainType != acc.ManagerChainType || !strings.EqualFold(addrHex.AddressHex, acc.Manager) {
			apiResp.ApiRespErr(api_code.ApiCodePermissionDenied, "edit records permission denied")
			return nil
		}
	} else if acc.Status == tables.AccountStatusOnUpgrade {
		didAccount, err := h.dbDao.GetDidAccountByAccountIdWithoutArgs(accountId)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did cell info")
			return fmt.Errorf("GetDidAccountByAccountId err: %s", err.Error())
		} else if didAccount.Id == 0 {
			apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "did cell not exist")
			return nil
		} else if addrHex.ParsedAddress == nil || bytes.Compare(common.Hex2Bytes(didAccount.Args), addrHex.ParsedAddress.Script.Args) != 0 {
			apiResp.ApiRespErr(http_api.ApiCodeNoAccountPermissions, "edit record permission denied")
			return nil
		}
		didCellOutPoint = didAccount.GetOutpoint()
	} else {
		apiResp.ApiRespErr(api_code.ApiCodeAccountStatusNotNormal, "account status is not normal")
		return nil
	}

	// check records
	builder, err := h.dasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsRecordNamespace)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
	}
	log.Info("ConfigCellRecordKeys:", builder.ConfigCellRecordKeys)
	var mapRecordKey = make(map[string]struct{})
	for _, v := range builder.ConfigCellRecordKeys {
		mapRecordKey[v] = struct{}{}
	}

	var editRecords []witness.Record
	for _, v := range req.RawParam.Records {
		record := fmt.Sprintf("%s.%s", v.Type, v.Key)
		if v.Type == "custom_key" { // (^[0-9a-z_]+$)
			if ok, _ := regexp.MatchString("^[0-9a-z_]+$", v.Key); !ok {
				apiResp.ApiRespErr(api_code.ApiCodeRecordInvalid, fmt.Sprintf("record [%s] is invalid", record))
				return nil
			}
		} else if v.Type == "address" {
			if ok, _ := regexp.MatchString("^(0|[1-9][0-9]*)$", v.Key); !ok {
				if _, ok2 := mapRecordKey[record]; !ok2 {
					apiResp.ApiRespErr(api_code.ApiCodeRecordInvalid, fmt.Sprintf("record [%s] is invalid", record))
					return nil
				}
			}
		} else if _, ok := mapRecordKey[record]; !ok {
			apiResp.ApiRespErr(api_code.ApiCodeRecordInvalid, fmt.Sprintf("record [%s] is invalid", record))
			return nil
		}
		ttl, err := strconv.ParseInt(v.TTL, 10, 64)
		if err != nil {
			ttl = 300
		}
		editRecords = append(editRecords, witness.Record{
			Key:   v.Key,
			Type:  v.Type,
			Label: v.Label,
			Value: v.Value,
			TTL:   uint32(ttl),
		})
	}

	txParams, err := txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
		DasCore:             h.dasCore,
		DasCache:            h.dasCache,
		Action:              common.DidCellActionEditRecords,
		DidCellOutPoint:     didCellOutPoint,
		AccountCellOutPoint: accountCellOutPoint,
		EditRecords:         editRecords,
	})

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionEditRecords
	reqBuild.Account = req.Account
	reqBuild.ChainType = addrHex.ChainType
	reqBuild.Address = addrHex.AddressHex
	reqBuild.EvmChainId = req.GetChainId(config.Cfg.Server.Net)

	records := witness.ConvertToCellRecords(editRecords)
	recordsBys := records.AsSlice()
	log.Info("doEditRecords recordsBys:", len(recordsBys))
	if len(recordsBys) >= 5000 {
		apiResp.ApiRespErr(http_api.ApiCodeTooManyRecords, "too many records")
		return nil
	}

	if didCellTx, si, err := h.buildTx(&reqBuild, txParams); err != nil {
		doBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
		if acc.Status == tables.AccountStatusOnUpgrade {
			resp.SignInfo.CKBTx = didCellTx
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
