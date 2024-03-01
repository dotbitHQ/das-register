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
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync"
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doReverseDeclare(&req, &apiResp); err != nil {
		log.Error("doReverseDeclare err:", err.Error(), funcName, clientIp, ctx)
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
	liveCells, totalCapacity, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        dasLock,
		CapacityNeed:      needCapacity + feeCapacity,
		CapacityForChange: common.DasLockWithBalanceTypeOccupiedCkb,
		SearchOrder:       indexer.SearchOrderDesc,
	})
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
		return fmt.Errorf("GetBalanceCells err: %s", err.Error())
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
	Action      common.DasAction
	ChainType   common.ChainType `json:"chain_type"`
	Address     string           `json:"address"`
	Account     string           `json:"account"`
	Capacity    uint64           `json:"capacity"`
	EvmChainId  int64            `json:"evm_chain_id"`
	AuctionInfo AuctionInfo      `json:"auction_info"`
}

type DeclareParams struct {
	DasLock         *types.Script
	DasType         *types.Script
	LiveCells       []*indexer.LiveCell
	TotalCapacity   uint64
	DeclareCapacity uint64
	FeeCapacity     uint64
}

var balanceLock sync.Mutex

type ParamBalance struct {
	DasLock      *types.Script
	DasType      *types.Script
	NeedCapacity uint64
}

func (h *HttpHandle) GetBalanceCell(p *ParamBalance) (uint64, []*indexer.LiveCell, error) {
	if p.NeedCapacity == 0 {
		return 0, nil, nil
	}
	balanceLock.Lock()
	defer balanceLock.Unlock()

	liveCells, total, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        p.DasLock,
		CapacityNeed:      p.NeedCapacity,
		CapacityForChange: common.DasLockWithBalanceTypeMinCkbCapacity,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}

	var outpoints []string
	for _, v := range liveCells {
		outpoints = append(outpoints, common.OutPointStruct2String(v.OutPoint))
	}
	h.dasCache.AddOutPoint(outpoints)

	return total - p.NeedCapacity, liveCells, nil
}

func (h *HttpHandle) checkTxFee(txBuilder *txbuilder.DasTxBuilder, txParams *txbuilder.BuildTransactionParams, txFee uint64) (*txbuilder.DasTxBuilder, error) {
	if txFee >= common.UserCellTxFeeLimit {
		log.Info("Das pay tx fee :", txFee)
		change, liveBalanceCell, err := h.GetBalanceCell(&ParamBalance{
			DasLock:      h.serverScript,
			NeedCapacity: txFee,
		})
		if err != nil {
			return nil, fmt.Errorf("GetBalanceCell err %s", err.Error())
		}
		for _, v := range liveBalanceCell {
			txParams.Inputs = append(txParams.Inputs, &types.CellInput{
				PreviousOutput: v.OutPoint,
			})
		}
		// change balance_cell
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     h.serverScript,
		})

		txParams.OutputsData = append(txParams.OutputsData, []byte{})
		txBuilder = txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
		err = txBuilder.BuildTransaction(txParams)
		if err != nil {
			return nil, fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
		}
		log.Info("buildTx: das pay tx fee: ", txBuilder.TxString())
	}
	return txBuilder, nil
}
func deepCopy(src interface{}) (*txbuilder.BuildTransactionParams, error) {
	var params txbuilder.BuildTransactionParams
	jsonString, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal err %s", err.Error())
	}
	err = json.Unmarshal(jsonString, &params)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal err %s", err.Error())

	}
	return &params, nil
}
func (h *HttpHandle) buildTx(req *reqBuildTx, txParams *txbuilder.BuildTransactionParams) (*SignInfo, error) {
	rebuildTxParams, err := deepCopy(txParams)
	if err != nil {
		return nil, fmt.Errorf("deepCopy err %s", err.Error())
	}
	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return nil, fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}
	sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
	txFeeRate := config.Cfg.Server.TxTeeRate
	if txFeeRate == 0 {
		txFeeRate = 1
	}
	txFee := txFeeRate*sizeInBlock + 1000
	var skipGroups []int
	checkTxFeeParam := &core.CheckTxFeeParam{
		TxParams:      rebuildTxParams,
		DasCache:      h.dasCache,
		TxFee:         txFee,
		FeeLock:       h.serverScript,
		TxBuilderBase: h.txBuilderBase,
	}
	switch req.Action {
	case common.DasActionConfigSubAccountCustomScript:
		skipGroups = []int{1}
		changeCapacity := txBuilder.Transaction.Outputs[1].Capacity - txFee
		txBuilder.Transaction.Outputs[1].Capacity = changeCapacity
	case common.DasActionTransfer:
		//sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
		changeCapacity := txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity + common.OneCkb - txFee
		txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity = changeCapacity
		//if txFee >= common.UserCellTxFeeLimit {
		//	rebuildTxParams.Outputs[len(rebuildTxParams.Outputs)-1].Capacity += common.OneCkb
		//}

	case common.DasActionEditRecords, common.DasActionEditManager, common.DasActionTransferAccount:
		changeCapacity := txBuilder.Transaction.Outputs[0].Capacity - txFee
		txBuilder.Transaction.Outputs[0].Capacity = changeCapacity
		log.Info("buildTx:", req.Action, sizeInBlock, changeCapacity)

	case common.DasActionBidExpiredAccountAuction:
		accTx, err := h.dasCore.Client().GetTransaction(h.ctx, txParams.Inputs[0].PreviousOutput.TxHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
		}
		accLock := accTx.Transaction.Outputs[txParams.Inputs[0].PreviousOutput.Index].Lock

		dpTx, err := h.dasCore.Client().GetTransaction(h.ctx, txParams.Inputs[1].PreviousOutput.TxHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
		}
		dpLock := dpTx.Transaction.Outputs[txParams.Inputs[1].PreviousOutput.Index].Lock
		if !accLock.Equals(dpLock) {
			skipGroups = []int{0}
		}
		changeCapacity := txBuilder.Transaction.Outputs[0].Capacity - txFee
		txBuilder.Transaction.Outputs[0].Capacity = changeCapacity
		log.Info("buildTx:", req.Action, sizeInBlock, changeCapacity)
	}

	//txBuilder, err = h.checkTxFee(txBuilder, rebuildTxParams, txFee)
	//if err != nil {
	//	return nil, fmt.Errorf("checkTxFee err %s ", err.Error())
	//}
	newTxBuilder, err := h.dasCore.CheckTxFee(checkTxFeeParam)
	if err != nil {
		return nil, fmt.Errorf("CheckTxFee err %s ", err.Error())
	}
	if newTxBuilder != nil {
		txBuilder = newTxBuilder
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

	//dutch auction
	if req.Action == common.DasActionBidExpiredAccountAuction {
		sic.AuctionInfo = req.AuctionInfo
	}

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

func doBuildTxErr(err error, apiResp *api_code.ApiResp) {
	if err == nil || apiResp == nil {
		return
	}
	if strings.Contains(err.Error(), "not live") && strings.Contains(err.Error(), "addInputsForBaseTx err") {
		apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, "build tx err: "+err.Error())
	} else {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
	}
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
