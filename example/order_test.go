package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/parnurzeal/gorequest"
	"testing"
)

func TestAccountSearch(t *testing.T) {
	var req handle.ReqAccountSearch
	req.Account = "7aaaaaaa.bit"
	req.ChainType = common.ChainTypeEth
	req.Address = "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	req.AccountCharStr = []tables.AccountCharSet{
		//{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		//{CharSetName: tables.AccountCharTypeNumber, Char: "1"},
		{CharSetName: tables.AccountCharTypeNumber, Char: "7"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "a"},
		{CharSetName: tables.AccountCharTypeEn, Char: "."},
		{CharSetName: tables.AccountCharTypeEn, Char: "b"},
		{CharSetName: tables.AccountCharTypeEn, Char: "i"},
		{CharSetName: tables.AccountCharTypeEn, Char: "t"},
	}

	url := TestUrl + "/account/search"
	var data handle.RespAccountSearch
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func TestOrderRegister(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainType: common.ChainTypeTron,
			Address:   "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			Account:   "1234567885.bit",
			AccountCharStr: []tables.AccountCharSet{
				{CharSetName: tables.AccountCharTypeNumber, Char: "1"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "2"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "3"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "4"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "5"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "6"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "7"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "8"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "8"},
				{CharSetName: tables.AccountCharTypeNumber, Char: "5"},
				//{CharSetName: tables.AccountCharTypeEn, Char: "a"},
				{CharSetName: tables.AccountCharTypeEn, Char: "."},
				{CharSetName: tables.AccountCharTypeEn, Char: "b"},
				{CharSetName: tables.AccountCharTypeEn, Char: "i"},
				{CharSetName: tables.AccountCharTypeEn, Char: "t"},
			},
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  1,
			InviterAccount: "9caaaaaa.bit",
			ChannelAccount: "9caaaaaa.bit",
		},
		PayChainType: 0,
		PayAddress:   "",
		PayTokenId:   tables.TokenIdTrx,
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
		ChainType:    common.ChainTypeEth,
		Address:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
		Account:      "11111111.bit",
		PayChainType: common.ChainTypeEth,
		PayTokenId:   tables.TokenIdBnb,
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

func TestAccountCharSet(t *testing.T) {
	account := "0123456789abcdefghijklmnopqrstuvwxyz.bit"
	fmt.Println(handle.AccountToCharSet(account))
}