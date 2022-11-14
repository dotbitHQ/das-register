package timer

import (
	"das_register_server/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
)

func (t *TxTimer) doRecycleApply() error {
	addrParse, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	applyContract, err := core.GetDasContractInfo(common.DasContractNameApplyRegisterCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	builder, err := t.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsApply)
	if err != nil {
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	applyMaxWaitingBlockNumber, err := molecule.Bytes2GoU32(builder.ConfigCellApply.ApplyMaxWaitingBlockNumber().RawData())
	if err != nil {
		return fmt.Errorf("ApplyMaxWaitingBlockNumber err: %s", err.Error())
	}
	log.Info("doRecycleApply:", applyMaxWaitingBlockNumber)

	searchKey := indexer.SearchKey{
		Script:     addrParse.Script,
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:             applyContract.ToScript(nil),
			OutputDataLenRange: &[2]uint64{48, 49},
		},
	}
	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 200, "")
	if err != nil {
		return fmt.Errorf("GetCells err: %s", err.Error())
	}
	log.Info("doRecycleApply:", len(liveCells.Objects))
	if len(liveCells.Objects) == 0 {
		return nil
	}

	tipBlockNumber, err := t.dasCore.Client().GetTipBlockNumber(t.ctx)
	if err != nil {
		return fmt.Errorf("GetTipBlockNumber err: %s", err.Error())
	}
	for _, v := range liveCells.Objects {
		if tipBlockNumber < v.BlockNumber+uint64(applyMaxWaitingBlockNumber) {
			break
		}
		log.Info("doRecycleApply:", v.BlockNumber, v.OutPoint.TxHash.String())
		var txParams txbuilder.BuildTransactionParams

		// inputs
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
			Since:          utils.SinceFromRelativeBlockNumber(uint64(applyMaxWaitingBlockNumber)),
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
		applyConfigCell, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsApply)
		if err != nil {
			return fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
		}
		txParams.CellDeps = append(txParams.CellDeps,
			applyContract.ToCellDep(),
			applyConfigCell.ToCellDep(),
		)

		txBuilder := txbuilder.NewDasTxBuilderFromBase(t.txBuilderBase, nil)
		if err := txBuilder.BuildTransaction(&txParams); err != nil {
			return fmt.Errorf("BuildTransaction err: %s", err.Error())
		}
		if hash, err := txBuilder.SendTransaction(); err != nil {
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		} else {
			log.Info("doRecycleApply ok:", hash)
		}
	}

	return nil
}
