package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/compatible"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
)

type ReqBalanceWithdraw struct {
	core.ChainTypeAddress
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	//ReceiverChainType common.ChainType `json:"receiver_chain_type"`
	ReceiverAddress string          `json:"receiver_address"`
	Amount          decimal.Decimal `json:"amount"`
	WithdrawAll     bool            `json:"withdraw_all"`
	EvmChainId      int64           `json:"evm_chain_id"`
}

type RespBalanceWithdraw struct {
	SignInfo
}

func (h *HttpHandle) RpcBalanceWithdraw(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqBalanceWithdraw
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

	if err = h.doBalanceWithdraw(&req[0], apiResp); err != nil {
		log.Error("doBalanceWithdraw err:", err.Error())
	}
}

func (h *HttpHandle) BalanceWithdraw(ctx *gin.Context) {
	var (
		funcName = "BalanceWithdraw"
		clientIp = GetClientIp(ctx)
		req      ReqBalanceWithdraw
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

	if err = h.doBalanceWithdraw(&req, &apiResp); err != nil {
		log.Error("doBalanceWithdraw err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalanceWithdraw(req *ReqBalanceWithdraw, apiResp *api_code.ApiResp) error {
	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}

	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	var resp RespBalanceWithdraw

	// amount
	if !req.Amount.BigInt().IsUint64() {
		apiResp.ApiRespErr(api_code.ApiCodeAmountInvalid, fmt.Sprintf("amount [%s] is invalid", req.Amount.String()))
		return nil
	}

	// receiver address
	parseAddress, err := address.Parse(req.ReceiverAddress)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeNotSupportAddress, "address invalid")
		return fmt.Errorf("address.Parse err: %s [%s]", err.Error(), req.ReceiverAddress)
	}

	if config.Cfg.Server.Net == common.DasNetTypeMainNet && parseAddress.Mode != address.Mainnet {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "testnet address is invalid")
		return fmt.Errorf("testnet address: %s", req.ReceiverAddress)
	} else if config.Cfg.Server.Net != common.DasNetTypeMainNet && parseAddress.Mode != address.Testnet {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "main net address is invalid")
		return fmt.Errorf("mainnet address: %s", req.ReceiverAddress)
	}

	var toTypeScript *types.Script
	if parseAddress.Type == address.TypeFull {
		dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "get das lock type id fail")
			return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		}
		if parseAddress.Script.CodeHash.Hex() == dasContract.ContractTypeId.Hex() { // das lock 712
			if !req.WithdrawAll && req.Amount.BigInt().Uint64() < common.DasLockWithBalanceTypeMinCkbCapacity {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("amount [%s] is invalid", req.Amount))
				return fmt.Errorf("amount not enough: %s", req.Amount)
			}
			balContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, "get balance type id fail")
				return fmt.Errorf("GetDasContractInfo: %s", err.Error())
			}
			toTypeScript = &types.Script{
				CodeHash: balContract.ContractTypeId,
				HashType: types.HashTypeType,
				Args:     nil,
			}
		}
	} else {
		if !req.WithdrawAll && req.Amount.BigInt().Uint64() < common.MinCellOccupiedCkb {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("amount [%d] is invalid", req.Amount.BigInt().Uint64()/1e8))
			return fmt.Errorf("amount not enough: %s", req.Amount)
		}
	}

	// das-lock
	dasLockScript, dasTypeScript, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get das lock err")
		return fmt.Errorf("HexToScript err: %s", err.Error())
	}

	// check balance
	//fee := uint64(1e6)
	withdrawAmount := req.Amount.BigInt().Uint64()
	allAmount := withdrawAmount //+ fee
	if req.WithdrawAll {
		allAmount = uint64(0)
	}
	liveCell, totalAmount, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        dasLockScript,
		CapacityNeed:      allAmount,
		CapacityForChange: common.DasLockWithBalanceTypeMinCkbCapacity + common.OneCkb,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		if err == core.ErrRejectedOutPoint {
			apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, core.ErrRejectedOutPoint.Error())
		} else if err == core.ErrInsufficientFunds {
			apiResp.ApiRespErr(api_code.ApiCodeInsufficientBalance, "insufficient balance")
		} else if err == core.ErrNotEnoughChange {
			apiResp.ApiRespErr(api_code.ApiCodeNotEnoughChange, "not enough change")
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "check balance err: "+err.Error())
		}
		return fmt.Errorf("GetBalanceCells err: %s [%s]", err.Error(), req.ReceiverAddress)
	}
	if req.WithdrawAll {
		withdrawAmount = totalAmount // - fee
	}

	// build tx
	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionWithdrawFromWallet
	reqBuild.Account = ""
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = withdrawAmount
	reqBuild.EvmChainId = req.EvmChainId

	buildParam := balanceWithdrawParam{
		LiveCellList:   liveCell,
		InputsAmount:   totalAmount,
		WithdrawAmount: withdrawAmount,
		Fee:            0,
		ToLockScript:   parseAddress.Script,
		ToTypeScript:   toTypeScript,
		FromLockScript: dasLockScript,
		FromTypeScript: dasTypeScript,
	}

	txParams, err := h.buildBalanceWithdrawTx(&reqBuild, &buildParam)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildBalanceWithdrawTx err: %s", err.Error())
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

type balanceWithdrawParam struct {
	LiveCellList   []*indexer.LiveCell
	InputsAmount   uint64
	WithdrawAmount uint64
	Fee            uint64
	ToLockScript   *types.Script
	ToTypeScript   *types.Script
	FromLockScript *types.Script
	FromTypeScript *types.Script
}

func (h *HttpHandle) buildBalanceWithdrawTx(req *reqBuildTx, p *balanceWithdrawParam) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	for _, v := range p.LiveCellList {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: p.WithdrawAmount,
		Lock:     p.ToLockScript,
		Type:     p.ToTypeScript,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	change := p.InputsAmount - p.WithdrawAmount // - p.Fee
	if change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     p.FromLockScript,
			Type:     p.FromTypeScript,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionWithdrawFromWallet, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	return &txParams, nil
}
