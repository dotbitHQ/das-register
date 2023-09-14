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
)

type ReqReverseRetract struct {
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	EvmChainId int64            `json:"evm_chain_id"`
}

type RespReverseRetract struct {
	SignInfo
}

func (h *HttpHandle) RpcReverseRetract(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqReverseRetract
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

	if err = h.doReverseRetract(&req[0], apiResp); err != nil {
		log.Error("doReverseRetract err:", err.Error())
	}
}

func (h *HttpHandle) ReverseRetract(ctx *gin.Context) {
	var (
		funcName = "ReverseRetract"
		clientIp = GetClientIp(ctx)
		req      ReqReverseRetract
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

	if err = h.doReverseRetract(&req, &apiResp); err != nil {
		log.Error("doReverseRetract err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseRetract(req *ReqReverseRetract, apiResp *api_code.ApiResp) error {
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

	if exi := h.rc.ApiLimitExist(req.ChainType, req.Address, common.DasActionRetractReverseRecord); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "The operation is too frequent")
		return fmt.Errorf("api limit: %d %s", req.ChainType, req.Address)
	}

	var resp RespReverseRetract

	list, err := h.dbDao.SearchReverseList(req.ChainType, req.Address)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search reverse list err")
		return fmt.Errorf("SearchReverseList err: %s", err.Error())
	}
	if len(list) == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeReverseNotExist, "reverse not exist")
		return fmt.Errorf("reverse not exist")
	}

	// das lock
	dasLock, dasType, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "format das lock err")
		return fmt.Errorf("HexToScript err: %s", err.Error())
	}

	// config cell check
	configCellBuilder, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get ConfigCellReverseResolution err")
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	feeCapacity, _ := configCellBuilder.RecordCommonFee()

	// build tx
	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionRetractReverseRecord
	reqBuild.Account = ""
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.EvmChainId = req.EvmChainId

	var retractParams RetractParams
	retractParams.DasLock = dasLock
	retractParams.DasType = dasType
	retractParams.ReverseList = list
	retractParams.FeeCapacity = feeCapacity
	for _, v := range list {
		reqBuild.Capacity = v.Capacity
	}

	txParams, err := h.buildRetractReverseRecordTx(&reqBuild, &retractParams)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildRetractReverseRecordTx err: %s", err.Error())
	}

	if si, err := h.buildTx(&reqBuild, txParams); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err ")
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

type RetractParams struct {
	DasLock     *types.Script
	DasType     *types.Script
	ReverseList []tables.TableReverseInfo
	FeeCapacity uint64
}

func (h *HttpHandle) buildRetractReverseRecordTx(req *reqBuildTx, p *RetractParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	totalCapacity := uint64(0)
	for _, v := range p.ReverseList {
		totalCapacity += v.Capacity
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: common.String2OutPointStruct(v.Outpoint),
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: totalCapacity - p.FeeCapacity,
		Lock:     p.DasLock,
		Type:     p.DasType,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionRetractReverseRecord, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// cell deps
	configCellReverse, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	balContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	reverseContract, err := core.GetDasContractInfo(common.DasContractNameReverseRecordCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	txParams.CellDeps = append(txParams.CellDeps,
		configCellReverse.ToCellDep(),
		balContract.ToCellDep(),
		reverseContract.ToCellDep(),
	)
	return &txParams, nil
}
