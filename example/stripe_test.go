package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"testing"
)

func TestRegByStripe(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainType: common.ChainTypeEth,
			Address:   "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			Account:   "20230731.bit",
			AccountCharStr: []common.AccountCharSet{
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "7"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "1"},
				{CharSetName: common.AccountCharTypeEn, Char: "."},
				{CharSetName: common.AccountCharTypeEn, Char: "b"},
				{CharSetName: common.AccountCharTypeEn, Char: "i"},
				{CharSetName: common.AccountCharTypeEn, Char: "t"},
			},
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  1,
			InviterAccount: "",
			ChannelAccount: "",
		},
		PayChainType: 0,
		PayAddress:   "",
		PayTokenId:   tables.TokenIdStripeUSD,
		PayType:      "",
	}

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
