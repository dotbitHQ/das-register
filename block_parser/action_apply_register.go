package block_parser

import (
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
)

func (b *BlockParser) ActionApplyRegister(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameApplyRegisterCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	orderTx, err := b.DbDao.GetOrderTxByHash(tables.TxActionApplyRegister, req.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetOrderTxByHash err: %s", err.Error())
		return
	} else if orderTx.Id > 0 {
		order, err := b.DbDao.GetOrderByOrderId(orderTx.OrderId)
		if err != nil {
			resp.Err = fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
		}
		if err := b.DbDao.DoActionApplyRegister(orderTx.OrderId, orderTx.Hash); err != nil {
			resp.Err = fmt.Errorf("UpdatePreRegisterStatus err: %s", err.Error())
			return
		}
		notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
			Action:  common.DasActionApplyRegister,
			Account: order.Account,
			OrderId: order.OrderId,
			Time:    req.BlockTimestamp,
			Hash:    req.TxHash,
		})
	} else {
		notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
			Action:  common.DasActionApplyRegister,
			Account: "unknown",
			OrderId: "",
			Time:    req.BlockTimestamp,
			Hash:    req.TxHash,
		})
	}

	return
}
