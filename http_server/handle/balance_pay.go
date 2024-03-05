package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/compatible"
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
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqBalancePay struct {
	core.ChainTypeAddress
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	OrderId    string           `json:"order_id"`
	EvmChainId int64            `json:"evm_chain_id"`
}

type RespBalancePay struct {
	SignInfo
}

func (h *HttpHandle) RpcBalancePay(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqBalancePay
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

	if err = h.doBalancePay(&req[0], apiResp); err != nil {
		log.Error("doBalancePay err:", err.Error())
	}
}

func (h *HttpHandle) BalancePay(ctx *gin.Context) {
	var (
		funcName = "BalancePay"
		clientIp = GetClientIp(ctx)
		req      ReqBalancePay
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

	if err = h.doBalancePay(&req, &apiResp); err != nil {
		log.Error("doBalancePay err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalancePay(req *ReqBalancePay, apiResp *api_code.ApiResp) error {
	//addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
	//	ChainType:     req.ChainType,
	//	AddressNormal: req.Address,
	//	Is712:         true,
	//})
	//if err != nil {
	//	apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address NormalToHex err")
	//	return fmt.Errorf("NormalToHex err: %s", err.Error())
	//}
	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex
	var resp RespBalancePay

	if req.OrderId == "" || req.Address == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}

	// check order
	order, err := h.dbDao.GetOrderByOrderId(req.OrderId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "get order fail")
		return fmt.Errorf("GetOrderByOrderId err: %s [%s]", err.Error(), req.OrderId)
	} else if order.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeOrderNotExist, "order not exist")
		return nil
	}

	if order.PayStatus != tables.TxStatusDefault {
		apiResp.ApiRespErr(api_code.ApiCodeOrderPaid, "order paid")
		return nil
	} else if order.PayTokenId != tables.TokenIdDas {
		apiResp.ApiRespErr(api_code.ApiCodePayTypeInvalid, fmt.Sprintf("pay token id [%s] invalid", order.PayTokenId))
		return nil
	}

	// check balance
	dasLock, dasType, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("HexToScript err: %s", err.Error())
	}
	fee := common.OneCkb
	needCapacity := order.PayAmount.BigInt().Uint64() //+ fee
	liveCells, totalCapacity, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        dasLock,
		CapacityNeed:      needCapacity + fee,
		CapacityForChange: common.DasLockWithBalanceTypeMinCkbCapacity,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		checkBalanceErr(err, apiResp)
		return nil
	}

	// check pay address
	beneficiaryAddress := ""
	addr := config.GetUnipayAddress(order.PayTokenId)
	if addr == "" {
		apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("not supported [%s]", order.PayTokenId))
		return nil
	} else {
		beneficiaryAddress = addr
	}
	parseAddress, err := address.Parse(beneficiaryAddress)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}

	// build tx
	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionTransfer
	reqBuild.Account = order.Account
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = needCapacity
	reqBuild.EvmChainId = req.EvmChainId

	p := balancePayParams{
		orderId:        req.OrderId,
		liveCells:      liveCells,
		totalCapacity:  totalCapacity,
		payCapacity:    needCapacity,
		feeCapacity:    fee,
		fromLockScript: dasLock,
		fromTypeScript: dasType,
		toLockScript:   parseAddress.Script,
	}

	txParams, err := h.buildBalancePayTx(&p)
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

type balancePayParams struct {
	orderId        string
	liveCells      []*indexer.LiveCell
	totalCapacity  uint64
	payCapacity    uint64
	feeCapacity    uint64
	fromLockScript *types.Script
	fromTypeScript *types.Script
	toLockScript   *types.Script
}

func (h *HttpHandle) buildBalancePayTx(p *balancePayParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	for _, v := range p.liveCells {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: p.payCapacity,
		Lock:     p.toLockScript,
		Type:     nil,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte(p.orderId))

	// change
	if change := p.totalCapacity - p.payCapacity; change > 0 {
		changeList, err := core.SplitOutputCell(change, 2000*common.OneCkb, 5, p.fromLockScript, p.fromTypeScript)
		if err != nil {
			return nil, fmt.Errorf("SplitOutputCell err: %s", err.Error())
		}
		for _, cell := range changeList {
			txParams.Outputs = append(txParams.Outputs, cell)
			txParams.OutputsData = append(txParams.OutputsData, []byte{})
		}

		//txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		//	Capacity: change,
		//	Lock:     p.fromLockScript,
		//	Type:     p.fromTypeScript,
		//})
		//txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionTransfer, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	return &txParams, nil
}
