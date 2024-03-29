package timer

import (
	"das_register_server/config"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *TxTimer) doRecoverCkb() error {
	addrParse, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	searchKey := indexer.SearchKey{
		Script:     addrParse.Script,
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:              nil,
			OutputDataLenRange:  &[2]uint64{32, 33},
			OutputCapacityRange: nil,
			BlockRange:          nil,
		},
	}
	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 200, "")
	if err != nil {
		return fmt.Errorf("GetCells err: %s", err.Error())
	}

	// build tx
	var txParams txbuilder.BuildTransactionParams

	// inputs
	total := uint64(0)
	for _, v := range liveCells.Objects {
		if v.Output.Type != nil {
			continue
		}
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
		total += v.Output.Capacity
	}
	log.Info("doRecoverCkb:", total, len(liveCells.Objects), config.Cfg.Server.RecoverCkb)
	// outputs
	capacity := 2000 * common.OneCkb
	if config.Cfg.Server.RecoverCkb > 0 {
		capacity = config.Cfg.Server.RecoverCkb * common.OneCkb
	}
	if total < capacity*2 {
		return nil
	}
	for total > capacity*2 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: capacity,
			Lock:     addrParse.Script,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
		total -= capacity
	}
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: total,
		Lock:     addrParse.Script,
		Type:     nil,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	//
	actionWitness, err := witness.GenActionDataWitness(tables.DasActionRefundPay, nil)
	if err != nil {
		return fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	//
	txBuilder := txbuilder.NewDasTxBuilderFromBase(t.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(&txParams); err != nil {
		return fmt.Errorf("BuildTransaction err: %s", err.Error())
	}
	sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
	changeCapacity := txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity
	changeCapacity = changeCapacity - sizeInBlock - 5000
	txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity = changeCapacity

	if hash, err := txBuilder.SendTransaction(); err != nil {
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		log.Info("doRecoverCkb:", hash.String())
	}
	return nil
}
