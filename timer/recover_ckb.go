package timer

import (
	"das_register_server/config"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/txbuilder"
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
	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 50, "")
	if err != nil {
		return fmt.Errorf("GetCells err: %s", err.Error())
	}

	// build tx
	var txParams txbuilder.BuildTransactionParams

	// inputs
	total := uint64(0)
	for _, v := range liveCells.Objects {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
		total += v.Output.Capacity
	}

	// outputs
	capacity := 500 * common.OneCkb
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
