package block_parser

import (
	"das_register_server/config"
	"das_register_server/notify"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
)

func (b *BlockParser) ActionBidExpiredAccountAuction(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	builder, err := witness.AccountCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	account := builder.Account
	oHex, _, err := b.DasCore.Daf().ArgsToHex(req.Tx.Outputs[builder.Index].Lock.Args)
	if err != nil {
		resp.Err = fmt.Errorf("ArgsToHex err: %s", err.Error())
		return
	}

	outpoint := common.OutPoint2String(req.TxHash, 0)
	if err := b.DbDao.UpdatePendingStatusToConfirm(req.Action, outpoint, req.BlockNumber, req.BlockTimestamp); err != nil {
		resp.Err = fmt.Errorf("UpdatePendingStatusToConfirm err: %s", err.Error())
		return
	}

	owner := oHex.AddressHex
	if len(owner) > 4 {
		owner = owner[len(owner)-4:]
	}

	order, err := b.DbDao.GetAuctionOrderStatus(oHex.ChainType, oHex.AddressHex, req.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetAuctionOrderStatus err: %s", err.Error())
		return
	}
	price := ""
	if order.Id != 0 {
		price = order.BasicPrice.Add(order.PremiumPrice).String()
	}

	larkText := fmt.Sprintf("Auction: %s, %s, %s", account, owner, price)
	notify.SendLarkTextNotify(config.Cfg.Notify.LarkRegisterOkKey, "", larkText)
	return
}
