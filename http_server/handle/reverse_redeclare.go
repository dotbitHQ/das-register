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
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type ReqReverseRedeclare struct {
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account"`
	EvmChainId int64            `json:"evm_chain_id"`
}

type RespReverseRedeclare struct {
	SignInfo
}

func (h *HttpHandle) RpcReverseRedeclare(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqReverseRedeclare
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

	if err = h.doReverseRedeclare(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doReverseRedeclare err:", err.Error())
	}
}

func (h *HttpHandle) ReverseRedeclare(ctx *gin.Context) {
	var (
		funcName = "ReverseRedeclare"
		clientIp = GetClientIp(ctx)
		req      ReqReverseRedeclare
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

	if err = h.doReverseRedeclare(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doReverseRedeclare err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseRedeclare(ctx context.Context, req *ReqReverseRedeclare, apiResp *api_code.ApiResp) error {
	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.ChainType,
		AddressNormal: req.Address,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address NormalToHex err")
		return fmt.Errorf("NormalToHex err: %s", err.Error())
	}
	req.Account = strings.ToLower(req.Account)
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if exi := h.rc.ApiLimitExist(req.ChainType, req.Address, common.DasActionRedeclareReverseRecord); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "The operation is too frequent")
		return fmt.Errorf("api limit: %d %s", req.ChainType, req.Address)
	}

	var resp RespReverseRedeclare

	// reverse check

	reverse, err := h.dbDao.SearchLatestReverse(req.ChainType, req.Address)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search reverse err")
			return fmt.Errorf("SearchLatestReverse err: %s", err.Error())
		}
	}
	if reverse.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeReverseNotExist, "not exist")
		return fmt.Errorf("reverse not exist: %d %s", req.ChainType, req.Address)
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

	if strings.EqualFold(acc.Account, reverse.Account) {
		apiResp.ApiRespErr(api_code.ApiCodeReverseAlreadyExist, "same account reverse")
		return fmt.Errorf("account not exist: %s", req.Account)
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
	reqBuild.Action = common.DasActionRedeclareReverseRecord
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.EvmChainId = req.EvmChainId

	var redeclareParams RedeclareParams
	redeclareParams.Reverse = &reverse
	redeclareParams.FeeCapacity = feeCapacity
	//redeclareParams.AccountInfo = &acc

	txParams, err := h.buildRedeclareReverseRecordTx(&reqBuild, &redeclareParams)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildRedeclareReverseRecordTx err: %s", err.Error())
	}

	if _, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

type RedeclareParams struct {
	//AccountInfo *tables.TableAccountInfo
	Reverse     *tables.TableReverseInfo
	FeeCapacity uint64
}

func (h *HttpHandle) buildRedeclareReverseRecordTx(req *reqBuildTx, p *RedeclareParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: common.String2OutPointStruct(p.Reverse.Outpoint),
	})

	// outputs
	res, err := h.dasCore.Client().GetTransaction(h.ctx, txParams.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	newCapacity := res.Transaction.Outputs[0].Capacity - p.FeeCapacity

	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: newCapacity,
		Lock:     res.Transaction.Outputs[0].Lock,
		Type:     res.Transaction.Outputs[0].Type,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte(req.Account))

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionRedeclareReverseRecord, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// accoun witness
	//resAcc, err := h.dasCore.Client().GetTransaction(h.ctx, common.String2OutPointStruct(p.AccountInfo.Outpoint).TxHash)
	//if err != nil {
	//	return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	//}
	//accMap, err := witness.AccountCellDataBuilderMapFromTx(resAcc.Transaction, common.DataTypeNew)
	//if err != nil {
	//	return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	//}
	//if acc, ok := accMap[req.Account]; !ok {
	//	return nil, fmt.Errorf("acc map not exist [%s]", req.Account)
	//} else {
	//	witnessAcc, _, err := acc.GenWitness(&witness.AccountCellParam{
	//		OldIndex: 0,
	//		Action:   common.DasActionRedeclareReverseRecord,
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
	accContract, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
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
		//accountDep,
		accContract.ToCellDep(),
		configCellReverse.ToCellDep(),
		balContract.ToCellDep(),
		reverseContract.ToCellDep(),
	)
	return &txParams, nil
}
