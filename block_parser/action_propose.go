package block_parser

import (
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
	"strings"
)

func (b *BlockParser) ActionPropose(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameProposalCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	builderMap, err := witness.PreAccountCellDataBuilderMapFromTx(req.Tx, common.DataTypeDep)
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountCellDataBuilderMapFromTx err: %s", err.Error())
		return
	}

	var accounts []string
	var preHashList []string
	for _, v := range builderMap {
		txHash := req.Tx.CellDeps[v.Index].OutPoint.TxHash.Hex()
		preHashList = append(preHashList, txHash)
		accounts = append(accounts, v.Account)
	}
	// pre order action
	list, err := b.DbDao.GetOrderTxByActionAndHashList(tables.TxActionPreRegister, preHashList)
	if err != nil {
		resp.Err = fmt.Errorf("GetOrderTxByActionAndHashList err: %s", err.Error())
		return
	}
	var orderIds []string
	var orderTxList []tables.TableDasOrderTxInfo
	for _, v := range list {
		orderIds = append(orderIds, v.OrderId)
		orderTxList = append(orderTxList, tables.TableDasOrderTxInfo{
			OrderId:   v.OrderId,
			Action:    tables.TxActionPropose,
			Hash:      req.TxHash,
			Status:    tables.OrderTxStatusConfirm,
			Timestamp: int64(req.BlockTimestamp),
		})
	}
	// update
	if err := b.DbDao.DoActionPropose(orderIds, orderTxList); err != nil {
		resp.Err = fmt.Errorf("DoActionPropose err: %s", err.Error())
		return
	}
	// notify
	notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
		Action:  common.DasActionPropose,
		Account: strings.Join(accounts, ","),
		OrderId: "",
		Time:    req.BlockTimestamp,
		Hash:    req.TxHash,
	})

	return
}
