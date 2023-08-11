package handle

import (
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
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqEditManager struct {
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account"`
	EvmChainId int64            `json:"evm_chain_id"`
	RawParam   struct {
		ManagerChainType common.ChainType `json:"manager_chain_type"`
		ManagerAddress   string           `json:"manager_address"`
	} `json:"raw_param"`
}

type RespEditManager struct {
	SignInfo
}

func (h *HttpHandle) RpcEditManager(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqEditManager
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

	if err = h.doEditManager(&req[0], apiResp); err != nil {
		log.Error("doEditManager err:", err.Error())
	}
}

func (h *HttpHandle) EditManager(ctx *gin.Context) {
	var (
		funcName = "EditManager"
		clientIp = GetClientIp(ctx)
		req      ReqEditManager
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

	if err = h.doEditManager(&req, &apiResp); err != nil {
		log.Error("doEditManager err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEditManager(req *ReqEditManager, apiResp *api_code.ApiResp) error {
	var resp RespEditManager
	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.ChainType,
		AddressNormal: req.Address,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address NormalToHex err")
		return fmt.Errorf("NormalToHex err: %s", err.Error())
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	managerHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.RawParam.ManagerChainType,
		AddressNormal: req.RawParam.ManagerAddress,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "manager address NormalToHex err")
		return fmt.Errorf("manager NormalToHex err: %s", err.Error())
	}
	req.RawParam.ManagerChainType, req.RawParam.ManagerAddress = managerHex.ChainType, managerHex.AddressHex
	if !checkChainType(req.RawParam.ManagerChainType) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("chain type [%d] inavlid", req.RawParam.ManagerChainType))
		return nil
	}
	//
	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "account is invalid")
		return nil
	}
	if req.Address == "" || req.RawParam.ManagerAddress == "" {
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
	} else if req.ChainType != acc.OwnerChainType || !strings.EqualFold(req.Address, acc.Owner) {
		apiResp.ApiRespErr(api_code.ApiCodePermissionDenied, "edit manager permission denied")
		return nil
	} else if req.RawParam.ManagerChainType == acc.ManagerChainType && strings.EqualFold(req.RawParam.ManagerAddress, acc.Manager) {
		apiResp.ApiRespErr(api_code.ApiCodeSameLock, "same address")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "not support sub account")
		return nil
	}

	if (req.ChainType == common.ChainTypeMixin && req.RawParam.ManagerChainType != common.ChainTypeMixin) ||
		(req.ChainType != common.ChainTypeMixin && req.RawParam.ManagerChainType == common.ChainTypeMixin) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "ChainType is invalid")
		return nil
	}

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionEditManager
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = 0
	reqBuild.EvmChainId = req.EvmChainId

	var p editManagerParams
	p.account = &acc
	p.managerChainType = req.RawParam.ManagerChainType
	p.managerAddress = req.RawParam.ManagerAddress
	txParams, err := h.buildEditManagerTx(&reqBuild, &p)
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

type editManagerParams struct {
	account          *tables.TableAccountInfo
	managerChainType common.ChainType
	managerAddress   string
}

func (h *HttpHandle) buildEditManagerTx(req *reqBuildTx, p *editManagerParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs account cell
	accOutPoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	// witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasActionEditManager, nil)
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
		Action:            common.DasActionEditManager,
		LastEditManagerAt: timeCell.Timestamp(),
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
	capacity := res.Transaction.Outputs[builder.Index].Capacity //- commonFee

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
		DasAlgorithmId: p.managerChainType.ToDasAlgorithmId(true),
		AddressHex:     p.managerAddress,
		IsMulti:        false,
		ChainType:      p.managerChainType,
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

	txParams.CellDeps = append(txParams.CellDeps,
		heightCell.ToCellDep(),
		timeCell.ToCellDep(),
		configCellAcc.ToCellDep(),
	)

	return &txParams, nil
}
