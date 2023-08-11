package handle

import (
	"das_register_server/config"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"das_register_server/internal"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqEditScript struct {
	core.ChainTypeAddress
	Account          string `json:"account"`
	CustomScriptArgs string `json:"custom_script_args"`
	EvmChainId       int64  `json:"evm_chain_id"`
}

type RespEditScript struct {
	SignInfo
}

func (h *HttpHandle) RpcEditScript(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqEditScript
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

	if err = h.doEditScript(&req[0], apiResp); err != nil {
		log.Error("doEditScript err:", err.Error())
	}
}

func (h *HttpHandle) EditScript(ctx *gin.Context) {
	var (
		funcName = "EditScript"
		clientIp = GetClientIp(ctx)
		req      ReqEditScript
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doEditScript(&req, &apiResp); err != nil {
		log.Error("doEditScript err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEditScript(req *ReqEditScript, apiResp *api_code.ApiResp) error {
	var resp RespEditScript

	hexAddress, err := req.FormatChainTypeAddress(h.dasCore.NetType(), true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
		return nil
	}
	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "account is nil")
		return nil
	}
	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	// check account
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "GetAccountInfoByAccountId err")
		return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
	} else if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.Status != tables.AccountStatusNormal {
		apiResp.ApiRespErr(api_code.ApiCodeAccountStatusNotNormal, "account on sale or auction")
		return nil
	} else if acc.IsExpired() {
		apiResp.ApiRespErr(api_code.ApiCodeAccountIsExpired, "account is expired")
		return nil
	} else if acc.OwnerChainType != hexAddress.ChainType || !strings.EqualFold(acc.Owner, hexAddress.AddressHex) {
		apiResp.ApiRespErr(api_code.ApiCodePermissionDenied, "owner permission required")
		return nil
	} else if acc.EnableSubAccount != tables.AccountEnableStatusOn {
		apiResp.ApiRespErr(api_code.ApiCodeSubAccountNotEnabled, "sub-account not enabled")
		return nil
	}
	// build tx
	reqBuild := reqBuildTx{
		Action:     common.DasActionConfigSubAccountCustomScript,
		ChainType:  hexAddress.ChainType,
		Address:    hexAddress.AddressHex,
		Account:    req.Account,
		Capacity:   0,
		EvmChainId: req.EvmChainId,
	}
	customScriptArgs := make([]byte, 33)
	if req.CustomScriptArgs != "" {
		tmpArgs := common.Hex2Bytes(req.CustomScriptArgs)
		if len(tmpArgs) != 33 {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "CustomScriptArgs err")
			return nil
		}
		customScriptArgs = tmpArgs
	}

	p := editScriptParams{
		acc:              &acc,
		customScriptArgs: customScriptArgs,
	}
	txParams, err := h.buildEditScript(&reqBuild, &p)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildEditManagerTx err: %s", err.Error())
	}
	if si, err := h.buildTx(&reqBuild, txParams); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

type editScriptParams struct {
	acc              *tables.TableAccountInfo
	customScriptArgs []byte
}

func (h *HttpHandle) buildEditScript(req *reqBuildTx, p *editScriptParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	accOutPoint := common.String2OutPointStruct(p.acc.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	contractSubAcc, err := core.GetDasContractInfo(common.DASContractNameSubAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	subAccountLiveCell, err := h.getSubAccountCell(contractSubAcc, p.acc.AccountId)
	if err != nil {
		return nil, fmt.Errorf("getSubAccountCell err: %s", err.Error())
	}
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: subAccountLiveCell.OutPoint,
	})

	// outputs account cell
	txAcc, err := h.dasCore.Client().GetTransaction(h.ctx, accOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: txAcc.Transaction.Outputs[accOutPoint.Index].Capacity,
		Lock:     txAcc.Transaction.Outputs[accOutPoint.Index].Lock,
		Type:     txAcc.Transaction.Outputs[accOutPoint.Index].Type,
	})
	txParams.OutputsData = append(txParams.OutputsData, txAcc.Transaction.OutputsData[accOutPoint.Index])

	// outputs sub-sccount cell
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: subAccountLiveCell.Output.Capacity,
		Lock:     subAccountLiveCell.Output.Lock,
		Type:     subAccountLiveCell.Output.Type,
	})
	subDataDetail := witness.ConvertSubAccountCellOutputData(subAccountLiveCell.OutputData)
	subDataDetail.CustomScriptArgs = p.customScriptArgs
	subAccountOutputData := witness.BuildSubAccountCellOutputData(subDataDetail)
	txParams.OutputsData = append(txParams.OutputsData, subAccountOutputData)

	// action witness
	actionWitness, err := witness.GenActionDataWitnessV2(common.DasActionConfigSubAccountCustomScript, nil, common.ParamOwner)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// account witness
	builderMap, err := witness.AccountIdCellDataBuilderFromTx(txAcc.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
	}
	builder, ok := builderMap[p.acc.AccountId]
	if !ok {
		return nil, fmt.Errorf("builderMap not exist account: %s", req.Account)
	}
	accWitness, _, err := builder.GenWitness(&witness.AccountCellParam{
		OldIndex: 0,
		NewIndex: 0,
		Action:   common.DasActionConfigSubAccountCustomScript,
	})
	txParams.Witnesses = append(txParams.Witnesses, accWitness)

	// cell deps
	heightCell, err := h.dasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}
	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	configCellAcc, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	configCellSubAcc, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsSubAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	contractAcc, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	contractDasLock, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	txParams.CellDeps = append(txParams.CellDeps,
		contractDasLock.ToCellDep(),
		contractAcc.ToCellDep(),
		contractSubAcc.ToCellDep(),
		heightCell.ToCellDep(),
		timeCell.ToCellDep(),
		configCellAcc.ToCellDep(),
		configCellSubAcc.ToCellDep(),
	)

	return &txParams, nil
}

func (h *HttpHandle) getSubAccountCell(contractSubAcc *core.DasContractInfo, parentAccountId string) (*indexer.LiveCell, error) {
	searchKey := indexer.SearchKey{
		Script:     contractSubAcc.ToScript(common.Hex2Bytes(parentAccountId)),
		ScriptType: indexer.ScriptTypeType,
		ArgsLen:    0,
		Filter:     nil,
	}
	subAccLiveCells, err := h.dasCore.Client().GetCells(h.ctx, &searchKey, indexer.SearchOrderDesc, 1, "")
	if err != nil {
		return nil, fmt.Errorf("GetCells err: %s", err.Error())
	}
	if subLen := len(subAccLiveCells.Objects); subLen != 1 {
		return nil, fmt.Errorf("sub account outpoint len: %d", subLen)
	}
	return subAccLiveCells.Objects[0], nil
}
