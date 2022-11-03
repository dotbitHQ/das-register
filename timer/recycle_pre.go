package timer

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
)

var recyclePreBlockNumber uint64

func (t *TxTimer) doRecyclePre() error {
	blockChainInfo, err := t.dasCore.Client().GetBlockchainInfo(t.ctx)
	if err != nil {
		return fmt.Errorf("GetBlockchainInfo err: %s", err.Error())
	}

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
	log.Info("doRecyclePre:", blockChainInfo.Chain, blockChainInfo.MedianTime)
	searchKey.Filter.BlockRange = &[2]uint64{recyclePreBlockNumber, 0}

	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 100, "")
	if err != nil {
		return fmt.Errorf("GetCells err: %s", err.Error())
	}
	if len(liveCells.Objects) == 0 {
		return nil
	}
	for _, v := range liveCells.Objects {
		numberBlock, err := t.dasCore.Client().GetBlockByNumber(t.ctx, v.BlockNumber)
		if err != nil {
			return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
		}

		if blockChainInfo.MedianTime > numberBlock.Header.Timestamp+24*60*60*1e3 {
			break
		}
		if err := t.doRecyclePreTx(v, dasContract, balanceContract, preContract); err != nil {
			log.Warn(err.Error(), v.OutPoint.TxHash.String())
		}
	}
	return nil
}

func (t *TxTimer) doRecyclePreTx(liveCell *indexer.LiveCell, dasContract, balanceContract, preContract *core.DasContractInfo) error {
	res, err := t.dasCore.Client().GetTransaction(t.ctx, liveCell.OutPoint.TxHash)
	if err != nil {
		return fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	preBuilder, err := witness.PreAccountCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return fmt.Errorf("PreAccountCellDataBuilderFromTx err: %s", err.Error())
	}
	refundLock := preBuilder.RefundLock
	if refundLock == nil {
		return fmt.Errorf("refundLock is nil")
	}
	refundLockScript := molecule.MoleculeScript2CkbScript(refundLock)

	// check lock code hash
	if !dasContract.IsSameTypeId(refundLockScript.CodeHash) && refundLockScript.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
		return fmt.Errorf("doRecyclePreTx code hash: %s", refundLockScript.CodeHash.String())
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

	log.Info("doRecyclePreTx tx:", common.OutPointStruct2String(liveCell.OutPoint), common.Bytes2Hex(refundLockScript.Args))

	// do refund
	var txParams txbuilder.BuildTransactionParams
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: liveCell.OutPoint,
		Since:          utils.SinceFromRelativeTimestamp(24 * 60 * 60),
	})

	fee := uint64(1e4)
	capacity := liveCell.Output.Capacity - fee
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
	txParams.CellDeps = append(txParams.CellDeps,
		preContract.ToCellDep(),
	)

	txBuilder := txbuilder.NewDasTxBuilderFromBase(t.txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(&txParams); err != nil {
		return fmt.Errorf("BuildTransaction err: %s", err.Error())
	}
	if hash, err := txBuilder.SendTransaction(); err != nil {
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		log.Info("doRecyclePreTx ok:", hash)
	}
	return nil
}
