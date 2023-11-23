package block_parser

import (
	"das_register_server/dao"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/types"
)

func (b *BlockParser) registerTransactionHandle() {
	b.mapTransactionHandle = make(map[string]FuncTransactionHandle)
	b.mapTransactionHandle[common.DasActionConfig] = b.ActionConfigCell
	b.mapTransactionHandle[tables.DasActionTransferBalance] = b.ActionTransferBalance
	b.mapTransactionHandle[common.DasActionWithdrawFromWallet] = b.ActionWithdrawFromWallet
	b.mapTransactionHandle[common.DasActionTransfer] = b.ActionTransferPayment
	//
	b.mapTransactionHandle[common.DasActionApplyRegister] = b.ActionApplyRegister
	b.mapTransactionHandle[common.DasActionPreRegister] = b.ActionPreRegister
	b.mapTransactionHandle[common.DasActionPropose] = b.ActionPropose
	b.mapTransactionHandle[common.DasActionExtendPropose] = b.ActionPropose
	b.mapTransactionHandle[common.DasActionConfirmProposal] = b.ActionConfirmProposal
	b.mapTransactionHandle[common.DasActionRenewAccount] = b.ActionRenewAccount

	b.mapTransactionHandle[common.DasActionEditRecords] = b.ActionEditRecords
	b.mapTransactionHandle[common.DasActionEditManager] = b.ActionEditManager
	b.mapTransactionHandle[common.DasActionTransferAccount] = b.ActionTransferAccount
	b.mapTransactionHandle[common.DasActionBidExpiredAccountAuction] = b.ActionBidExpiredAccountAuction
	//
	//b.mapTransactionHandle[common.DasActionDeclareReverseRecord] = b.ActionDeclareReverseRecord
	//b.mapTransactionHandle[common.DasActionRedeclareReverseRecord] = b.ActionRedeclareReverseRecord
	//b.mapTransactionHandle[common.DasActionRetractReverseRecord] = b.ActionRetractReverseRecord
}

func isCurrentVersionTx(tx *types.Transaction, name common.DasContractName) (bool, error) {
	contract, err := core.GetDasContractInfo(name)
	if err != nil {
		return false, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	isCV := false
	for _, v := range tx.Outputs {
		if v.Type == nil {
			continue
		}
		if contract.IsSameTypeId(v.Type.CodeHash) {
			isCV = true
			break
		}
	}
	return isCV, nil
}

type FuncTransactionHandleReq struct {
	DbDao          *dao.DbDao
	Tx             *types.Transaction
	TxHash         string
	BlockNumber    uint64
	BlockTimestamp uint64
	Action         common.DasAction
}

type FuncTransactionHandleResp struct {
	ActionName string
	Err        error
}

type FuncTransactionHandle func(FuncTransactionHandleReq) FuncTransactionHandleResp
