package block_parser

import (
	"das_register_server/config"
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

	var outpoints []string
	for _, v := range req.Tx.Inputs {
		outpoints = append(outpoints, common.OutPointStruct2String(v.PreviousOutput))
	}
	b.DasCache.ClearOutPoint(outpoints)

	builderOld, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	oldTx, err := b.DasCore.Client().GetTransaction(b.Ctx, req.Tx.Inputs[builderOld.Index].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	builderPreMap, err := witness.AccountIdCellDataBuilderFromTx(oldTx.Transaction, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	builderPre := builderPreMap[builderOld.AccountId]

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
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

	// notify
	owner := ownerHex.AddressHex
	if len(owner) > 4 {
		owner = owner[len(owner)-4:]
	}

	renewYears := (builder.ExpiredAt - builderPre.ExpiredAt) / uint64(common.OneYearSec)
	log.Info("ActionRenewAccount:", builder.Account, renewYears, builder.ExpiredAt, builderPre.ExpiredAt)
	if renewYears == 0 {
		renewYears = 1
	}
	larkText := fmt.Sprintf("Renew: %s, %d, %s", builder.Account, renewYears, owner)
	notify.SendLarkTextNotify(config.Cfg.Notify.LarkRegisterOkKey, "", larkText)

	return
}
