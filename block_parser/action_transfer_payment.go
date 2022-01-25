package block_parser

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
)

func (b *BlockParser) ActionTransferPayment(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	dasLock, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	balanceType, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		resp.Err = fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		return
	}

	res, err := b.DasCore.Client().GetTransaction(b.Ctx, req.Tx.Inputs[0].PreviousOutput.TxHash)
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}
	cellOutput := res.Transaction.Outputs[req.Tx.Inputs[0].PreviousOutput.Index]
	if !dasLock.IsSameTypeId(cellOutput.Lock.CodeHash) {
		return
	}
	if cellOutput.Type != nil && !balanceType.IsSameTypeId(cellOutput.Type.CodeHash) {
		return
	}
	outpoint := common.OutPoint2String(req.TxHash, 0)
	if err := b.DbDao.UpdatePendingStatusToConfirm(req.Action, outpoint, req.BlockNumber, req.BlockTimestamp); err != nil {
		resp.Err = fmt.Errorf("UpdatePendingStatusToConfirm err: %s", err.Error())
		return
	}

	return
}
