package timer

import (
	"das_register_server/config"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/txbuilder"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *TxTimer) doRefundApply() error {
	addrParse, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	applyContract, err := core.GetDasContractInfo(common.DasContractNameApplyRegisterCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	blockNumber, err := t.dasCore.Client().GetTipBlockNumber(t.ctx)
	if err != nil {
		return fmt.Errorf("GetTipBlockNumber err: %s", err.Error())
	}
	searchKey := indexer.SearchKey{
		Script:     addrParse.Script,
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:              applyContract.ToScript(nil),
			OutputDataLenRange:  nil,
			OutputCapacityRange: nil,
			BlockRange:          nil,
		},
	}
	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 10, "")
	if err != nil {
		return fmt.Errorf("GetCells err: %s", err.Error())
	}
	applyConfigCell, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsApply)
	if err != nil {
		return fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	for _, v := range liveCells.Objects {
		if v.BlockNumber+10000 > blockNumber {
			break
		}

		var txParams txbuilder.BuildTransactionParams

		// inputs
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})

		// witness action
		actionWitness, err := witness.GenActionDataWitness(common.DasActionRefundApply, nil)
		if err != nil {
			return fmt.Errorf("GenActionDataWitness err: %s", err.Error())
		}
		txParams.Witnesses = append(txParams.Witnesses, actionWitness)

		// outputs
		fee := uint64(1e4)
		capacity := v.Output.Capacity - fee
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: capacity,
			Lock:     v.Output.Lock,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})

		// cell deps
		timeCell, err := t.dasCore.GetTimeCell()
		if err != nil {
			return fmt.Errorf("GetTimeCell err: %s", err.Error())
		}
		heightCell, err := t.dasCore.GetHeightCell()
		if err != nil {
			return fmt.Errorf("GetHeightCell err: %s", err.Error())
		}

		txParams.CellDeps = append(txParams.CellDeps,
			applyContract.ToCellDep(),
			timeCell.ToCellDep(),
			heightCell.ToCellDep(),
			applyConfigCell.ToCellDep(),
		)

		txBuilder := txbuilder.NewDasTxBuilderFromBase(t.txBuilderBase, nil)
		if err := txBuilder.BuildTransaction(&txParams); err != nil {
			return fmt.Errorf("BuildTransaction err: %s", err.Error())
		}
		if hash, err := txBuilder.SendTransaction(); err != nil {
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		} else {
			log.Info("doRefundApply ok:", hash)
		}
	}
	return nil
}
