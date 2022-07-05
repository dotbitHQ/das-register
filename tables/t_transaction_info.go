package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableTransactionInfo struct {
	Id             uint64           `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber    uint64           `json:"block_number" gorm:"column:block_number"`
	AccountId      string           `json:"account_id" gorm:"account_id"`
	Account        string           `json:"account" gorm:"column:account"`
	Action         string           `json:"action" gorm:"column:action"`
	ServiceType    int              `json:"service_type" gorm:"column:service_type"`
	ChainType      common.ChainType `json:"chain_type" gorm:"column:chain_type"`
	Address        string           `json:"address" gorm:"column:address"`
	Capacity       uint64           `json:"capacity" gorm:"column:capacity"`
	Outpoint       string           `json:"outpoint" gorm:"column:outpoint"`
	BlockTimestamp uint64           `json:"block_timestamp" gorm:"column:block_timestamp"`
	Status         int              `json:"status" gorm:"column:status"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at"`
}

const (
	TableNameTransactionInfo = "t_transaction_info"
)

func (t *TableTransactionInfo) TableName() string {
	return TableNameTransactionInfo
}

type TxAction = int

const (
	ActionUndefined          TxAction = 99 // undefined
	ActionWithdrawFromWallet TxAction = 0  // withdraw
	ActionConsolidateIncome  TxAction = 1  // rewards
	ActionStartAccountSale   TxAction = 2  // on sale
	ActionEditAccountSale    TxAction = 3  // edit sale
	ActionCancelAccountSale  TxAction = 4  // cancel sale
	ActionBuyAccount         TxAction = 5  // buy account
	ActionSaleAccount        TxAction = 6  // sale account
	ActionTransferBalance    TxAction = 7  // transfer balance

	ActionDeclareReverseRecord   TxAction = 8  // declare reverse
	ActionRedeclareReverseRecord TxAction = 9  // edit reverse
	ActionRetractReverseRecord   TxAction = 10 // delete reverse

	ActionDasBalanceTransfer TxAction = 11 // das balance pay
	ActionEditRecords        TxAction = 12 // edit records
	ActionTransferAccount    TxAction = 13 // edit owner
	ActionEditManager        TxAction = 14 // edit manager
	ActionRenewAccount       TxAction = 15 // renew

	ActionMakeOffer     TxAction = 16
	ActionEditOfferAdd  TxAction = 17
	ActionCancelOffer   TxAction = 18
	ActionAcceptOffer   TxAction = 19
	ActionOfferAccepted TxAction = 20
	ActionEditOfferSub  TxAction = 21
	ActionOrderRefund   TxAction = 22

	ActionEnableSubAccount          TxAction = 23
	ActionCreateSubAccount          TxAction = 24
	ActionBalanceDeposit            TxAction = 25
	ActionRecycleExpiredAccount     TxAction = 26
	ActionForceRecoverAccountStatus TxAction = 27

	DasActionTransferBalance = "transfer_balance"
	DasActionBalanceDeposit  = "balance_deposit"
	DasActionRefundPay       = "refund_pay"
	DasActionSaleAccount     = "sale_account"
	DasActionOfferAccepted   = "offer_accepted"
	DasActionEditOfferAdd    = "offer_edit_add"
	DasActionEditOfferSub    = "offer_edit_sub"
	DasActionOrderRefund     = "order_refund"
)

func FormatTxAction(action string) TxAction {
	switch action {
	case common.DasActionWithdrawFromWallet:
		return ActionWithdrawFromWallet
	case common.DasActionConsolidateIncome:
		return ActionConsolidateIncome
	case common.DasActionStartAccountSale:
		return ActionStartAccountSale
	case common.DasActionEditAccountSale:
		return ActionEditAccountSale
	case common.DasActionCancelAccountSale:
		return ActionCancelAccountSale
	case common.DasActionBuyAccount:
		return ActionBuyAccount
	case DasActionSaleAccount:
		return ActionSaleAccount
	case DasActionTransferBalance:
		return ActionTransferBalance
	case common.DasActionTransfer:
		return ActionDasBalanceTransfer
	case common.DasActionEditRecords:
		return ActionEditRecords
	case common.DasActionTransferAccount:
		return ActionTransferAccount
	case common.DasActionEditManager:
		return ActionEditManager
	case common.DasActionRenewAccount:
		return ActionRenewAccount
	case common.DasActionDeclareReverseRecord:
		return ActionDeclareReverseRecord
	case common.DasActionRedeclareReverseRecord:
		return ActionRedeclareReverseRecord
	case common.DasActionRetractReverseRecord:
		return ActionRetractReverseRecord
	case common.DasActionMakeOffer:
		return ActionMakeOffer
	case DasActionEditOfferAdd:
		return ActionEditOfferAdd
	case DasActionEditOfferSub:
		return ActionEditOfferSub
	case common.DasActionCancelOffer:
		return ActionCancelOffer
	case common.DasActionAcceptOffer:
		return ActionAcceptOffer
	case DasActionOfferAccepted:
		return ActionOfferAccepted
	case DasActionOrderRefund:
		return ActionOrderRefund
	case DasActionBalanceDeposit:
		return ActionBalanceDeposit
	case common.DasActionEnableSubAccount:
		return ActionEnableSubAccount
	case common.DasActionCreateSubAccount:
		return ActionCreateSubAccount
	case common.DasActionRecycleExpiredAccount:
		return ActionRecycleExpiredAccount
	case common.DasActionForceRecoverAccountStatus:
		return ActionForceRecoverAccountStatus
	}
	return ActionUndefined
}

func FormatActionType(actionType TxAction) string {
	switch actionType {
	case ActionWithdrawFromWallet:
		return common.DasActionWithdrawFromWallet
	case ActionConsolidateIncome:
		return common.DasActionConsolidateIncome
	case ActionStartAccountSale:
		return common.DasActionStartAccountSale
	case ActionEditAccountSale:
		return common.DasActionEditAccountSale
	case ActionCancelAccountSale:
		return common.DasActionCancelAccountSale
	case ActionBuyAccount:
		return common.DasActionBuyAccount
	case ActionSaleAccount:
		return DasActionSaleAccount
	case ActionTransferBalance:
		return DasActionTransferBalance
	case ActionDasBalanceTransfer:
		return common.DasActionTransfer
	case ActionEditRecords:
		return common.DasActionEditRecords
	case ActionTransferAccount:
		return common.DasActionTransferAccount
	case ActionEditManager:
		return common.DasActionEditManager
	case ActionRenewAccount:
		return common.DasActionRenewAccount
	case ActionDeclareReverseRecord:
		return common.DasActionDeclareReverseRecord
	case ActionRedeclareReverseRecord:
		return common.DasActionRedeclareReverseRecord
	case ActionRetractReverseRecord:
		return common.DasActionRetractReverseRecord
	case ActionMakeOffer:
		return common.DasActionMakeOffer
	case ActionEditOfferAdd:
		return DasActionEditOfferAdd
	case ActionEditOfferSub:
		return DasActionEditOfferSub
	case ActionCancelOffer:
		return common.DasActionCancelOffer
	case ActionAcceptOffer:
		return common.DasActionAcceptOffer
	case ActionOfferAccepted:
		return DasActionOfferAccepted
	case ActionOrderRefund:
		return DasActionOrderRefund
	case ActionRecycleExpiredAccount:
		return common.DasActionRecycleExpiredAccount
	case ActionForceRecoverAccountStatus:
		return common.DasActionForceRecoverAccountStatus
	}
	return ""
}
