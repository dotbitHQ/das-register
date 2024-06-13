package timer

import (
	"das_register_server/config"
	"das_register_server/internal"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (t *TxTimer) checkRejected() error {
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		return fmt.Errorf("sync block number")
	}
	list, err := t.dbDao.SearchMaybeRejectedPending()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}

	var rejectedIDs []uint64
	for _, v := range list {
		res, err := t.dasCore.Client().GetTransaction(t.ctx, common.String2OutPointStruct(v.Outpoint).TxHash)
		if err != nil {
			return err
		}
		log.Info("checkRejected:", v.ChainType, v.Address, v.Action, v.Outpoint, res.TxStatus.Status)
		//if res.TxStatus.Status != types.TransactionStatusCommitted && res.TxStatus.Status != types.TransactionStatusPending && res.TxStatus.Status != types.TransactionStatusProposed {
		//	rejectedNum++
		//}
		if res.TxStatus.Status == types.TransactionStatusRejected {
			rejectedIDs = append(rejectedIDs, v.Id)
		}
		if res.TxStatus.Status == types.TransactionStatusCommitted {
			if block, err := t.dasCore.Client().GetBlock(t.ctx, *res.TxStatus.BlockHash); err == nil && block != nil && block.Header != nil {
				log.Info("UpdatePendingToConfirm:", v.Id)
				_ = t.dbDao.UpdatePendingToConfirm(v.Id, block.Header.Number, block.Header.Timestamp)
			}
		}
	}
	if len(rejectedIDs) > 0 {
		log.Info("checkRejected rejectedIDs:", len(rejectedIDs))
		if err = t.dbDao.UpdatePendingToRejected(rejectedIDs); err != nil {
			return fmt.Errorf("UpdatePendingToRejected err: %s", err.Error())
		}
	}
	return nil
}
