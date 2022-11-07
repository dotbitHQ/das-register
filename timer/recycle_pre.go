package timer

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
)

var recyclePreBlockNumber uint64

const recycleTimestamp = uint64(24 * 60 * 60)

func (t *TxTimer) doRecyclePre() error {
	p, err := t.getPreCellRecycleParams()
	if err != nil {
		return fmt.Errorf("getPreCellRecycleParams err: %s", err.Error())
	}

	list, err := t.getPreCellByMedianTime(p, recyclePreBlockNumber, recycleTimestamp)
	if err != nil {
		return fmt.Errorf("getPreCellByMedianTime err: %s", err.Error())
	}
	for _, v := range list {
		log.Info("doRecyclePre:", v.liveCell.OutPoint.TxHash.String())
		if err := t.doRecyclePreTx(v, p, recycleTimestamp); err != nil {
			log.Error("doRecyclePreTx err:", err.Error(), v.liveCell.OutPoint.TxHash.String())
		} else {
			recyclePreBlockNumber = v.liveCell.BlockNumber
		}
	}

	return nil
}

func (t *TxTimer) doRecyclePreTx(preInfo preCellRecycleInfo, p *preCellRecycleParams, timestamp uint64) error {
	// check lock code hash
	if !p.dasContract.IsSameTypeId(preInfo.refundLockScript.CodeHash) && preInfo.refundLockScript.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
		return fmt.Errorf("doRecyclePreTx code hash: %s", preInfo.refundLockScript.CodeHash.String())
	}

	var refundTypeScript *types.Script
	if p.dasContract.IsSameTypeId(preInfo.refundLockScript.CodeHash) {
		ownerHex, _, err := t.dasCore.Daf().ScriptToHex(preInfo.refundLockScript)
		if err != nil {
			return fmt.Errorf("ScriptToHex err: %s", err.Error())
		}
		if ownerHex.DasAlgorithmId == common.DasAlgorithmIdEth712 {
			refundTypeScript = p.balanceContract.ToScript(nil)
		}
	}

	log.Info("doRecyclePreTx tx:", common.OutPointStruct2String(preInfo.liveCell.OutPoint), common.Bytes2Hex(preInfo.refundLockScript.Args))

	// do refund
	var txParams txbuilder.BuildTransactionParams
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: preInfo.liveCell.OutPoint,
		Since:          utils.SinceFromRelativeTimestamp(timestamp),
	})

	fee := uint64(1e4)
	capacity := preInfo.liveCell.Output.Capacity - fee
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: capacity,
		Lock:     preInfo.refundLockScript,
		Type:     refundTypeScript,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	// witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasActionRefundPreRegister, nil)
	if err != nil {
		return fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	witnessPre, _, _ := preInfo.preBuilder.GenWitness(&witness.PreAccountCellParam{
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
