package timer

import "fmt"

func (t *TxTimer) checkPending(pendingLimit int) (int, error) {
	list, err := t.dbDao.GetPendingList(pendingLimit)
	if err != nil {
		return pendingLimit, fmt.Errorf("GetPendingList err: %s", err.Error())
	}
	if len(list) == 0 {
		return pendingLimit, nil
	}

	var outpoints []string
	mapPending := make(map[string]uint64)
	for _, v := range list {
		outpoints = append(outpoints, v.Outpoint)
		mapPending[v.Outpoint] = v.Id
	}
	txs, err := t.dbDao.GetTransactionListByOutpoints(outpoints)
	if err != nil {
		return pendingLimit, fmt.Errorf("GetTransactionListByOutpoints err: %s", err.Error())
	}
	for _, v := range txs {
		if pendingId, ok := mapPending[v.Outpoint]; ok {
			if err = t.dbDao.UpdatePendingToConfirm(pendingId, v.BlockNumber, v.BlockTimestamp); err != nil {
				return pendingLimit, fmt.Errorf("UpdatePendingToConfirm err: %s", err.Error())
			}
		}
	}
	if len(list) == pendingLimit && len(txs) == 0 {
		pendingLimit += 10
	} else if pendingLimit > 10 && len(txs) > 0 {
		pendingLimit -= 10
	}
	return pendingLimit, nil
}
