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
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"strings"
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
	if applyMaxWaitingBlockNumber < 8640 {
		applyMaxWaitingBlockNumber = 8640
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
	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 200, "")
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
		// check lock code hash
		if v.Output.Lock.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
			log.Warn("doRefundApply tx code hash err:", common.OutPointStruct2String(v.OutPoint))
			continue
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
	//addrParse, err := address.Parse(config.Cfg.Server.PayServerAddress)
	//if err != nil {
	//	return fmt.Errorf("address.Parse err: %s", err.Error())
	//}

	dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	balanceContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return fmt.Errorf("GetDasContractInfo err: %s", err.Error())
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
	preMaxWaitingBlockNumber := uint64(8640)

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
	log.Info("doRefundPre:", preBlockNumber, blockNumber-preMaxWaitingBlockNumber, blockNumber)
	searchKey.Filter.BlockRange = &[2]uint64{preBlockNumber, blockNumber - preMaxWaitingBlockNumber}

	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 100, "")
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
			refundLock := preBuilder.RefundLock
			if refundLock == nil {
				continue
			}
			refundLockScript := molecule.MoleculeScript2CkbScript(refundLock)
			//if bytes.Compare(addrParse.Script.Args, refundLockScript.Args) != 0 {
			//	continue
			//}

			// check lock code hash
			if !dasContract.IsSameTypeId(refundLockScript.CodeHash) && refundLockScript.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
				log.Warn("doRefundPre tx code hash err:", common.OutPointStruct2String(v.OutPoint))
				continue
			}

			var refundTypeScript *types.Script
			if dasContract.IsSameTypeId(refundLockScript.CodeHash) {
				ownerHex, _, err := t.dasCore.Daf().ScriptToHex(refundLockScript)
				if err != nil {
					return fmt.Errorf("ScriptToHex err: %s", err.Error())
				}
				if ownerHex.DasAlgorithmId == common.DasAlgorithmIdEth712 {
					refundTypeScript = balanceContract.ToScript(nil)
				}
			}

			log.Info("doRefundPre tx:", common.OutPointStruct2String(v.OutPoint), common.Bytes2Hex(refundLockScript.Args))

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
				Type:     refundTypeScript,
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
				if strings.Contains(err.Error(), "error code 83") {
					preBlockNumber--
					break
				}
				log.Error("doRefundPre SendTransaction err: ", err.Error())
				continue
				//return fmt.Errorf("SendTransaction err: %s", err.Error())
			} else {
				log.Info("doRefundPre ok:", hash)
			}
		}
	}
	preBlockNumber++

	return nil
}

func (t *TxTimer) doCheckClosedAndUnRefund() error {
	//SELECT o.order_id,o.pay_status,o.pre_register_status
	//FROM t_das_order_info o JOIN t_das_order_pay_info p ON o.order_id=p.order_id
	//AND o.action='renew_account' AND o.order_type='1' AND o.order_status='1'
	//AND o.pay_status='1' AND p.`status`='1' AND p.refund_status='0'

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
