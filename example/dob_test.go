package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"testing"
)

func TestDobAccountSearch(t *testing.T) {
	req := handle.ReqAccountSearch{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "",
			},
		},
		Account:        "",
		AccountCharStr: nil,
	}

	url := TestUrl + "/account/search"
	var data handle.RespAccountSearch
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
}

func TestDobOrderRegister(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainTypeAddress: core.ChainTypeAddress{
				Type: "blockchain",
				KeyInfo: core.KeyInfo{
					CoinType: common.CoinTypeCKB,
					Key:      "",
				},
			},
			Account:        "",
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

}

func TestDobOrderChange(t *testing.T) {

}

func TestDobAccountDetail(t *testing.T) {

}

func TestDobOrderDetail(t *testing.T) {

}

func TestDobRegisteringList(t *testing.T) {

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
