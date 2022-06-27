package timer

import (
	"das_register_server/config"
	"das_register_server/internal"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"time"
)

func (t *TxTimer) checkRejected() error {
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		return fmt.Errorf("sync block number")
	}
	timestamp := time.Now().Add(-time.Minute*6).UnixNano() / 1e6
	list, err := t.dbDao.SearchMaybeRejectedPending(timestamp)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}
	rejectedNum := 0
	for _, v := range list {
		res, err := t.dasCore.Client().GetTransaction(t.ctx, common.String2OutPointStruct(v.Outpoint).TxHash)
		if err != nil {
			return err
		}
		log.Info("checkRejected:", v.ChainType, v.Address, v.Action, v.Outpoint, res.TxStatus.Status)
		if res.TxStatus.Status != types.TransactionStatusCommitted && res.TxStatus.Status != types.TransactionStatusPending && res.TxStatus.Status != types.TransactionStatusProposed {
			rejectedNum++
		}
		if res.TxStatus.Status == types.TransactionStatusCommitted {
			if block, err := t.dasCore.Client().GetBlock(t.ctx, *res.TxStatus.BlockHash); err == nil && block != nil && block.Header != nil {
				log.Info("UpdatePendingToConfirm:", v.Id)
				_ = t.dbDao.UpdatePendingToConfirm(v.Id, block.Header.Number, block.Header.Timestamp)
			}
		}
	}
	if rejectedNum > 0 && rejectedNum == len(list) {
		if err = t.dbDao.UpdatePendingToRejected(timestamp); err != nil {
			return fmt.Errorf("UpdatePendingToRejected err: %s", err.Error())
		}
	}
	return nil
}
