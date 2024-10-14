package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/internal"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type ReqEditRecords struct {
	core.ChainTypeAddress
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account" binding:"required"`
	EvmChainId int64            `json:"evm_chain_id"`
	RawParam   struct {
		Records []ReqRecord `json:"records"`
	} `json:"raw_param"`
}

type ReqRecord struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	TTL   string `json:"ttl"`
}

type RespEditRecords struct {
	SignInfo
}

func (h *HttpHandle) RpcEditRecords(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqEditRecords
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

	if err = h.doEditRecords(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doEditRecords err:", err.Error())
	}
}

func (h *HttpHandle) EditRecords(ctx *gin.Context) {
	var (
		funcName = "EditRecords"
		clientIp = GetClientIp(ctx)
		req      ReqEditRecords
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doEditRecords(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doEditRecords err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEditRecords(ctx context.Context, req *ReqEditRecords, apiResp *api_code.ApiResp) error {
	var resp RespEditRecords

	addressHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "account is invalid")
		return nil
	}
	if req.Address == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if exi := h.rc.AccountLimitExist(req.Account); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
		return fmt.Errorf("AccountActionLimitExist: %d %s %s", req.ChainType, req.Address, req.Account)
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if statusOk := acc.CheckStatus(); !statusOk {
		apiResp.ApiRespErr(api_code.ApiCodeAccountStatusNotNormal, "account status is not normal")
		return nil
	} else if acc.IsExpired() {
		apiResp.ApiRespErr(api_code.ApiCodeAccountIsExpired, "account is expired")
		return nil
	} else if req.ChainType != acc.ManagerChainType || !strings.EqualFold(req.Address, acc.Manager) {
		apiResp.ApiRespErr(api_code.ApiCodePermissionDenied, "edit records permission denied")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "not support sub account")
		return nil
	}

	// check records
	var p editRecordsParams
	p.account = &acc

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
		p.records = append(p.records, witness.Record{
			Key:   v.Key,
			Type:  v.Type,
			Label: v.Label,
			Value: v.Value,
			TTL:   uint32(ttl),
		})
	}

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionEditRecords
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = 0
	reqBuild.EvmChainId = req.EvmChainId

	records := witness.ConvertToCellRecords(p.records)
	recordsBys := records.AsSlice()
	log.Info("doEditRecords recordsBys:", len(recordsBys))
	if len(recordsBys) >= 5000 {
		apiResp.ApiRespErr(api_code.ApiCodeTooManyRecords, "too many records")
		return nil
	}

	txParams, err := h.buildEditRecordsTx(&reqBuild, &p)
	if err != nil {
		checkBuildTxErr(err, apiResp)
		return fmt.Errorf("buildEditManagerTx err: %s", err.Error())
	}
	if _, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
		doBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) doEditRecordsForDidCell(ctx context.Context, req *ReqEditRecords, apiResp *api_code.ApiResp, addrParse *address.ParsedAddress) error {
	var resp RespEditRecords

	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "account is invalid")
		return nil
	}
	args := common.Bytes2Hex(addrParse.Script.Args)
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	didAccount, err := h.dbDao.GetDidAccountByAccountId(accountId, args)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	} else if didAccount.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if didAccount.IsExpired() {
		apiResp.ApiRespErr(api_code.ApiCodeAccountIsExpired, "account is expired")
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

	didCellOutpoint := common.String2OutPointStruct(didAccount.Outpoint)
	txParams, err := txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
		DasCore:             h.dasCore,
		DasCache:            h.dasCache,
		Action:              common.DidCellActionEditRecords,
		DidCellOutPoint:     didCellOutpoint,
		AccountCellOutPoint: nil,
		EditRecords:         editRecords,
		EditOwnerLock:       nil,
		RenewYears:          0,
		NormalCellScript:    nil,
	})

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionEditRecords
	reqBuild.Account = req.Account
	reqBuild.ChainType = 0
	reqBuild.Address = req.KeyInfo.Key
	reqBuild.Capacity = 0
	reqBuild.EvmChainId = req.EvmChainId

	if _, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
		doBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func checkBuildTxErr(err error, apiResp *api_code.ApiResp) {
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), "not live") {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
	} else {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to build tx")
	}
}

type editRecordsParams struct {
	account *tables.TableAccountInfo
	records []witness.Record
}

func (h *HttpHandle) buildEditRecordsTx(req *reqBuildTx, p *editRecordsParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs account cell
	accOutPoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	// witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasActionEditRecords, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// witness account cell
	res, err := h.dasCore.Client().GetTransaction(h.ctx, accOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	builderMap, err := witness.AccountCellDataBuilderMapFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	}
	builder, ok := builderMap[req.Account]
	if !ok {
		return nil, fmt.Errorf("builderMap not exist account: %s", req.Account)
	}

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}

	accWitness, accData, err := builder.GenWitness(&witness.AccountCellParam{
		OldIndex:          0,
		NewIndex:          0,
		Action:            common.DasActionEditRecords,
		LastEditRecordsAt: timeCell.Timestamp(),
		Records:           p.records,
	})
	txParams.Witnesses = append(txParams.Witnesses, accWitness)
	accData = append(accData, res.Transaction.OutputsData[builder.Index][32:]...)

	// outputs account cell
	//builderConfigCell, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsAccount)
	//if err != nil {
	//	return nil, fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	//}
	//commonFee, err := builderConfigCell.AccountCommonFee()
	//if err != nil {
	//	return nil, fmt.Errorf("AccountCommonFee err: %s", err.Error())
	//}
	capacity := res.Transaction.Outputs[builder.Index].Capacity // - commonFee

	contractAcc, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	lockArgs, err := h.dasCore.Daf().HexToArgs(core.DasAddressHex{
		DasAlgorithmId: p.account.OwnerChainType.ToDasAlgorithmId(true),
		AddressHex:     p.account.Owner,
		IsMulti:        false,
		ChainType:      p.account.OwnerChainType,
	}, core.DasAddressHex{
		DasAlgorithmId: p.account.ManagerChainType.ToDasAlgorithmId(true),
		AddressHex:     p.account.Manager,
		IsMulti:        false,
		ChainType:      p.account.ManagerChainType,
	})
	if err != nil {
		return nil, fmt.Errorf("HexToArgs err: %s", err.Error())
	}

	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: capacity,
		Lock:     contractDas.ToScript(lockArgs),
		Type:     contractAcc.ToScript(nil),
	})
	txParams.OutputsData = append(txParams.OutputsData, accData)

	// cell deps
	heightCell, err := h.dasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}

	configCellAcc, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	configCellRecord, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsRecordNamespace)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		heightCell.ToCellDep(),
		timeCell.ToCellDep(),
		configCellAcc.ToCellDep(),
		configCellRecord.ToCellDep(),
	)

	return &txParams, nil
}
