package block_parser

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/smt"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (b *BlockParser) ActionReverseRecordRoot(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameReverseRecordRootCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}

	outpoint := common.OutPoint2String(req.TxHash, 0)
	res, err := b.DasCore.Client().GetTransaction(b.Ctx, types.HexToHash(req.TxHash))
	if err != nil {
		resp.Err = fmt.Errorf("GetTransaction err: %s", err.Error())
		return
	}

	if res.TxStatus.Status == types.TransactionStatusRejected {
		resp.Err = b.DbDao.UpdateReverseSmtTaskInfoStatus(outpoint, tables.ReverseSmtStatusRollback, tables.ReverseSmtTxStatusReject)
		return
	}

	// 查询本地交易是否存在
	reverseSmtTaskInfo, err := b.DbDao.SearchReverseSmtTaskInfo(outpoint)
	if err != nil {
		resp.Err = err
		return
	}
	if reverseSmtTaskInfo.ID == 0 {
		// TODO 同步交易到本地
		return
	}

	smtRoot := string(res.Transaction.OutputsData[0])
	tree := smt.NewSmtSrv(b.SmtServer, "reverse_record")
	rootH256, err := tree.GetSmtRoot()
	if err != nil {
		resp.Err = err
		return
	}
	if smtRoot == rootH256.String() {
		resp.Err = b.DbDao.UpdateReverseSmtTaskInfoStatus(outpoint, tables.ReverseSmtStatusConfirm, tables.ReverseSmtTxStatusConfirm)
		return
	}

	// TODO 更新SMT到本地
	return
}
