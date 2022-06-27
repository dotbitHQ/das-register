package block_parser

import (
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/shopspring/decimal"
)

func (b *BlockParser) ActionRenewAccount(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	orderTx, err := b.DbDao.GetOrderTxByHash(tables.TxActionRenewAccount, req.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetOrderTxByHash err: %s", err.Error())
		return
	} else if orderTx.Id > 0 {
		order, err := b.DbDao.GetOrderByOrderId(orderTx.OrderId)
		if err != nil {
			resp.Err = fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
		}
		if err := b.DbDao.DoActionRenewAccount(orderTx.OrderId, req.TxHash); err != nil {
			resp.Err = fmt.Errorf("DoActionRenewAccount err: %s", err.Error())
			return
		}
		// notify
		notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
			Action:  common.DasActionRenewAccount,
			Account: order.Account,
			OrderId: fmt.Sprintf("%s[%d]", order.OrderId, order.OrderType),
			Time:    req.BlockTimestamp,
			Hash:    req.TxHash,
		})
	} else {
		builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
		if err != nil {
			resp.Err = fmt.Errorf("PreAccountCellDataBuilderFromTx err: %s", err.Error())
			return
		}
		account := builder.Account
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
		args := req.Tx.Outputs[builder.Index].Lock.Args

		ownerHex, _, err := b.DasCore.Daf().ArgsToHex(args)
		if err != nil {
			resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
			return
		}
		// order info
		order := tables.TableDasOrderInfo{
			Id:                0,
			OrderType:         tables.OrderTypeOther,
			OrderId:           "",
			AccountId:         accountId,
			Account:           account,
			Action:            common.DasActionRenewAccount,
			ChainType:         ownerHex.ChainType,
			Address:           ownerHex.AddressHex,
			Timestamp:         int64(req.BlockTimestamp),
			PayTokenId:        "",
			PayType:           "",
			PayAmount:         decimal.Zero,
			Content:           "",
			PayStatus:         tables.TxStatusDefault,
			HedgeStatus:       tables.TxStatusDefault,
			PreRegisterStatus: tables.TxStatusDefault,
			RegisterStatus:    tables.RegisterStatusDefault,
			OrderStatus:       tables.OrderStatusClosed,
		}
		order.CreateOrderId()
		var orderTxList []tables.TableDasOrderTxInfo
		orderTxList = append(orderTxList, tables.TableDasOrderTxInfo{
			OrderId:   order.OrderId,
			Action:    tables.TxActionRenewAccount,
			Hash:      req.TxHash,
			Status:    tables.OrderTxStatusConfirm,
			Timestamp: int64(req.BlockTimestamp),
		})
		if err := b.DbDao.CreateOrderAndOrderTxs(&order, orderTxList); err != nil {
			resp.Err = fmt.Errorf("CreateOrderAndOrderTxs err: %s", err.Error())
			return
		}
		// notify
		notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
			Action:  common.DasActionRenewAccount,
			Account: account,
			OrderId: fmt.Sprintf("%s[%d]", order.OrderId, order.OrderType),
			Time:    req.BlockTimestamp,
			Hash:    req.TxHash,
		})
	}

	return
}
