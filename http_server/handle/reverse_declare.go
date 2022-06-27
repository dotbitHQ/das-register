package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
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
	"gorm.io/gorm"
	"net/http"
)

type ReqReverseDeclare struct {
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account"`
	EvmChainId int64            `json:"evm_chain_id"`
}

type RespReverseDeclare struct {
	SignInfo
}

func (h *HttpHandle) RpcReverseDeclare(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqReverseDeclare
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

	if err = h.doReverseDeclare(&req[0], apiResp); err != nil {
		log.Error("doReverseDeclare err:", err.Error())
	}
}

func (h *HttpHandle) ReverseDeclare(ctx *gin.Context) {
	var (
		funcName = "ReverseDeclare"
		clientIp = GetClientIp(ctx)
		req      ReqReverseDeclare
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

	if err = h.doReverseDeclare(&req, &apiResp); err != nil {
		log.Error("doReverseDeclare err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseDeclare(req *ReqReverseDeclare, apiResp *api_code.ApiResp) error {
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

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if exi := h.rc.ApiLimitExist(req.ChainType, req.Address, common.DasActionDeclareReverseRecord); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "The operation is too frequent")
		return fmt.Errorf("api limit: %d %s", req.ChainType, req.Address)
	}

	var resp RespReverseDeclare

	// reverse check

	reverse, err := h.dbDao.SearchLatestReverse(req.ChainType, req.Address)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search reverse err")
			return fmt.Errorf("SearchLatestReverse err: %s", err.Error())
		}
	}
	if reverse.Id > 0 {
		apiResp.ApiRespErr(api_code.ApiCodeReverseAlreadyExist, "already exist")
		return fmt.Errorf("reverse already exist: %s", reverse.Account)
	}

	// account check

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
			return fmt.Errorf("SearchAccount err: %s", err.Error())
		}
	}
	if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return fmt.Errorf("account not exist: %s", req.Account)
	}

	// balance check
	dasLock, dasType, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "format das lock err")
		return fmt.Errorf("FormatAddressToDasLockScript err: %s", err.Error())
	}
	configCellBuilder, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get ConfigCellReverseResolution err")
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	baseCapacity, _ := configCellBuilder.RecordBasicCapacity()
	preparedFeeCapacity, _ := configCellBuilder.RecordPreparedFeeCapacity()
	needCapacity := baseCapacity + preparedFeeCapacity
	feeCapacity, _ := configCellBuilder.RecordCommonFee()
	liveCells, totalCapacity, err := core.GetSatisfiedCapacityLiveCell(h.dasCore.Client(), h.dasCache, dasLock, dasType, needCapacity+feeCapacity, common.DasLockWithBalanceTypeOccupiedCkb)
	if err != nil {
		if err == core.ErrRejectedOutPoint {
			apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, core.ErrRejectedOutPoint.Error())
		} else if err == core.ErrInsufficientFunds {
			apiResp.ApiRespErr(api_code.ApiCodeInsufficientBalance, core.ErrInsufficientFunds.Error())
		} else if err == core.ErrNotEnoughChange {
			apiResp.ApiRespErr(api_code.ApiCodeNotEnoughChange, core.ErrNotEnoughChange.Error())
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "get capacity err")
		}
		return fmt.Errorf("GetSatisfiedCapacityLiveCell err: %s", err.Error())
	}

	// build tx
	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionDeclareReverseRecord
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = needCapacity
	reqBuild.EvmChainId = req.EvmChainId

	var declareParams DeclareParams
	declareParams.DasLock = dasLock
	declareParams.DasType = dasType
	declareParams.LiveCells = liveCells
	declareParams.TotalCapacity = totalCapacity
	declareParams.DeclareCapacity = needCapacity
	declareParams.FeeCapacity = feeCapacity
	//declareParams.AccountInfo = &acc

	txParams, err := h.buildDeclareReverseRecordTx(&reqBuild, &declareParams)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildDeclareReverseRecordTx err: %s", err.Error())
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

type reqBuildTx struct {
	Action     common.DasAction
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account"`
	Capacity   uint64           `json:"capacity"`
	EvmChainId int64            `json:"evm_chain_id"`
}

type DeclareParams struct {
	DasLock         *types.Script
	DasType         *types.Script
	LiveCells       []*indexer.LiveCell
	TotalCapacity   uint64
	DeclareCapacity uint64
	FeeCapacity     uint64
	//AccountInfo     *tables.TableAccountInfo
}

func (h *HttpHandle) buildTx(req *reqBuildTx, txParams *txbuilder.BuildTransactionParams) (*SignInfo, error) {
	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return nil, fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}
	var skipGroups []int
	if req.Action == common.DasActionConfigSubAccountCustomScript {
		skipGroups = []int{1}
	}
	signList, err := txBuilder.GenerateDigestListFromTx(skipGroups)
	if err != nil {
		return nil, fmt.Errorf("txBuilder.GenerateDigestListFromTx err: %s", err.Error())
	}

	log.Info("buildTx:", txBuilder.TxString())

	var mmJsonObj *common.MMJsonObj
	switch req.Action {
	case common.DasActionConfigSubAccountCustomScript:
	default:
		mmJsonObj, err = txBuilder.BuildMMJsonObj(req.EvmChainId)
		if req.Action != tables.DasActionTransferBalance && err != nil {
			return nil, fmt.Errorf("txBuilder.BuildMMJsonObj err: %s", err.Error())
		} else {
			log.Info("BuildTx:", mmJsonObj.String())
		}
	}

	var sic SignInfoCache
	sic.Action = req.Action
	sic.ChainType = req.ChainType
	sic.Address = req.Address
	sic.Account = req.Account
	sic.Capacity = req.Capacity
	sic.BuilderTx = txBuilder.DasTxBuilderTransaction
	signKey := sic.SignKey()
	cacheStr := toolib.JsonString(&sic)
	if err = h.rc.SetSignTxCache(signKey, cacheStr); err != nil {
		return nil, fmt.Errorf("SetSignTxCache err: %s", err.Error())
	}

	var si SignInfo
	si.SignKey = signKey
	si.SignList = signList
	si.MMJson = mmJsonObj

	return &si, nil
}

func (h *HttpHandle) buildDeclareReverseRecordTx(req *reqBuildTx, p *DeclareParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	for _, v := range p.LiveCells {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// outputs
	reverseContract, err := core.GetDasContractInfo(common.DasContractNameReverseRecordCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: p.DeclareCapacity,
		Lock:     p.DasLock,
		Type:     reverseContract.ToScript(nil),
	})

	txParams.OutputsData = append(txParams.OutputsData, []byte(req.Account))

	// change
	changeCapacity := p.TotalCapacity - p.DeclareCapacity - p.FeeCapacity
	if changeCapacity > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: changeCapacity,
			Lock:     p.DasLock,
			Type:     p.DasType,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionDeclareReverseRecord, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// accoun witness
	//res, err := h.dasCore.Client().GetTransaction(h.ctx, common.String2OutPointStruct(p.AccountInfo.Outpoint).TxHash)
	//if err != nil {
	//	return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	//}
	//accMap, err := witness.AccountCellDataBuilderMapFromTx(res.Transaction, common.DataTypeNew)
	//if err != nil {
	//	return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	//}
	//if acc, ok := accMap[req.Account]; !ok {
	//	return nil, fmt.Errorf("acc map not exist [%s]", req.Account)
	//} else {
	//	witnessAcc, _, err := acc.GenWitness(&witness.AccountCellParam{
	//		OldIndex: 0,
	//		Action:   common.DasActionDeclareReverseRecord,
	//	})
	//	if err != nil {
	//		return nil, fmt.Errorf("acc.GenWitness err: %s", err.Error())
	//	}
	//	txParams.Witnesses = append(txParams.Witnesses, witnessAcc)
	//}
	//
	//accountDep := &types.CellDep{
	//	OutPoint: common.String2OutPointStruct(p.AccountInfo.Outpoint),
	//	DepType:  types.DepTypeCode,
	//}

	// cell deps
	balContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	accContract, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configCellReverse, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		//accountDep, // witness index
		accContract.ToCellDep(),
		configCellReverse.ToCellDep(),
		balContract.ToCellDep(),
		reverseContract.ToCellDep(),
	)
	return &txParams, nil
}
