package timer

import (
	"das_register_server/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
)

type preCellRecycleParams struct {
	tipBlockNumber  uint64
	asContract      *core.DasContractInfo
	preContract     *core.DasContractInfo
	balanceContract *core.DasContractInfo
	dasContract     *core.DasContractInfo
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
	var res preCellRecycleParams
	res.tipBlockNumber = tipBlockNumber
	res.asContract = asContract
	res.preContract = preContract
	res.balanceContract = balanceContract
	res.dasContract = dasContract
	return &res, nil
}

// 24h 60*60*1e3
// 1d 24*60*60*1e3
func (t *TxTimer) getPreCellByMedianTime(p *preCellRecycleParams, blockRange, timestamp uint64) ([]*indexer.LiveCell, error) {
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

	var list []*indexer.LiveCell
	for i, v := range liveCells.Objects {
		numberBlock, err := t.dasCore.Client().GetBlockByNumber(t.ctx, v.BlockNumber)
		if err != nil {
			return nil, fmt.Errorf("GetBlockByNumber err: %s", err.Error())
		}
		if blockChainInfo.MedianTime < numberBlock.Header.Timestamp+timestamp*1e3 {
			log.Info("getPreCellByMedianTime:", blockChainInfo.MedianTime, numberBlock.Header.Timestamp, v.OutPoint.TxHash.String())
			break
		}
		list = append(list, liveCells.Objects[i])
	}

	return list, nil
}

var recyclePreBlockNumberEarly uint64

func (t *TxTimer) doRecyclePreEarly() error {
	p, err := t.getPreCellRecycleParams()
	if err != nil {
		return fmt.Errorf("getPreCellRecycleParams err: %s", err.Error())
	}

	timestamp := uint64(2 * 60)
	list, err := t.getPreCellByMedianTime(p, recyclePreBlockNumberEarly, timestamp)
	if err != nil {
		return fmt.Errorf("getPreCellByMedianTime err: %s", err.Error())
	}
	for _, v := range list {
		log.Info("doRecyclePreEarly:", v.OutPoint.TxHash.String())
		if err := t.doRecyclePreTx(v, p, timestamp); err != nil {
			log.Warn("doRecyclePreTx err:", err.Error(), v.OutPoint.TxHash.String())
		} else {
			recyclePreBlockNumberEarly = v.BlockNumber
		}
	}
	return nil
}
