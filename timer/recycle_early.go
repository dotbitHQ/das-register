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
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
	"github.com/robfig/cron/v3"
)

type preCellRecycleParams struct {
	tipBlockNumber  uint64
	asContract      *core.DasContractInfo
	preContract     *core.DasContractInfo
	balanceContract *core.DasContractInfo
	dasContract     *core.DasContractInfo
	addrParse       *address.ParsedAddress
}

func (t *TxTimer) getPreCellRecycleParams() (*preCellRecycleParams, error) {
	tipBlockNumber, err := t.dasCore.Client().GetTipBlockNumber(t.ctx)
	if err != nil {
		return nil, fmt.Errorf("GetTipBlockNumber err: %s", err.Error())
	}
	asContract, err := core.GetDasContractInfo(common.DasContractNameAlwaysSuccess)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	preContract, err := core.GetDasContractInfo(common.DasContractNamePreAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	balanceContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	addrParse, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		return nil, fmt.Errorf("address.Parse err: %s", err.Error())
	}
	var res preCellRecycleParams
	res.tipBlockNumber = tipBlockNumber
	res.asContract = asContract
	res.preContract = preContract
	res.balanceContract = balanceContract
	res.dasContract = dasContract
	res.addrParse = addrParse
	return &res, nil
}

type preCellRecycleInfo struct {
	liveCell         *indexer.LiveCell
	preBuilder       *witness.PreAccountCellDataBuilder
	refundLockScript *types.Script
}

// 24h 60*60*1e3
// 1d 24*60*60*1e3
func (t *TxTimer) getPreCellByMedianTime(p *preCellRecycleParams, blockRange, timestamp uint64) ([]preCellRecycleInfo, error) {
	searchKey := indexer.SearchKey{
		Script:     p.asContract.ToScript(nil),
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:              p.preContract.ToScript(nil),
			OutputDataLenRange:  nil,
			OutputCapacityRange: nil,
			BlockRange:          nil,
		},
	}
	if blockRange == 0 {
		if config.Cfg.Server.Net != common.DasNetTypeMainNet {
			blockRange = 1927285
		} else {
			blockRange = 4872287
		}
	}

	searchKey.Filter.BlockRange = &[2]uint64{blockRange, p.tipBlockNumber}
	liveCells, err := t.dasCore.Client().GetCells(t.ctx, &searchKey, indexer.SearchOrderAsc, 100, "")
	if err != nil {
		return nil, fmt.Errorf("GetCells err: %s", err.Error())
	}

	blockChainInfo, err := t.dasCore.Client().GetBlockchainInfo(t.ctx)
	if err != nil {
		return nil, fmt.Errorf("GetBlockchainInfo err: %s", err.Error())
	}

	var list []preCellRecycleInfo
	for i, v := range liveCells.Objects {
		numberBlock, err := t.dasCore.Client().GetBlockByNumber(t.ctx, v.BlockNumber)
		if err != nil {
			return nil, fmt.Errorf("GetBlockByNumber err: %s", err.Error())
		}
		if blockChainInfo.MedianTime < numberBlock.Header.Timestamp+timestamp*1e3 {
			log.Info("getPreCellByMedianTime:", blockChainInfo.MedianTime, numberBlock.Header.Timestamp, v.OutPoint.TxHash.String())
			break
		}

		res, err := t.dasCore.Client().GetTransaction(t.ctx, v.OutPoint.TxHash)
		if err != nil {
			return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
		}
		preBuilder, err := witness.PreAccountCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
		if err != nil {
			return nil, fmt.Errorf("PreAccountCellDataBuilderFromTx err: %s", err.Error())
		}
		refundLockScript := molecule.MoleculeScript2CkbScript(preBuilder.RefundLock)
		if !p.dasContract.IsSameTypeId(refundLockScript.CodeHash) && refundLockScript.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
			log.Warn("getPreCellByMedianTime code hash err:", common.OutPointStruct2String(v.OutPoint))
			continue
		}
		if recycleTimestampEarly == timestamp {
			if bytes.Compare(p.addrParse.Script.Args, refundLockScript.Args) != 0 {
				continue
			}
		}
		list = append(list, preCellRecycleInfo{
			liveCell:         liveCells.Objects[i],
			preBuilder:       preBuilder,
			refundLockScript: refundLockScript,
		})
	}

	return list, nil
}

var recyclePreBlockNumberEarly uint64

const recycleTimestampEarly = uint64(60 * 60)

func (t *TxTimer) doRecyclePreEarly() error {
	if !config.Cfg.Server.RecyclePreEarly {
		return nil
	}
	p, err := t.getPreCellRecycleParams()
	if err != nil {
		return fmt.Errorf("getPreCellRecycleParams err: %s", err.Error())
	}

	list, err := t.getPreCellByMedianTime(p, recyclePreBlockNumberEarly, recycleTimestampEarly)
	if err != nil {
		return fmt.Errorf("getPreCellByMedianTime err: %s", err.Error())
	}
	log.Info("doRecyclePreEarly:", len(list))

	var txParams txbuilder.BuildTransactionParams
	// witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasActionRefundPreRegister, nil)
	if err != nil {
		return fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	for i, v := range list {
		// check lock code hash
		if !p.dasContract.IsSameTypeId(v.refundLockScript.CodeHash) && v.refundLockScript.CodeHash.String() != transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
			log.Error("doRecyclePreEarly code hash: ", v.refundLockScript.CodeHash.String(), v.liveCell.OutPoint.TxHash.String())
			continue
		}
		var refundTypeScript *types.Script
		if p.dasContract.IsSameTypeId(v.refundLockScript.CodeHash) {
			ownerHex, _, err := t.dasCore.Daf().ScriptToHex(v.refundLockScript)
			if err != nil {
				return fmt.Errorf("ScriptToHex err: %s", err.Error())
			}
			if ownerHex.DasAlgorithmId == common.DasAlgorithmIdEth712 {
				refundTypeScript = p.balanceContract.ToScript(nil)
			}
		}
		log.Info("doRecyclePreTx tx:", common.OutPointStruct2String(v.liveCell.OutPoint), common.Bytes2Hex(v.refundLockScript.Args))
		// inputs
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.liveCell.OutPoint,
			Since:          utils.SinceFromRelativeTimestamp(recycleTimestampEarly),
		})

		// outputs
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: v.liveCell.Output.Capacity,
			Lock:     v.refundLockScript,
			Type:     refundTypeScript,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})

		// witness
		witnessPre, _, _ := v.preBuilder.GenWitness(&witness.PreAccountCellParam{
			OldIndex: uint32(i),
			Action:   common.DasActionRefundPreRegister,
		})
		txParams.Witnesses = append(txParams.Witnesses, witnessPre)
	}
	if len(txParams.Outputs) == 0 {
		return nil
	}

	// fee
	fee := uint64(1e4)
	liveCell, totalCapacity, err := t.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          t.dasCache,
		LockScript:        p.addrParse.Script,
		CapacityNeed:      common.MinCellOccupiedCkb + fee,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
			Since:          0,
		})
	}
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: totalCapacity - fee,
		Lock:     p.addrParse.Script,
		Type:     nil,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

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
		log.Info("doRecyclePreEarly ok:", hash)
	}

	return nil
}

func (t *TxTimer) DoRecyclePreEarly() {
	if config.Cfg.Server.RecyclePreEarlyCronSpec == "" {
		return
	}
	log.Info("DoRecyclePreEarly:", config.Cfg.Server.RecyclePreEarlyCronSpec)
	t.cron = cron.New(cron.WithSeconds())
	_, err := t.cron.AddFunc(config.Cfg.Server.RecyclePreEarlyCronSpec, func() {
		log.Info("doRecyclePreEarly start ...")
		if err := t.doRecyclePreEarly(); err != nil {
			log.Error("doRecyclePreEarly err: ", err.Error())
		}
		log.Info("doRecyclePreEarly end ...")
	})
	if err != nil {
		log.Error("DoRecyclePreEarly err: %s", err.Error())
		return
	}
	t.cron.Start()
}

func (t *TxTimer) CloseCron() {
	log.Warn("cron done")
	if t.cron != nil {
		t.cron.Stop()
	}
}
