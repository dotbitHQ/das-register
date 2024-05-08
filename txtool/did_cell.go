package txtool

import (
	"das_register_server/notify"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"time"
)

func (t *TxTool) RunDidCellTx() {
	tic := time.NewTicker(time.Second * 10)
	t.Wg.Add(1)
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tic.C:
				log.Debug("doDidCellTx start ...")
				if err := t.doDidCellTx(); err != nil {
					log.Error("doDidCellTx err: ", err.Error())
					notify.SendLarkErrNotify(common.DasActionTransferAccount, notify.GetLarkTextNotifyStr("doDidCellTx", "", err.Error()))
				}
				log.Debug("doDidCellTx end ...")
			case <-t.Ctx.Done():
				log.Debug("tx tool done")
				t.Wg.Done()
				return
			}
		}
	}()
}

type DidCellTxCache struct {
	BuilderTx *txbuilder.DasTxBuilderTransaction `json:"builder_tx"`
}

func (t *TxTool) doDidCellTx() error {
	actions := []common.DasAction{common.DasActionTransferAccount, common.DasActionRenewAccount}
	list, err := t.DbDao.GetNeedSendDidCellOrderList(actions)
	if err != nil {
		return fmt.Errorf("GetNeedSendPayOrderList err: %s", err.Error())
	}
	for _, v := range list {
		didCellTxStr, err := t.RC.GetCache(v.OrderId)
		if err != nil {
			return fmt.Errorf("GetCache err: %s", err.Error())
		}
		var txCache DidCellTxCache
		if err := json.Unmarshal([]byte(didCellTxStr), &txCache); err != nil {
			return fmt.Errorf("json.Unmarshal err: %s", err.Error())
		}
		txBuilder := txbuilder.NewDasTxBuilderFromBase(t.TxBuilderBase, txCache.BuilderTx)
		if err := t.DbDao.UpdatePayStatus(v.OrderId, tables.TxStatusSending, tables.TxStatusOk); err != nil {
			return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
		}
		log.Info("doDidCellTx:", txBuilder.TxString())
		hash, err := txBuilder.SendTransaction()
		if err != nil {
			if err := t.DbDao.UpdatePayStatus(v.OrderId, tables.TxStatusOk, tables.TxStatusSending); err != nil {
				log.Error("UpdatePayStatus err:", err.Error(), v.OrderId)
				notify.SendLarkErrNotify(common.DasActionTransferAccount, notify.GetLarkTextNotifyStr("UpdatePayStatus", v.OrderId, err.Error()))
			}
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		log.Info("SendTransaction ok:", common.DasActionTransferAccount, hash)
		// update tx hash
		orderTx := tables.TableDasOrderTxInfo{
			OrderId:   v.OrderId,
			Action:    tables.OrderTxAction(v.Action),
			Hash:      hash.Hex(),
			Status:    tables.OrderTxStatusDefault,
			Timestamp: time.Now().UnixMilli(),
		}
		if err := t.DbDao.CreateOrderTx(&orderTx); err != nil {
			log.Error("CreateOrderTx err:", err.Error(), v.OrderId, hash.Hex())
			notify.SendLarkErrNotify(common.DasActionTransferAccount, notify.GetLarkTextNotifyStr("CreateOrderTx", v.OrderId, err.Error()))
		}
	}
	return nil
}
