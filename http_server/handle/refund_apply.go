package handle

import (
	"das_register_server/config"
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

// curl -X POST http://127.0.0.1:8119/v1/refund/apply -d'{"is_refund":true,"block_number":2623541}'

type ReqRefundApply struct {
	IsRefund    bool   `json:"is_refund"`
	IsAll       bool   `json:"is_all"`
	BlockNumber uint64 `json:"block_number"`
}

type RespRefundApply struct {
}

func (h *HttpHandle) RpcRefundApply(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqRefundApply
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

	if err = h.doRefundApply(&req[0], apiResp); err != nil {
		log.Error("doRefundApply err:", err.Error())
	}
}

func (h *HttpHandle) RefundApply(ctx *gin.Context) {
	var (
		funcName = "RefundApply"
		clientIp = GetClientIp(ctx)
		req      ReqRefundApply
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

	if err = h.doRefundApply(&req, &apiResp); err != nil {
		log.Error("doRefundApply err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doRefundApply(req *ReqRefundApply, apiResp *api_code.ApiResp) error {
	var resp RespRefundApply

	parseAddress, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}

	applyContract, err := core.GetDasContractInfo(common.DasContractNameApplyRegisterCellType)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	num, err := h.dasCore.Client().GetTipBlockNumber(h.ctx)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetTipBlockNumber err: %s", err.Error())
	}
	// search need refund apply tx
	key := indexer.SearchKey{
		Script:     parseAddress.Script,
		ScriptType: indexer.ScriptTypeLock,
		Filter: &indexer.CellsFilter{
			Script:     applyContract.ToScript(nil),
			BlockRange: &[2]uint64{req.BlockNumber, num},
		},
	}
	cells, err := h.dasCore.Client().GetCells(h.ctx, &key, indexer.SearchOrderAsc, 100, "")
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetCells err: %s", err.Error())
	}

	var refundList []*indexer.LiveCell
	diffBlockNumber := uint64(24 * 60 * 60 / 3)
	log.Info("cells.Objects:", len(cells.Objects))
	for i, v := range cells.Objects {
		if num-v.BlockNumber > diffBlockNumber {
			log.Info(v.OutPoint.TxHash.Hex(), v.BlockNumber, num-v.BlockNumber)
			refundList = append(refundList, cells.Objects[i])
		}
	}
	log.Info("refundList:", len(refundList))
	if req.IsRefund {
		for i, _ := range refundList {
			txParams, err := h.buildRefundApplyTx(refundList[i])
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return fmt.Errorf("buildRefundApplyTx err: %s", err.Error())
			}
			txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, nil)
			if err := txBuilder.BuildTransaction(txParams); err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return fmt.Errorf("BuildTransaction err: %s", err.Error())
			}
			log.Info("ServerSignGroup:", txBuilder.ServerSignGroup)
			if hash, err := txBuilder.SendTransaction(); err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return fmt.Errorf("SendTransaction err: %s", err.Error())
			} else {
				log.Info("SendTransaction ok:", hash)
				if !req.IsAll {
					break
				}
			}
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) buildRefundApplyTx(applyCell *indexer.LiveCell) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: applyCell.OutPoint,
	})

	// witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasActionRefundApply, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// outputs
	fee := uint64(1e4)
	capacity := applyCell.Output.Capacity - fee
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: capacity,
		Lock:     applyCell.Output.Lock,
		Type:     nil,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	// cell deps
	applyContract, err := core.GetDasContractInfo(common.DasContractNameApplyRegisterCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	heightCell, err := h.dasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}
	applyConfigCell, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsApply)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		applyContract.ToCellDep(),
		timeCell.ToCellDep(),
		heightCell.ToCellDep(),
		applyConfigCell.ToCellDep(),
	)

	return &txParams, nil
}
