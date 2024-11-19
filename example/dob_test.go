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
			RegisterYears:  1,
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

func TestDobOrderDetail(t *testing.T) {
	req := handle.ReqOrderDetail{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
		Account: "2024110401.bit",
		Action:  common.DasActionApplyRegister,
	}
	url := TestUrl + "/account/order/detail"
	var data handle.RespOrderDetail
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobAccountDetail(t *testing.T) {
	req := handle.ReqAccountDetail{Account: "20241104.bit"}
	url := TestUrl + "/account/detail"
	var data handle.RespAccountDetail
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

//
func TestDobRewardsMine(t *testing.T) {
	req := handle.ReqRewardsMine{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
		Pagination: handle.Pagination{
			Page: 1,
			Size: 20,
		},
	}
	url := TestUrl + "/rewards/mine"
	var data handle.RespRewardsMine
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobTransactionList(t *testing.T) {
	//bid_expired_account_dutch_auction
	//confirm_proposal
	//did_recycle
	//recycle_expired_account
	//transfer_account

	//accept_offer
	//buy_account
	//cancel_account_sale
	//cancel_offer
	//make_offer
	//offer_accepted
	//order_refund
	//sale_account
	//start_account_sale
	//transfer
	//transfer_balance
	//withdraw_from_wallet

	//apply_register
	//balance_deposit
	//collect_sub_account_profit
	//config_sub_account_custom_script
	//consolidate_income
	//create_approval
	//create_sub_account
	//cross_refund
	//declare_reverse_record
	//delay_approval
	//did_edit_owner
	//did_edit_records
	//edit_account_sale
	//edit_manager
	//edit_records
	//edit_sub_account
	//enable_sub_account
	//force_recover_account_status
	//fulfill_approval
	//lock_account_for_cross_chain
	//offer_edit_add
	//offer_edit_sub
	//pre_register
	//propose
	//redeclare_reverse_record
	//renew_account
	//renew_sub_account
	//retract_reverse_record
	//revoke_approval
	//unlock_account_for_cross_chain

	req := handle.ReqTransactionList{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
		Pagination: handle.Pagination{
			Page: 1,
			Size: 20,
		},
	}
	url := TestUrl + "/transaction/list"
	var data handle.RespTransactionList
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDobTransactionStatus(t *testing.T) {
	req := handle.ReqTransactionStatus{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgjzk3ntzys3nuwmvnar2lrs54l9pat6wy3qq5glj65",
			},
		},
		Actions: []tables.TxAction{tables.ActionEditRecords},
	}
	url := TestUrl + "/transaction/status"
	var data handle.RespTransactionStatus
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

// todo dob padge mint svr
func TestDobAccountRegister(t *testing.T) {

}
func TestDobAccountRenew(t *testing.T) {
}
