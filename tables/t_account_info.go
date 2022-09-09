package tables

import (
	"github.com/dotbitHQ/das-lib/common"
	"time"
)

type TableAccountInfo struct {
	Id                   uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber          uint64                `json:"block_number" gorm:"column:block_number"`
	Outpoint             string                `json:"outpoint" gorm:"column:outpoint"`
	AccountId            string                `json:"account_id" gorm:"account_id"`
	Account              string                `json:"account" gorm:"column:account"`
	OwnerChainType       common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type"`
	Owner                string                `json:"owner" gorm:"column:owner"`
	OwnerAlgorithmId     common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id"`
	ManagerChainType     common.ChainType      `json:"manager_chain_type" gorm:"column:manager_chain_type"`
	Manager              string                `json:"manager" gorm:"column:manager"`
	ManagerAlgorithmId   common.DasAlgorithmId `json:"manager_algorithm_id" gorm:"column:manager_algorithm_id"`
	Status               AccountStatus         `json:"status" gorm:"column:status"`
	RegisteredAt         uint64                `json:"registered_at" gorm:"column:registered_at"`
	ExpiredAt            uint64                `json:"expired_at" gorm:"column:expired_at"`
	ConfirmProposalHash  string                `json:"confirm_proposal_hash" gorm:"column:confirm_proposal_hash"`
	ParentAccountId      string                `json:"parent_account_id" gorm:"column:parent_account_id"`
	EnableSubAccount     EnableSubAccount      `json:"enable_sub_account" gorm:"column:enable_sub_account"`
	RenewSubAccountPrice uint64                `json:"renew_sub_account_price" gorm:"column:renew_sub_account_price"`
	Nonce                uint64                `json:"nonce" gorm:"column:nonce"`
}

type AccountStatus int

const (
	AccountStatusNotOpenRegister AccountStatus = -1
	AccountStatusNormal          AccountStatus = 0
	AccountStatusOnSale          AccountStatus = 1
	AccountStatusOnAuction       AccountStatus = 2
	AccountStatusOnCross         AccountStatus = 3
	TableNameAccountInfo                       = "t_account_info"
)

type EnableSubAccount = uint8

const (
	AccountEnableStatusOff EnableSubAccount = 0
	AccountEnableStatusOn  EnableSubAccount = 1
)

type SearchStatus int

const (
	SearchStatusRegisterNotOpen      SearchStatus = -1
	SearchStatusRegisterAble         SearchStatus = 0
	SearchStatusPaymentConfirm       SearchStatus = 1
	SearchStatusLockedAccount        SearchStatus = 2
	SearchStatusRegistering          SearchStatus = 3
	SearchStatusProposal             SearchStatus = 4
	SearchStatusConfirmProposal      SearchStatus = 5
	SearchStatusRegistered           SearchStatus = 6
	SearchStatusReservedAccount      SearchStatus = 7
	SearchStatusOnSale               SearchStatus = 8
	SearchStatusOnAuction            SearchStatus = 9
	SearchStatusUnAvailableAccount   SearchStatus = 13
	SearchStatusSubAccountUnRegister SearchStatus = 14
	SearchStatusOnCross              SearchStatus = 15
)

func (t *TableAccountInfo) IsExpired() bool {
	if int64(t.ExpiredAt) <= time.Now().Unix() {
		return true
	}
	return false
}

func (t *TableAccountInfo) TableName() string {
	return TableNameAccountInfo
}

func (t *TableAccountInfo) FormatAccountStatus() SearchStatus {
	switch t.Status {
	case AccountStatusNormal:
		return SearchStatusRegistered
	case AccountStatusOnSale:
		return SearchStatusOnSale
	case AccountStatusOnAuction:
		return SearchStatusOnAuction
	case AccountStatusOnCross:
		return SearchStatusOnCross
	default:
		return SearchStatusRegisterAble
	}
}

// ============= account category ===============

type Category int

const (
	CategoryDefault        Category = 0
	CategoryMainAccount    Category = 1
	CategorySubAccount     Category = 2
	CategoryOnSale         Category = 3
	CategoryExpireSoon     Category = 4
	CategoryToBeRecycled   Category = 5
	CategoryForInviterLink Category = 6
)
