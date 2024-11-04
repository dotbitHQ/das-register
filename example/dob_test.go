package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestDobAccountSearch(t *testing.T) {
	req := handle.ReqAccountSearch{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
		Account:        "2024110401.bit",
		AccountCharStr: nil,
	}

	url := TestUrl + "/account/search"
	var data handle.RespAccountSearch
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobOrderRegister(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainTypeAddress: core.ChainTypeAddress{
				Type: "blockchain",
				KeyInfo: core.KeyInfo{
					CoinType: common.CoinTypeCKB,
					Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
				},
			},
			Account:        "2024110401.bit",
			AccountCharStr: nil,
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears: 1,
		},
		PayTokenId: tables.TokenIdCkbCCC,
	}
	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobRegisteringList(t *testing.T) {
	req := handle.ReqRegisteringList{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
	}
	url := TestUrl + "/account/registering/list"
	var data handle.RespRegisteringList
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobOrderChange(t *testing.T) {
	req := handle.ReqOrderChange{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
		Account:    "2024110401.bit",
		PayTokenId: tables.TokenIdPol,
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  2,
			InviterAccount: "",
			ChannelAccount: "",
		},
	}

	url := TestUrl + "/account/order/change"
	var data handle.RespOrderRegister
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobAccountDetail(t *testing.T) {

}

func TestDobOrderDetail(t *testing.T) {

}

//
func TestDobRewardsMine(t *testing.T) {

}

func TestDobTransactionList(t *testing.T) {

}

func TestDobTransactionStatus(t *testing.T) {

}

func TestDobDidCellRenew(t *testing.T) {

}

func TestDobOrderPayHash(t *testing.T) {

}

func TestDobCheckCoupon(t *testing.T) {

}

//
func TestDobAccountRegister(t *testing.T) {

}
func TestDobAccountRenew(t *testing.T) {

}
