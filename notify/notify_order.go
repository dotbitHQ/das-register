package notify

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
	"time"
)

type SendLarkOrderNotifyParam struct {
	Key        string
	Action     string
	Account    string
	OrderId    string
	ChainType  common.ChainType
	Address    string
	PayTokenId tables.PayTokenId
	Amount     decimal.Decimal
}

func SendLarkOrderNotify(p *SendLarkOrderNotifyParam) {
	msg := `> account: %s
> order_id: %s
> address: %s
> pay token id: %s
> amount: %s
> time: %s`
	address := fmt.Sprintf("(%s)%s", p.ChainType.ToString(), p.Address)
	amount := p.Amount
	switch p.PayTokenId {
	case tables.TokenIdBnb, tables.TokenIdEth, tables.TokenIdMatic:
		amount = amount.DivRound(decimal.New(1, 18), 18)
	case tables.TokenIdCkb, tables.TokenIdDas:
		amount = amount.DivRound(decimal.New(1, 8), 8)
	case tables.TokenIdTrx:
		amount = amount.DivRound(decimal.New(1, 6), 6)
	case tables.TokenIdWx:
		amount = amount.DivRound(decimal.New(1, 2), 2)
	case tables.TokenIdDoge:
		amount = amount.DivRound(decimal.New(1, 8), 8)
	case tables.TokenIdStripeUSD:
		amount = amount.DivRound(decimal.New(1, 2), 2)
	}
	msg = fmt.Sprintf(msg, p.Account, p.OrderId, address, p.PayTokenId, amount.String(), time.Now().Format("2006-01-02 15:04:05"))
	SendLarkTextNotify(p.Key, p.Action, msg)
}
