package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/txbuilder"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqBalanceDeposit struct {
	FromCkbAddress string `json:"from_ckb_address"`
	ToCkbAddress   string `json:"to_ckb_address"`
	Amount         uint64 `json:"amount"`
}

type RespBalanceDeposit struct {
	SignInfo
}

func (h *HttpHandle) RpcBalanceDeposit(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqBalanceDeposit
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

	if err = h.doBalanceDeposit(&req[0], apiResp); err != nil {
		log.Error("doBalanceDeposit err:", err.Error())
	}
}

func (h *HttpHandle) BalanceDeposit(ctx *gin.Context) {
	var (
		funcName = "BalanceDeposit"
		clientIp = GetClientIp(ctx)
		req      ReqBalanceDeposit
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

	if err = h.doBalanceDeposit(&req, &apiResp); err != nil {
		log.Error("doBalanceDeposit err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalanceDeposit(req *ReqBalanceDeposit, apiResp *api_code.ApiResp) error {
	var resp RespBalanceDeposit

	fromAddress, err := address.Parse(req.FromCkbAddress)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("from address.Parse err: %s", err.Error())
	}
	toAddress, err := address.Parse(req.ToCkbAddress)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("to address.Parse err: %s", err.Error())
	}
	if config.Cfg.Server.Net == common.DasNetTypeMainNet {
		if fromAddress.Mode != address.Mainnet || toAddress.Mode != address.Mainnet {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "testnet address")
			return nil
		}
	} else {
		if fromAddress.Mode != address.Testnet || toAddress.Mode != address.Testnet {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "mainnet address")
			return nil
		}
	}

	dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	balanceContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	var fromTypeScript, toTypeScript *types.Script
	if dasContract.IsSameTypeId(fromAddress.Script.CodeHash) {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "from a das address")
		return nil
	}
	if dasContract.IsSameTypeId(toAddress.Script.CodeHash) {
		oID, _, _, _, _, _ := core.FormatDasLockToHexAddress(toAddress.Script.Args)
		if oID == common.DasAlgorithmIdEth712 {
			toTypeScript = balanceContract.ToScript(nil)
		}
	}

	fee := uint64(1e4)
	liveCell, total, err := core.GetSatisfiedCapacityLiveCell(h.dasCore.Client(), h.dasCache, fromAddress.Script, fromTypeScript, req.Amount+fee, common.DasLockWithBalanceTypeOccupiedCkb)
	if err != nil {
		if err == core.ErrRejectedOutPoint {
			apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, err.Error())
			return nil
		} else if err == core.ErrNotEnoughChange {
			apiResp.ApiRespErr(api_code.ApiCodeNotEnoughChange, err.Error())
			return nil
		} else if err == core.ErrInsufficientFunds {
			apiResp.ApiRespErr(api_code.ApiCodeInsufficientBalance, err.Error())
			return nil
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("GetSatisfiedCapacityLiveCell err: %s", err.Error())
		}
	}

	action := tables.DasActionTransferBalance
	txParams, err := h.buildBalanceDepositTx(&paramBuildBalanceDepositTx{
		liveCellList:   liveCell,
		total:          total,
		amount:         req.Amount,
		fee:            fee,
		action:         action,
		toLockScript:   toAddress.Script,
		toTypeScript:   toTypeScript,
		fromLockScript: fromAddress.Script,
		fromTypeScript: fromTypeScript,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildBalanceDepositTx err: %s", err.Error())
	}

	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "BuildTransaction err: "+err.Error())
		return fmt.Errorf("txBuilder.BuildTransaction err: %s", err.Error())
	}
	txBuilder.ServerSignGroup = []int{}

	signList, err := txBuilder.GenerateDigestListFromTx([]int{})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "GenerateDigestListFromTx err: "+err.Error())
		return fmt.Errorf("txBuilder.GenerateDigestListFromTx err: %s", err.Error())
	}

	log.Info("buildTx:", txBuilder.TxString())

	var sic SignInfoCache
	sic.Action = action
	sic.ChainType = common.ChainTypeCkb
	sic.Address = common.Bytes2Hex(fromAddress.Script.Args)
	sic.BuilderTx = txBuilder.DasTxBuilderTransaction
	signKey := sic.SignKey()
	cacheStr := toolib.JsonString(&sic)
	if err = h.rc.SetSignTxCache(signKey, cacheStr); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "SetSignTxCache err: "+err.Error())
		return fmt.Errorf("SetSignTxCache err: %s", err.Error())
	}

	var si SignInfo
	si.SignKey = signKey
	si.SignList = signList

	resp.SignInfo = si

	apiResp.ApiRespOK(resp)
	return nil
}

type paramBuildBalanceDepositTx struct {
	liveCellList   []*indexer.LiveCell
	total          uint64
	amount         uint64
	fee            uint64
	action         string
	toLockScript   *types.Script
	toTypeScript   *types.Script
	fromLockScript *types.Script
	fromTypeScript *types.Script
}

func (h *HttpHandle) buildBalanceDepositTx(p *paramBuildBalanceDepositTx) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	for _, v := range p.liveCellList {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: &types.OutPoint{
				TxHash: v.OutPoint.TxHash,
				Index:  v.OutPoint.Index,
			},
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: p.amount,
		Lock:     p.toLockScript,
		Type:     p.toTypeScript,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	// change
	if change := p.total - p.amount - p.fee; change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     p.fromLockScript,
			Type:     p.fromTypeScript,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(p.action, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)
	return &txParams, nil
}
