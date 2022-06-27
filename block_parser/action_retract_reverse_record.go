package block_parser

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
)

func (b *BlockParser) ActionRetractReverseRecord(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	res, err := b.DasCore.Client().GetTransaction(b.Ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	if isCV, err := isCurrentVersionTx(res.Transaction, common.DasContractNameReverseRecordCellType); err != nil {
		resp.Err = fmt.Errorf("isisCurrentVersionTx err: %s", err.Error())
		return
	} else if !isCV {
		return
	}
	outpoint := common.OutPoint2String(req.TxHash, 0)
	if err := b.DbDao.UpdatePendingStatusToConfirm(req.Action, outpoint, req.BlockNumber, req.BlockTimestamp); err != nil {
		resp.Err = fmt.Errorf("UpdatePendingStatusToConfirm err: %s", err.Error())
		return
	}

	return
}
