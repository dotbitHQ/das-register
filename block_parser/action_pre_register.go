package block_parser

import (
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/shopspring/decimal"
)

func (b *BlockParser) ActionPreRegister(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNamePreAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	orderTx, err := b.DbDao.GetOrderTxByHash(tables.TxActionPreRegister, req.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetOrderTxByHash err: %s", err.Error())
		return
	} else if orderTx.Id > 0 {
		order, err := b.DbDao.GetOrderByOrderId(orderTx.OrderId)
		if err != nil {
			resp.Err = fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
		}
		if err := b.DbDao.DoActionPreRegister(orderTx.OrderId, orderTx.Hash); err != nil {
			resp.Err = fmt.Errorf("DoActionPreRegister err: %s", err.Error())
			return
		}
		// notify
		notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
			Action:  common.DasActionPreRegister,
			Account: order.Account,
			OrderId: fmt.Sprintf("%s[%d]", order.OrderId, order.OrderType),
			Time:    req.BlockTimestamp,
			Hash:    req.TxHash,
		})
	} else {
		builder, err := witness.PreAccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
		if err != nil {
			resp.Err = fmt.Errorf("PreAccountCellDataBuilderFromTx err: %s", err.Error())
			return
		}
		account := builder.Account
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
		owner, _ := builder.OwnerLockArgsStr()
		_, _, oct, _, oa, _ := core.FormatDasLockToHexAddress(common.Hex2Bytes(owner))

		order := tables.TableDasOrderInfo{
			Id:                0,
			OrderType:         tables.OrderTypeOther,
			OrderId:           "",
			AccountId:         accountId,
			Account:           account,
			Action:            common.DasActionApplyRegister,
			ChainType:         oct,
			Address:           oa,
			Timestamp:         int64(req.BlockTimestamp),
			PayTokenId:        "",
			PayType:           "",
			PayAmount:         decimal.Zero,
			Content:           "",
			PayStatus:         tables.TxStatusDefault,
			HedgeStatus:       tables.TxStatusDefault,
			PreRegisterStatus: tables.TxStatusDefault,
			RegisterStatus:    tables.RegisterStatusProposal,
			OrderStatus:       tables.OrderStatusDefault,
		}
		order.CreateOrderId()
		var orderTxList []tables.TableDasOrderTxInfo
		orderTxList = append(orderTxList, tables.TableDasOrderTxInfo{
			OrderId:   order.OrderId,
			Action:    tables.TxActionApplyRegister,
			Hash:      req.Tx.Inputs[0].PreviousOutput.TxHash.Hex(),
			Status:    tables.OrderTxStatusConfirm,
			Timestamp: int64(req.BlockTimestamp),
		})
		orderTxList = append(orderTxList, tables.TableDasOrderTxInfo{
			OrderId:   order.OrderId,
			Action:    tables.TxActionPreRegister,
			Hash:      req.TxHash,
			Status:    tables.OrderTxStatusConfirm,
			Timestamp: int64(req.BlockTimestamp),
		})

		if err := b.DbDao.CreateOrderAndOrderTxs(&order, orderTxList); err != nil {
			resp.Err = fmt.Errorf("CreateOrderStatusAndOrderTxs err: %s", err.Error())
			return
		}
		// notify
		notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
			Action:  common.DasActionPreRegister,
			Account: account,
			OrderId: fmt.Sprintf("%s[%d]", order.OrderId, order.OrderType),
			Time:    req.BlockTimestamp,
			Hash:    req.TxHash,
		})
	}
	return
}