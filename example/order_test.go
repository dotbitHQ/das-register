package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/parnurzeal/gorequest"
	"testing"
)

func TestAccountSearch(t *testing.T) {
	dc, _ := getNewDasCoreTestnet2()

	configRelease, err := dc.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsRelease)
	if err != nil {
		t.Fatal(err)
	}
	luckyNumber, _ := configRelease.LuckyNumber()
	fmt.Println("config release lucky number: ", luckyNumber)
	if resNum, _ := handle.Blake256AndFourBytesBigEndian([]byte("12as.bit")); resNum < luckyNumber {
		t.Fatal("luckyNumber:", resNum)
	}

	var req handle.ReqAccountSearch
	req.Account = "1234.bit"
	req.ChainType = common.ChainTypeEth
	req.Address = "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	accountChars, err := dc.GetAccountCharSetList(req.Account)
	if err != nil {
		t.Fatal(err)
	}

	req.AccountCharStr = accountChars

	url := TestUrl + "/account/search"
	var data handle.RespAccountSearch
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

// mixin: 0xf2011f49d9ad51fe64ce0f03afcff509e0324a046d8ef9b509805678fd2d9254e1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4
// mixin: 0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4
// 0x11ebce55b1cc815df4d82e7c387c7428589875f60c01a56dcd7b9589c660081e99c648a7968540a630dc665a676cf90adaeaad923685f03803abd23bc17c5b58
// 0x99c648a7968540a630dc665a676cf90adaeaad923685f03803abd23bc17c5b58
func TestOrderRegister(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainType: common.ChainTypeMixin,
			Address:   "0x99c648a7968540a630dc665a676cf90adaeaad923685f03803abd23bc17c5b58",
			Account:   "1234567873.bit",
			AccountCharStr: []common.AccountCharSet{
				{CharSetName: common.AccountCharTypeDigit, Char: "1"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "4"},
				{CharSetName: common.AccountCharTypeDigit, Char: "5"},
				{CharSetName: common.AccountCharTypeDigit, Char: "6"},
				{CharSetName: common.AccountCharTypeDigit, Char: "7"},
				{CharSetName: common.AccountCharTypeDigit, Char: "8"},
				{CharSetName: common.AccountCharTypeDigit, Char: "7"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				//{CharSetName: common.AccountCharTypeEn, Char: "a"},
				{CharSetName: common.AccountCharTypeEn, Char: "."},
				{CharSetName: common.AccountCharTypeEn, Char: "b"},
				{CharSetName: common.AccountCharTypeEn, Char: "i"},
				{CharSetName: common.AccountCharTypeEn, Char: "t"},
			},
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  1,
			InviterAccount: "9caaaaaa.bit",
			ChannelAccount: "9caaaaaa.bit",
		},
		PayChainType: 0,
		PayAddress:   "",
		PayTokenId:   tables.TokenIdDas,
		PayType:      "",
	}

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestOrderDetail(t *testing.T) {
	req := handle.ReqOrderDetail{
		ChainType: common.ChainTypeEth,
		Address:   "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
		Account:   "11111111.bit",
		Action:    common.DasActionRenewAccount,
	}
	url := TestUrl + "/account/order/detail"
	var data handle.RespOrderDetail
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestOrderRenew(t *testing.T) {
	req := handle.ReqOrderRenew{
		ChainType:    common.ChainTypeMixin,
		Address:      "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4",
		Account:      "1234567871.bit",
		PayChainType: common.ChainTypeCkb,
		PayTokenId:   tables.TokenIdDas,
		PayAddress:   "",
		PayType:      "",
		RenewYears:   1,
	}
	url := TestUrl + "/account/order/renew"
	var data handle.RespOrderRenew
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}

	fmt.Println(data)
}

func TestOrderOrderChange(t *testing.T) {
	req := handle.ReqOrderChange{
		ChainType: common.ChainTypeEth,
		Address:   "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
		Account:   "1234567885.bit",
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  2,
			InviterAccount: "",
			ChannelAccount: "",
		},
		PayChainType: 0,
		PayAddress:   "",
		PayTokenId:   tables.TokenIdMatic,
		PayType:      "",
	}

	url := TestUrl + "/account/order/change"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestCkbRpc(t *testing.T) {
	req := handle.ReqCkbRpc{
		ID:      1,
		JsonRpc: "2.0",
		Method:  "get_blockchain_info",
		Params:  nil,
	}

	url := TestUrl + "/node/ckb/rpc"
	_, body, errs := gorequest.New().Post(url).SendStruct(&req).End()
	if errs != nil {
		t.Fatal(errs)
	}
	fmt.Println(body)
}

func TestRegisteringList(t *testing.T) {
	req := handle.ReqRegisteringList{
		ChainType: common.ChainTypeEth,
		Address:   "0xc9f53b1d85356b60453f867610888d89a0b667ad",
	}
	url := TestUrl + "/account/registering/list"
	var data handle.RespRegisteringList
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestRegisterByDogecoin(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainType: common.ChainTypeDogeCoin,
			Address:   "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6",
			Account:   "20230301.bit",
			AccountCharStr: []common.AccountCharSet{
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
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
		PayTokenId: tables.TokenIdDoge,
	}

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
