package tables

import (
	"github.com/shopspring/decimal"
)

type TableTokenPriceInfo struct {
	Id        uint64          `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	TokenId   PayTokenId      `json:"token_id" gorm:"column:token_id"`
	GeckoId   string          `json:"gecko_id" gorm:"column:gecko_id"`
	ChainType int             `json:"chain_type" gorm:"column:chain_type"`
	Name      string          `json:"name" gorm:"column:name"`
	Symbol    string          `json:"symbol" gorm:"column:symbol"`
	Decimals  int32           `json:"decimals" gorm:"column:decimals"`
	Price     decimal.Decimal `json:"price" gorm:"column:price"`
	Logo      string          `json:"logo" gorm:"column:logo"`
	Status    int             `json:"status" gorm:"column:status"`
}

const (
	TableNameTokenPriceInfo = "t_token_price_info"
)

func (t *TableTokenPriceInfo) TableName() string {
	return TableNameTokenPriceInfo
}
