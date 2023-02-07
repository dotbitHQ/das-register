package handle

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

type ReverseRecordSmtParams struct {
	DasLock       *types.Script
	DasType       *types.Script
	BalanceCells  []*indexer.LiveCell
	TotalCapacity uint64
	FeeCapacity   uint64
	SmtRoot       string
	LatestTxHash  types.Hash
}

func (h *HttpHandle) buildReverseSmtTx(req *ReverseRecordSmtParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// cell header dep
	latestReverseRecordTx, err := h.dasCore.Client().GetTransaction(h.ctx, req.LatestTxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err)
	}
	txParams.HeadDeps = append(txParams.HeadDeps, *latestReverseRecordTx.TxStatus.BlockHash)

	// cell deps
	balContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	reverseRecordRootContract, err := core.GetDasContractInfo(common.DasContractNameReverseRecordRootCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configCellMain, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsMain)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	configCellTypeArgsReverseRecord, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	configCellTypeArgsSMTNodeWhitelist, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsSMTNodeWhitelist)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		balContract.ToCellDep(),
		reverseRecordRootContract.ToCellDep(),
		configCellMain.ToCellDep(),
		configCellTypeArgsReverseRecord.ToCellDep(),
		configCellTypeArgsSMTNodeWhitelist.ToCellDep(),
	)

	// inputs
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: &types.OutPoint{
			TxHash: latestReverseRecordTx.Transaction.Hash,
			Index:  uint(0),
		},
	})
	for _, v := range req.BalanceCells {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, latestReverseRecordTx.Transaction.Outputs[0])
	txParams.OutputsData = append(txParams.OutputsData, []byte(req.SmtRoot))

	// change
	changeCapacity := req.TotalCapacity - req.FeeCapacity
	if changeCapacity > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: changeCapacity,
			Lock:     req.DasLock,
			Type:     req.DasType,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionUpdateReverseRecordRoot, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	return &txParams, nil
}
