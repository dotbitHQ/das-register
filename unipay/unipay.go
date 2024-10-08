package unipay

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/shopspring/decimal"
	"sync"
)

var (
	log = logger.NewLogger("unipay", logger.LevelDebug)
)

const (
	BusinessIdDasRegisterSvr = "das-register-svr"
)

type ReqOrderCreate struct {
	core.ChainTypeAddress
	BusinessId        string            `json:"business_id"`
	Amount            decimal.Decimal   `json:"amount"`
	PayTokenId        tables.PayTokenId `json:"pay_token_id"`
	PaymentAddress    string            `json:"payment_address"`
	PremiumPercentage decimal.Decimal   `json:"premium_percentage"`
	PremiumBase       decimal.Decimal   `json:"premium_base"`
	PremiumAmount     decimal.Decimal   `json:"premium_amount"`
	MetaData          map[string]string `json:"meta_data"`
}

type RespOrderCreate struct {
	OrderId               string `json:"order_id"`
	PaymentAddress        string `json:"payment_address"`
	ContractAddress       string `json:"contract_address"`
	StripePaymentIntentId string `json:"stripe_payment_intent_id"`
	ClientSecret          string `json:"client_secret"`
}

func CreateOrder(req ReqOrderCreate) (resp RespOrderCreate, err error) {
	url := fmt.Sprintf("%s/v1/order/create", config.Cfg.Server.UniPayUrl)
	err = http_api.SendReq(url, &req, &resp)
	return
}

type RefundInfo struct {
	OrderId string `json:"order_id"`
	PayHash string `json:"pay_hash"`
}

type ReqOrderRefund struct {
	BusinessId string       `json:"business_id"`
	RefundList []RefundInfo `json:"refund_list"`
}

type RespOrderRefund struct {
}

func RefundOrder(req ReqOrderRefund) (resp RespOrderRefund, err error) {
	url := fmt.Sprintf("%s/v1/order/refund", config.Cfg.Server.UniPayUrl)
	err = http_api.SendReq(url, &req, &resp)
	return
}

type ReqPaymentInfo struct {
	BusinessId  string   `json:"business_id"`
	OrderIdList []string `json:"order_id_list"`
	PayHashList []string `json:"pay_hash_list"`
}

type RespPaymentInfo struct {
	PaymentList []PaymentInfo `json:"payment_list"`
}

type PaymentInfo struct {
	OrderId       string                    `json:"order_id"`
	PayHash       string                    `json:"pay_hash"`
	PayAddress    string                    `json:"pay_address"`
	AlgorithmId   common.DasAlgorithmId     `json:"algorithm_id"`
	PayHashStatus tables.PayHashStatus      `json:"pay_hash_status"`
	RefundStatus  tables.UniPayRefundStatus `json:"refund_status"`
	RefundHash    string                    `json:"refund_hash"`
}

func GetPaymentInfo(req ReqPaymentInfo) (resp RespPaymentInfo, err error) {
	url := fmt.Sprintf("%s/v1/payment/info", config.Cfg.Server.UniPayUrl)
	err = http_api.SendReq(url, &req, &resp)
	return
}

type ReqOrderInfo struct {
	BusinessId string `json:"business_id"`
	OrderId    string `json:"order_id"`
}

type RespOrderInfo struct {
	OrderId         string `json:"order_id"`
	PaymentAddress  string `json:"payment_address"`
	ContractAddress string `json:"contract_address"`
	ClientSecret    string `json:"client_secret"`
}

func GetOrderInfo(req ReqOrderInfo) (resp RespOrderInfo, err error) {
	url := fmt.Sprintf("%s/v1/order/info", config.Cfg.Server.UniPayUrl)
	err = http_api.SendReq(url, &req, &resp)
	return
}

//

type ToolUniPay struct {
	Ctx   context.Context
	Wg    *sync.WaitGroup
	DbDao *dao.DbDao
}

func RoundAmount(amount decimal.Decimal, tokenId tables.PayTokenId) decimal.Decimal {
	switch tokenId {
	case tables.TokenIdEth, tables.TokenIdBnb, tables.TokenIdPol: //, tables.TokenIdMatic:
		dec := decimal.New(1, 8)
		amount = amount.Div(dec).Ceil().Mul(dec)
	case tables.TokenIdDas, tables.TokenIdCkb, tables.TokenIdCkbInternal, tables.TokenIdDoge, tables.TokenIdPadgeInternal:
		dec := decimal.New(1, 4)
		amount = amount.Div(dec).Ceil().Mul(dec)
	case tables.TokenIdTrx, tables.TokenIdErc20USDT,
		tables.TokenIdBep20USDT, tables.TokenIdTrc20USDT:
		dec := decimal.New(1, 3)
		amount = amount.Div(dec).Ceil().Mul(dec)
	}
	return amount
}
