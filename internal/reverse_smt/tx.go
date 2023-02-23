package reverse_smt

import (
	"errors"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

type ReverseSmtParams struct {
	ServerLock    *types.Script
	BalanceCells  []*indexer.LiveCell
	TotalCapacity uint64
	FeeCapacity   uint64
	SmtRoot       string
	PreTx         *types.TransactionWithStatus
}

func BuildReverseSmtTx(req *ReverseSmtParams) (*txbuilder.BuildTransactionParams, error) {
	if req.PreTx == nil {
		return nil, errors.New("params error PreTx can't be nil")
	}

	var txParams txbuilder.BuildTransactionParams

	// cell header dep
	txParams.HeadDeps = append(txParams.HeadDeps, *req.PreTx.TxStatus.BlockHash)

	// cell deps
	soEth, err := core.GetDasSoScript(common.SoScriptTypeEth)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasSoScript err: %s", err.Error())
	}
	soTron, err := core.GetDasSoScript(common.SoScriptTypeTron)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasSoScript err: %s", err.Error())
	}
	soEd25519, err := core.GetDasSoScript(common.SoScriptTypeEd25519)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasSoScript err: %s", err.Error())
	}
	balContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasContractInfo err: %s", err.Error())
	}
	reverseRecordRootContract, err := core.GetDasContractInfo(common.DasContractNameReverseRecordRootCellType)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasContractInfo err: %s", err.Error())
	}
	configCellMain, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsMain)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasConfigCellInfo err: %s", err.Error())
	}
	configCellTypeArgsReverseRecord, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsReverseRecord)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasConfigCellInfo err: %s", err.Error())
	}
	configCellTypeArgsSMTNodeWhitelist, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsSMTNodeWhitelist)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		soEth.ToCellDep(),
		soTron.ToCellDep(),
		soEd25519.ToCellDep(),
		balContract.ToCellDep(),
		reverseRecordRootContract.ToCellDep(),
		configCellMain.ToCellDep(),
		configCellTypeArgsReverseRecord.ToCellDep(),
		configCellTypeArgsSMTNodeWhitelist.ToCellDep(),
	)

	// inputs
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: &types.OutPoint{
			TxHash: req.PreTx.Transaction.Hash,
			Index:  uint(0),
		},
	})
	for _, v := range req.BalanceCells {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// outputs
	txParams.Outputs = append(txParams.Outputs, req.PreTx.Transaction.Outputs[0])
	txParams.OutputsData = append(txParams.OutputsData, []byte(req.SmtRoot))

	// change
	changeCapacity := req.TotalCapacity - req.FeeCapacity
	if changeCapacity > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: changeCapacity,
			Lock:     req.ServerLock,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionUpdateReverseRecordRoot, nil)
	if err != nil {
		return nil, fmt.Errorf("BuildReverseSmtTx GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	return &txParams, nil
}
