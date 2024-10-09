package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"testing"
)

func TestRegisterByDogecoin(t *testing.T) {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainType: common.ChainTypeDogeCoin,
			Address:   "DP86MSmWjEZw8GKotxcvAaW5D4e3qoEh6f",
			Account:   "20230315.bit",
			AccountCharStr: []common.AccountCharSet{
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "2"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "0"},
				{CharSetName: common.AccountCharTypeDigit, Char: "3"},
				{CharSetName: common.AccountCharTypeDigit, Char: "1"},
				{CharSetName: common.AccountCharTypeDigit, Char: "5"},
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

func TestAccountDetail(t *testing.T) {
	req := handle.ReqAccountDetail{Account: "20230301.bit"}
	url := TestUrl + "/account/detail"
	var data handle.RespAccountDetail
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestAccountMine(t *testing.T) {
	req := handle.ReqAccountMine{
		ChainType:  common.ChainTypeDogeCoin,
		Address:    "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6",
		Pagination: handle.Pagination{},
		Keyword:    "",
		Category:   0,
	}
	url := TestUrl + "/account/mine"
	var data handle.RespAccountMine
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestEditRecords2(t *testing.T) {
	url := TestUrl + "/account/edit/records"

	var req handle.ReqEditRecords
	req.ChainType = common.ChainTypeDogeCoin
	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.Account = "20230301.bit"
	req.RawParam.Records = []handle.ReqRecord{{
		Key:   "3",
		Type:  "address",
		Label: "",
		Value: "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6",
		TTL:   "300",
	}}
	req.EvmChainId = 0

	var data handle.RespEditRecords

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	//fmt.Println(fmt.Sprintf(`curl -X POST http://127.0.0.1:8119/v1/sign/tx -d'{"private":"%s","sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":{}}'`,private, data.SignKey, data.SignList[0].SignType, data.SignList[0].SignMsg))
	fmt.Println("signMasg:", data.SignList[0].SignMsg)
	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))
	// curl -X POST http://127.0.0.1:8119/v1/sign/tx -d'{"private":"","sign_key":"","sign_list":[{"sign_type":5,"sign_msg":"0xad9bc80a25d3753354f085074c271477db309dff75b15f7ab4db9e192a7cf768"}],"mm_json":{}}'
}

var privateKey = ""

//signMasg: 0x537fbc582f5e4510997c1a122550f38845bb4fda275094daf996ce04dae3a33e
//magicHash: 0x04e213c8d9f322f973dcc6583763af047c0c75f34271ad5cfd6f69269bbbac30
//DogeSignature: 0x950025f61992dc05db20d4f593ac93b6c32427595d31e3aa8807080ce77e721b01c1a7107fd57617075fde3b289cb420ffc0e6d47cdb6b0b46b044e2c3162eb301
//{"sign_key":"76914f3c0c1421612b8bca4f8681fa8c","sign_list":[{"sign_type":7,"sign_msg":"0x950025f61992dc05db20d4f593ac93b6c32427595d31e3aa8807080ce77e721b01c1a7107fd57617075fde3b289cb420ffc0e6d47cdb6b0b46b044e2c3162eb30100"}],"mm_json":null}
func TestTransactionSend2(t *testing.T) {
	str := `{"sign_key":"2de6cbcae5f5ab3cba6a37c40afcedef","sign_list":[{"sign_type":5,"sign_msg":"0x3b714893f7eb7d471c168e6d72b6a3ecf8b0329eebf3a0919e8d8b560a2c682b2a577669915990ddfdf132b225c753f7c11aa52f628ac804db37852d756f4e701b3aa8ab557764c6508a5e0de31b5f01cddbc4e9ff544095dc64ffd0ddffea0dee0000000000000005"}],"mm_json":null}`

	var req handle.ReqTransactionSend
	if err := json.Unmarshal([]byte(str), &req); err != nil {
		t.Fatal(err)
	}
	url := TestUrl + "/transaction/send"
	var data handle.RespTransactionSend
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data.Hash)
}

func TestEditManager2(t *testing.T) {
	url := TestUrl + "/account/edit/manager"

	var req handle.ReqEditManager
	req.ChainType = common.ChainTypeDogeCoin
	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.Account = "20230301.bit"
	req.RawParam.ManagerChainType = common.ChainTypeEth
	req.RawParam.ManagerAddress = "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
	var data handle.RespEditManager

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println("signMsg:", data.SignList[0].SignMsg)
	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))

}

func TestEditOwner2(t *testing.T) {
	url := TestUrl + "/account/edit/owner"

	var req handle.ReqEditOwner
	req.ChainType = common.ChainTypeDogeCoin
	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.Account = "20230301.bit"
	req.RawParam.ReceiverChainType = common.ChainTypeEth
	req.RawParam.ReceiverAddress = "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
	req.EvmChainId = 97

	var data handle.RespEditOwner

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println("signMsg:", data.SignList[0].SignMsg)
	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))
}

//
//func TestReverse(t *testing.T) {
//	url := TestUrl + "/reverse/declare"
//	var req handle.ReqReverseDeclare
//	req.ChainType = common.ChainTypeDogeCoin
//	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
//	req.Account = "20230301.bit"
//	req.EvmChainId = 5
//
//	var data handle.RespReverseDeclare
//	if err := doReq(url, req, &data); err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("signMsg:", data.SignList[0].SignMsg)
//	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))
//}
//
//func TestReverseRedeclare2(t *testing.T) {
//	url := TestUrl + "/reverse/redeclare"
//	var req handle.ReqReverseRedeclare
//	req.ChainType = common.ChainTypeDogeCoin
//	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
//	req.Account = "20230315.bit"
//	req.EvmChainId = 5
//
//	var data handle.RespReverseRedeclare
//	if err := doReq(url, req, &data); err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("signMsg:", data.SignList[0].SignMsg)
//	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))
//}
//
//func TestReverseRetract2(t *testing.T) {
//	var req handle.ReqReverseRetract
//	req.ChainType = common.ChainTypeDogeCoin
//	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
//	req.EvmChainId = 5
//	url := TestUrl + "/reverse/retract"
//
//	var data handle.RespReverseRetract
//	if err := doReq(url, req, &data); err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("signMsg:", data.SignList[0].SignMsg)
//	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))
//
//}

func TestBalanceWithdraw2(t *testing.T) {
	url := TestUrl + "/balance/withdraw"

	var req handle.ReqBalanceWithdraw
	req.ChainType = common.ChainTypeDogeCoin
	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.ReceiverAddress = "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qgrzk3ntzys3nuwmvnar2lrs54l9pat6wy3qv26xdvgjzx03mdj05dtuwzjhu5840fcjyg4hdja"
	req.Amount = decimal.NewFromInt(50000000000)
	req.EvmChainId = 5

	var data handle.RespBalanceWithdraw

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println("signMsg:", data.SignList[0].SignMsg)
	signData, err := sign.DogeSignature(common.Hex2Bytes(data.SignList[0].SignMsg), privateKey, false)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`{"sign_key":"%s","sign_list":[{"sign_type":%d,"sign_msg":"%s"}],"mm_json":null}`, data.SignKey, data.SignList[0].SignType, common.Bytes2Hex(signData)))

}
