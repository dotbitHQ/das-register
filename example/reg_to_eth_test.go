package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"testing"
)

func TestRegToEth(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			//ChainType: common.ChainTypeEth,
			//Address:   "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
			Account: "tzh20220913-03.bit",
			AccountCharStr: []common.AccountCharSet{
				{CharSetName: common.AccountCharTypeEn, Char: "t"},
				{CharSetName: common.AccountCharTypeEn, Char: "z"},
				{CharSetName: common.AccountCharTypeEn, Char: "h"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "9"},
				{CharSetName: common.AccountCharTypeDigit, Char: "1"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "-"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				//{CharSetName: common.AccountCharTypeEn, Char: "a"},
				//{CharSetName: common.AccountCharTypeEn, Char: "."},
				//{CharSetName: common.AccountCharTypeEn, Char: "b"},
				//{CharSetName: common.AccountCharTypeEn, Char: "i"},
				//{CharSetName: common.AccountCharTypeEn, Char: "t"},
			},
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  1,
			InviterAccount: "0001.bit",
			ChannelAccount: "0001.bit",
		},

		PayChainType: 0,
		PayAddress:   "",
		PayTokenId:   tables.TokenIdDas,
		PayType:      "",
		CoinType:     "60",
		//CrossCoinType: "60",
	}

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
