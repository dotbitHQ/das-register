package timer

import (
	"das_register_server/notify"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"strings"
	"time"
)

func (t *TxTimer) doTxRejected() error {
	list, err := t.dbDao.GetMaybeRejectedRegisterTxs(time.Now().Add(-time.Hour*72).UnixNano()/1e6, time.Now().Add(-time.Minute*10).UnixNano()/1e6)
	if err != nil {
		return fmt.Errorf("GetMaybeRejectedRegisterTxs err: %s", err.Error())
	}
	if len(list) == 0 {
		return nil
	}

	msg := `> rejected register tx: %d
%s`
	var orderList []string
	for _, v := range list {
		txRes, err := t.dasCore.Client().GetTransaction(t.ctx, types.HexToHash(v.Hash))
		if err != nil {
			log.Error("GetTransaction err: ", err.Error())
		} else {
			log.Info("GetTransaction:", v.OrderId, v.Hash, txRes.TxStatus.Status)
			if txRes.TxStatus.Status == types.TransactionStatusRejected {
				notify.SendLarkErrNotify("UpdateRejectedTx", v.OrderId)
				if err := t.dbDao.UpdateRejectedTx(v.Action, v.OrderId); err != nil {
					log.Error("UpdateRejectedTx err: ", err.Error())
				}
				continue
			}
		}
		//
		sinceMin := time.Since(time.Unix(v.Timestamp/1000, 0)).Minutes()
		orderList = append(orderList, fmt.Sprintf("%s : %s (%.2f min)", v.Action, v.OrderId, sinceMin))
	}
	if len(orderList) > 0 {
		msg = fmt.Sprintf(msg, len(list), strings.Join(orderList, "\n"))
		notify.SendLarkErrNotify("Rejected Txs", msg)
	}
	return nil
}
