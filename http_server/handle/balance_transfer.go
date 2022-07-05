package handle

import (
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqBalanceTransfer struct {
	TransferAddress string           `json:"transfer_address"`
	ChainType       common.ChainType `json:"chain_type"`
	Address         string           `json:"address"`
	//EvmChainId      int64            `json:"evm_chain_id"`
}

type RespBalanceTransfer struct {
	SignInfo
}

func (h *HttpHandle) RpcBalanceTransfer(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqBalanceTransfer
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

	if err = h.doBalanceTransfer(&req[0], apiResp); err != nil {
		log.Error("doBalanceTransfer err:", err.Error())
	}
}

func (h *HttpHandle) BalanceTransfer(ctx *gin.Context) {
	var (
		funcName = "BalanceTransfer"
		clientIp = GetClientIp(ctx)
		req      ReqBalanceTransfer
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

	if err = h.doBalanceTransfer(&req, &apiResp); err != nil {
		log.Error("doBalanceTransfer err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalanceTransfer(req *ReqBalanceTransfer, apiResp *api_code.ApiResp) error {
	var resp RespBalanceTransfer

	if req.ChainType != common.ChainTypeEth {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "not support this chain type")
		return nil
	}
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

	// das-lock
	toLock, toType, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToScript err")
		return fmt.Errorf("HexToScript err: %s", err.Error())
	}

	fromLock, _, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(false),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToScript err")
		return fmt.Errorf("HexToScript err: %s", err.Error())
	}
	liveCells, totalAmount, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        fromLock,
		CapacityNeed:      0,
		CapacityForChange: 0,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	//liveCells, totalAmount, err := core.GetSatisfiedCapacityLiveCell(h.dasCore.Client(), nil, fromLock, nil, 0, 0)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "check balance err: "+err.Error())
		return fmt.Errorf("GetBalanceCells err: %s", err.Error())
	} else if totalAmount <= common.DasLockWithBalanceTypeOccupiedCkb { // 余额不足
		if req.TransferAddress != "" {
			parseAddress, err := address.Parse(req.TransferAddress)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address invalid")
				return fmt.Errorf("address.Parse err: %s [%s]", err.Error(), req.TransferAddress)
			}
			if parseAddress.Type == address.TypeFull {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address type invalid")
				return fmt.Errorf("full address: %s", req.TransferAddress)
			}
			fromLock = parseAddress.Script
			liveCells, totalAmount, err = h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
				DasCache:          nil,
				LockScript:        fromLock,
				CapacityNeed:      0,
				CapacityForChange: 0,
				SearchOrder:       indexer.SearchOrderDesc,
			})
			//liveCells, totalAmount, err = core.GetSatisfiedCapacityLiveCell(h.dasCore.Client(), nil, fromLock, nil, 0, 0)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, "check balance err: "+err.Error())
				return fmt.Errorf("GetBalanceCells err: %s", err.Error())
			}
			if totalAmount <= common.DasLockWithBalanceTypeOccupiedCkb { // 余额不足
				apiResp.ApiRespErr(api_code.ApiCodeInsufficientBalance, "insufficient balance")
				return nil
			}
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeInsufficientBalance, "insufficient balance")
			return nil
		}
	}

	feeAmount := uint64(1e4)
	transferAmount := totalAmount - feeAmount
	var reqBuild reqBuildTx
	reqBuild.Action = tables.DasActionTransferBalance
	reqBuild.Account = ""
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = transferAmount
	//reqBuild.EvmChainId = req.EvmChainId

	buildParam := balanceTransferParam{
		LiveCellList:   liveCells,
		InputsAmount:   totalAmount,
		TransferAmount: transferAmount,
		Fee:            feeAmount,
		FromScript:     fromLock,
		ToScript:       toLock,
		ToType:         toType,
	}

	txParams, err := h.buildBalanceTransferTx(&reqBuild, &buildParam)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildBalanceTransferTx err: %s", err.Error())
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

type balanceTransferParam struct {
	LiveCellList   []*indexer.LiveCell
	InputsAmount   uint64
	TransferAmount uint64
	Fee            uint64
	FromScript     *types.Script
	ToScript       *types.Script
	ToType         *types.Script
}

func (h *HttpHandle) buildBalanceTransferTx(req *reqBuildTx, p *balanceTransferParam) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	for _, v := range p.LiveCellList {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: &types.OutPoint{
				TxHash: v.OutPoint.TxHash,
				Index:  v.OutPoint.Index,
			},
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: p.TransferAmount,
		Lock:     p.ToScript,
		Type:     p.ToType,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})
	change := p.InputsAmount - p.TransferAmount - p.Fee
	if change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     p.FromScript,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(tables.DasActionTransferBalance, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	return &txParams, nil
}
