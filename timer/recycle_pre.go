package timer

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
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
	p, err := t.getPreCellRecycleParams()
	if err != nil {
		return fmt.Errorf("getPreCellRecycleParams err: %s", err.Error())
	}

	timestamp := uint64(24 * 60 * 60)
	list, err := t.getPreCellByMedianTime(p, recyclePreBlockNumber, timestamp)
	if err != nil {
		return fmt.Errorf("getPreCellByMedianTime err: %s", err.Error())
	}
	for _, v := range list {
		log.Info("doRecyclePre:", v.OutPoint.TxHash.String())
		if err := t.doRecyclePreTx(v, p, timestamp); err != nil {
			log.Error("doRecyclePreTx err:", err.Error(), v.OutPoint.TxHash.String())
		} else {
			recyclePreBlockNumber = v.BlockNumber
		}
	}

	return nil
}

func (t *TxTimer) doRecyclePreTx(liveCell *indexer.LiveCell, p *preCellRecycleParams, timestamp uint64) error {
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
	if !p.dasContract.IsSameTypeId(refundLockScript.CodeHash) && refundLockScript.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
		return fmt.Errorf("doRecyclePreTx code hash: %s", refundLockScript.CodeHash.String())
	}

	var refundTypeScript *types.Script
	if p.dasContract.IsSameTypeId(refundLockScript.CodeHash) {
		ownerHex, _, err := t.dasCore.Daf().ScriptToHex(refundLockScript)
		if err != nil {
			return fmt.Errorf("ScriptToHex err: %s", err.Error())
		}
		if ownerHex.DasAlgorithmId == common.DasAlgorithmIdEth712 {
			refundTypeScript = p.balanceContract.ToScript(nil)
		}
	}

	log.Info("doRecyclePreTx tx:", common.OutPointStruct2String(liveCell.OutPoint), common.Bytes2Hex(refundLockScript.Args))

	// do refund
	var txParams txbuilder.BuildTransactionParams
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: liveCell.OutPoint,
		Since:          utils.SinceFromRelativeTimestamp(timestamp),
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
		p.preContract.ToCellDep(),
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
