package timer

import (
	"bytes"
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

	builder, err := t.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsApply)
	if err != nil {
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	applyMaxWaitingBlockNumber, err := molecule.Bytes2GoU32(builder.ConfigCellApply.ApplyMaxWaitingBlockNumber().RawData())
	if err != nil {
		return fmt.Errorf("ApplyMaxWaitingBlockNumber err: %s", err.Error())
	}
	if applyMaxWaitingBlockNumber < 5760 {
		applyMaxWaitingBlockNumber = 5760
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
			OutputDataLenRange:  &[2]uint64{48, 49},
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
		if v.BlockNumber+uint64(applyMaxWaitingBlockNumber) > blockNumber {
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

var preBlockNumber uint64

func (t *TxTimer) doRefundPre() error {
	addrParse, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}

	asContract, err := core.GetDasContractInfo(common.DasContractNameAlwaysSuccess)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	preContract, err := core.GetDasContractInfo(common.DasContractNamePreAccountCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	blockNumber, err := t.dasCore.Client().GetTipBlockNumber(t.ctx)
	if err != nil {
		return fmt.Errorf("GetTipBlockNumber err: %s", err.Error())
	}
	preMaxWaitingBlockNumber := uint64(5760)

	searchKey := indexer.SearchKey{
		Script:     asContract.ToScript(nil),
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:              preContract.ToScript(nil),
			OutputDataLenRange:  nil,
			OutputCapacityRange: nil,
			BlockRange:          nil,
		},
	}
	log.Info("doRefundPre:", preBlockNumber, blockNumber)
	if preBlockNumber > 0 && preBlockNumber < blockNumber {
		searchKey.Filter.BlockRange = &[2]uint64{preBlockNumber, blockNumber - preMaxWaitingBlockNumber}
	}

	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 10, "")
	if err != nil {
		return fmt.Errorf("GetCells err: %s", err.Error())
	}
	if len(liveCells.Objects) == 0 {
		preBlockNumber = 0
		return nil
	}

	for _, v := range liveCells.Objects {
		preBlockNumber = v.BlockNumber
		res, err := t.dasCore.Client().GetTransaction(t.ctx, v.OutPoint.TxHash)
		if err != nil {
			return fmt.Errorf("GetTransaction err: %s", err.Error())
		}
		preBuilder, err := witness.PreAccountCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
		if err != nil {
			continue
		} else {
			refundLock, _ := preBuilder.RefundLock()
			refundLockScript := molecule.MoleculeScript2CkbScript(refundLock)
			if bytes.Compare(addrParse.Script.Args, refundLockScript.Args) != 0 {
				continue
			}
			log.Info("doRefundPre:", common.OutPointStruct2String(v.OutPoint))

			// do refund
			var txParams txbuilder.BuildTransactionParams
			txParams.Inputs = append(txParams.Inputs, &types.CellInput{
				PreviousOutput: v.OutPoint,
			})

			fee := uint64(1e4)
			capacity := v.Output.Capacity - fee
			txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
				Capacity: capacity,
				Lock:     refundLockScript,
				Type:     nil,
			})
			txParams.OutputsData = append(txParams.OutputsData, []byte{})

			// witness action
			actionWitness, err := witness.GenActionDataWitness(common.DasActionRefundPreRegister, nil)
			if err != nil {
				return fmt.Errorf("GenActionDataWitness err: %s", err.Error())
			}
			txParams.Witnesses = append(txParams.Witnesses, actionWitness)

			witnessPre, _, _ := preBuilder.GenWitness(&witness.PreAccountCellParam{
				OldIndex: 0,
				Action:   common.DasActionRefundPreRegister,
			})
			txParams.Witnesses = append(txParams.Witnesses, witnessPre)

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
				preContract.ToCellDep(),
				timeCell.ToCellDep(),
				heightCell.ToCellDep(),
			)

			txBuilder := txbuilder.NewDasTxBuilderFromBase(t.txBuilderBase, nil)
			if err := txBuilder.BuildTransaction(&txParams); err != nil {
				return fmt.Errorf("BuildTransaction err: %s", err.Error())
			}
			if hash, err := txBuilder.SendTransaction(); err != nil {
				return fmt.Errorf("SendTransaction err: %s", err.Error())
			} else {
				log.Info("doRefundPre ok:", hash)
			}
		}
	}
	preBlockNumber++

	return nil
}

func (t *TxTimer) doCheckClosedAndUnRefund() error {
	list, err := t.dbDao.GetClosedAndUnRefundOrders()
	if err != nil {
		return fmt.Errorf("GetClosedAndUnRefundOrders err: %s", err.Error())
	}
	for _, v := range list {
		log.Info("doCheckClosedAndUnRefund:", v.OrderId)
		if err := t.dbDao.UpdatePayToRefund(v.OrderId); err != nil {
			return fmt.Errorf("UpdatePayToRefund err: %s", err.Error())
		}
	}
	return nil
}
