package tables

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/shopspring/decimal"
	"time"
)

type TableDasOrderInfo struct {
	Id                uint64           `json:"id" gorm:"column:id;primaryKey;type:bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT ''"`
	OrderType         OrderType        `json:"order_type" gorm:"column:order_type;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-self 2-other'"`
	OrderId           string           `json:"order_id" gorm:"column:order_id;uniqueIndex:uk_order_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	AccountId         string           `json:"account_id" gorm:"account_id;index:k_account_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Account           string           `json:"account" gorm:"column:account;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	Action            common.DasAction `json:"action" gorm:"column:action;index:k_action;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	ChainType         common.ChainType `json:"chain_type" gorm:"column:chain_type;index:k_chain_type_address;type:smallint(6) NOT NULL DEFAULT '0' COMMENT 'order chain type'"`
	Address           string           `json:"address" gorm:"column:address;index:k_chain_type_address;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'order address'"`
	Timestamp         int64            `json:"timestamp" gorm:"column:timestamp;type:bigint(20) NOT NULL DEFAULT '0' COMMENT 'order time'"`
	PayTokenId        PayTokenId       `json:"pay_token_id" gorm:"column:pay_token_id;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	PayType           PayType          `json:"pay_type" gorm:"column:pay_type;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT ''"`
	PayAmount         decimal.Decimal  `json:"pay_amount" gorm:"column:pay_amount;type:decimal(60,0) NOT NULL DEFAULT '0' COMMENT ''"`
	Content           string           `json:"content" gorm:"column:content;type:text NOT NULL COMMENT 'order detail'"`
	PayStatus         TxStatus         `json:"pay_status" gorm:"column:pay_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok'"`
	HedgeStatus       TxStatus         `json:"hedge_status" gorm:"column:hedge_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok'"`
	PreRegisterStatus TxStatus         `json:"pre_register_status" gorm:"column:pre_register_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-ing 2-ok'"`
	RegisterStatus    RegisterStatus   `json:"register_status" gorm:"column:register_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-6'"`
	OrderStatus       OrderStatus      `json:"order_status" gorm:"column:order_status;index:k_order_status;type:smallint(6) NOT NULL DEFAULT '0' COMMENT '1-closed'"`
	CoinType          string           `json:"coin_type" gorm:"column:coin_type; type:varchar(255) NOT NULL DEFAULT '';"`
	CrossCoinType     string           `json:"cross_coin_type" gorm:"column:cross_coin_type; type:varchar(255) NOT NULL DEFAULT '';"`
	IsUniPay          IsUniPay         `json:"is_uni_pay" gorm:"column:is_uni_pay; type:smallint(6) NOT NULL DEFAULT '0' COMMENT '0-no 1-yes';"`
	PremiumPercentage decimal.Decimal  `json:"premium_percentage" gorm:"column:premium_percentage; type:decimal(20,10) NOT NULL DEFAULT '0' COMMENT '';"`
	PremiumBase       decimal.Decimal  `json:"premium_base" gorm:"column:premium_base; type:decimal(20,10) NOT NULL DEFAULT '0' COMMENT '';"`
	PremiumAmount     decimal.Decimal  `json:"premium_amount" gorm:"column:premium_amount; type:decimal(60,0) NOT NULL DEFAULT '0' COMMENT '';"`
	CreatedAt         time.Time        `json:"created_at" gorm:"column:created_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''"`
	UpdatedAt         time.Time        `json:"updated_at" gorm:"column:updated_at;type:timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''"`
}

const (
	TableNameDasOrderInfo = "t_das_order_info"
)

func (t *TableDasOrderInfo) TableName() string {
	return TableNameDasOrderInfo
}

func (t *TableDasOrderInfo) CreateOrderId() {
	t.OrderId = CreateOrderId(t.OrderType, t.AccountId, t.Action, t.ChainType, t.Address, t.Timestamp)
}

func CreateOrderId(orderType OrderType, accountId string, action common.DasAction, chainType common.ChainType, address string, timestamp int64) string {
	orderId := fmt.Sprintf("%d%s%s%d%s%d", orderType, accountId, action, chainType, address, timestamp)
	orderId = fmt.Sprintf("%x", md5.Sum([]byte(orderId)))
	return orderId
}

func (t *TableDasOrderInfo) GetContent() (TableOrderContent, error) {
	var content TableOrderContent
	if t.Content != "" {
		if err := json.Unmarshal([]byte(t.Content), &content); err != nil {
			return content, err
		}
	}
	return content, nil
}

type IsUniPay int

const (
	IsUniPayFalse IsUniPay = 0
	IsUniPayTrue  IsUniPay = 1
)

type OrderType int

const (
	OrderTypeSelf  OrderType = 1
	OrderTypeOther OrderType = 2
)

type PayTokenId string

const (
	TokenIdDas         PayTokenId = "ckb_das"
	TokenIdCkb         PayTokenId = "ckb_ckb"
	TokenIdCkbInternal PayTokenId = "ckb_internal"
	TokenIdEth         PayTokenId = "eth_eth"
	TokenIdTrx         PayTokenId = "tron_trx"
	TokenIdWx          PayTokenId = "wx_cny"
	TokenIdBnb         PayTokenId = "bsc_bnb"
	TokenIdMatic       PayTokenId = "polygon_matic"
	TokenCoupon        PayTokenId = "coupon"
	TokenIdDoge        PayTokenId = "doge_doge"
	TokenIdStripeUSD   PayTokenId = "stripe_usd"
	TokenIdErc20USDT   PayTokenId = "eth_erc20_usdt"
	TokenIdTrc20USDT   PayTokenId = "tron_trc20_usdt"
	TokenIdBep20USDT   PayTokenId = "bsc_bep20_usdt"
)

func (p PayTokenId) IsTokenIdCkbInternal() bool {
	return p == TokenIdCkbInternal
}

func (p PayTokenId) ToChainString() string {
	switch p {
	case TokenIdDas, TokenIdCkb, TokenIdCkbInternal:
		return "ckb"
	case TokenIdEth:
		return "eth"
	case TokenIdTrx:
		return "tron"
	case TokenIdBnb:
		return "bsc"
	case TokenIdMatic:
		return "polygon"
	case TokenIdDoge:
		return "doge"
	}
	return ""
}

type PayType string

const (
	PayTypeDefault  PayType = ""
	PayTypeWxh5     PayType = "wx_h5"
	PayTypeWxNative PayType = "wx_native"
)

type TxStatus int

const (
	TxStatusDefault TxStatus = 0
	TxStatusSending TxStatus = 1
	TxStatusOk      TxStatus = 2
)

type TableOrderContent struct {
	AccountCharStr []common.AccountCharSet `json:"account_char_str"`
	InviterAccount string                  `json:"inviter_account"`
	ChannelAccount string                  `json:"channel_account"`
	RegisterYears  int                     `json:"register_years"`
	AmountTotalUSD decimal.Decimal         `json:"amount_total_usd"`
	AmountTotalCKB decimal.Decimal         `json:"amount_total_ckb"`
	RenewYears     int                     `json:"renew_years"`
}

func EndWithDotBitChar(list []common.AccountCharSet) bool {
	if len(list) <= 4 {
		return false
	}
	accLen := len(list)
	if list[accLen-4].Char == "." && list[accLen-3].Char == "b" && list[accLen-2].Char == "i" && list[accLen-1].Char == "t" {
		return true
	}
	return false
}

func AccountCharSetListToMoleculeAccountChars(list []common.AccountCharSet) molecule.AccountChars {
	accountChars := molecule.NewAccountCharsBuilder()
	for _, item := range list {
		if item.Char == "." {
			break
		}
		accountChar := molecule.NewAccountCharBuilder().
			CharSetName(molecule.GoU32ToMoleculeU32(uint32(item.CharSetName))).
			Bytes(molecule.GoBytes2MoleculeBytes([]byte(item.Char))).Build()
		accountChars.Push(accountChar)
	}
	return accountChars.Build()
}

type OrderStatus int

const (
	OrderStatusDefault OrderStatus = 0
	OrderStatusClosed  OrderStatus = 1
)

type RegisterStatus int

const (
	RegisterStatusDefault         RegisterStatus = 0
	RegisterStatusConfirmPayment  RegisterStatus = 1
	RegisterStatusApplyRegister   RegisterStatus = 2
	RegisterStatusPreRegister     RegisterStatus = 3
	RegisterStatusProposal        RegisterStatus = 4
	RegisterStatusConfirmProposal RegisterStatus = 5
	RegisterStatusRegistered      RegisterStatus = 6
)

func FormatRegisterStatusToSearchStatus(status RegisterStatus) SearchStatus {
	switch status {
	case RegisterStatusRegistered:
		return SearchStatusRegistered
	case RegisterStatusConfirmProposal:
		return SearchStatusConfirmProposal
	case RegisterStatusProposal:
		return SearchStatusProposal
	case RegisterStatusPreRegister:
		return SearchStatusRegistering
	case RegisterStatusApplyRegister:
		return SearchStatusLockedAccount
	case RegisterStatusConfirmPayment:
		return SearchStatusPaymentConfirm
	}
	return SearchStatusRegisterAble
}
