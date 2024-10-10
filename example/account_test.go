package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

const (
	TestUrl = "https://test-register-api.did.id/v1"
)

func TestTransactionSend(t *testing.T) {
	str := `{"sign_key":"f4ecb014be94784065eab77bdd25b47d","sign_list":[{"sign_type":7,"sign_msg":"0x78a66e6fa283f5da55c98e917647fe783931bd0a22e16aadcad32711f65a06493e29a3aa41c753ddd476a228cd1d7074f261076c77686849d45e2fba1f1b0b850100"}],"mm_json":null}`

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

func doReq(url string, req, data interface{}) error {
	var resp api_code.ApiResp
	resp.Data = &data

	_, _, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&resp)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	if resp.ErrNo != api_code.ApiCodeSuccess {
		return fmt.Errorf("%d - %s", resp.ErrNo, resp.ErrMsg)
	}
	return nil
}

func TestEditManager(t *testing.T) {
	url := TestUrl + "/account/edit/manager"

	var req handle.ReqEditManager
	req.ChainType = common.ChainTypeEth
	req.Address = "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
	req.Account = "tzh2022070401.bit"
	req.RawParam.ManagerChainType = common.ChainTypeDogeCoin
	req.RawParam.ManagerAddress = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	var data handle.RespEditManager

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}

	var signReq handle.ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 5
	signReq.Private = ""
	fmt.Println(toolib.JsonString(signReq))

	// curl -X POST http://127.0.0.1:8119/v1/sign/tx -d'{"chain_id":5,"private":"","sign_key":"87c906d6121d7eeea5670a247600cd4c","sign_list":[{"sign_type":5,"sign_msg":"0xad9bc80a25d3753354f085074c271477db309dff75b15f7ab4db9e192a7cf768"}],"mm_json":{"types":{"EIP712Domain":[{"name":"chainId","type":"uint256"},{"name":"name","type":"string"},{"name":"verifyingContract","type":"address"},{"name":"version","type":"string"}],"Action":[{"name":"action","type":"string"},{"name":"params","type":"string"}],"Cell":[{"name":"capacity","type":"string"},{"name":"lock","type":"string"},{"name":"type","type":"string"},{"name":"data","type":"string"},{"name":"extraData","type":"string"}],"Transaction":[{"name":"DAS_MESSAGE","type":"string"},{"name":"inputsCapacity","type":"string"},{"name":"outputsCapacity","type":"string"},{"name":"fee","type":"string"},{"name":"action","type":"Action"},{"name":"inputs","type":"Cell[]"},{"name":"outputs","type":"Cell[]"},{"name":"digest","type":"bytes32"}]},"primaryType":"Transaction","domain":{"chainId":5,"name":"da.systems","verifyingContract":"0x0000000000000000000000000000000020210722","version":"1"},"message":{"DAS_MESSAGE":"EDIT MANAGER OF ACCOUNT 0001.bit","inputsCapacity":"214.9995 CKB","outputsCapacity":"214.9994 CKB","fee":"0.0001 CKB","digest":"","action":{"action":"edit_manager","params":"0x00"},"inputs":[{"capacity":"214.9995 CKB","lock":"das-lock,0x01,0x05c9f53b1d85356b60453f867610888d89a0b667...","type":"account-cell-type,0x01,0x","data":"{ account: 0001.bit, expired_at: 1822199174 }","extraData":"{ status: 0, records_hash: 0x5376adbb69986cf8192a1ab94fe438920e2046f1b450ef9af5a8ad0902890e28 }"}],"outputs":[{"capacity":"214.9994 CKB","lock":"das-lock,0x01,0x05c9f53b1d85356b60453f867610888d89a0b667...","type":"account-cell-type,0x01,0x","data":"{ account: 0001.bit, expired_at: 1822199174 }","extraData":"{ status: 0, records_hash: 0x5376adbb69986cf8192a1ab94fe438920e2046f1b450ef9af5a8ad0902890e28 }"}]}}}'
}

func TestEditOwner(t *testing.T) {
	url := TestUrl + "/account/edit/owner"

	var req handle.ReqEditOwner
	req.ChainType = common.ChainTypeTron
	req.Address = "TQoLh9evwUmZKxpD1uhFttsZk3EBs8BksV"
	req.Account = "account2022011906.bit"
	req.RawParam.ReceiverChainType = common.ChainTypeMixin
	req.RawParam.ReceiverAddress = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"
	req.EvmChainId = 97

	var data handle.RespEditOwner

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}

	var signReq handle.ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 97
	signReq.Private = ""
	fmt.Println(toolib.JsonString(signReq))
	// curl -X POST http://127.0.0.1:8119/v1/sign/tx
}

func TestEditRecords(t *testing.T) {
	url := TestUrl + "/account/edit/records"

	var req handle.ReqEditRecords
	req.ChainType = common.ChainTypeEth
	req.Address = "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	req.Account = "tzh20220718.bit"
	req.RawParam.Records = []handle.ReqRecord{{
		Key:   "60",
		Type:  "address",
		Label: "",
		Value: "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
		TTL:   "300",
	}}
	req.EvmChainId = 5

	var data handle.RespEditRecords

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	var signReq handle.ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 5
	signReq.Private = ""
	fmt.Println(toolib.JsonString(signReq))
	// curl -X POST http://127.0.0.1:8119/v1/sign/tx
}

func TestBalancePay(t *testing.T) {
	url := TestUrl + "/balance/pay"
	var req handle.ReqBalancePay
	req.EvmChainId = 5
	req.OrderId = "8cd2b0df823d243dc54b1bcb61a0e6cd"
	req.ChainType = common.ChainTypeMixin
	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"

	var data handle.RespBalancePay

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	var signReq handle.ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 5
	signReq.Private = ""
	fmt.Println(toolib.JsonString(signReq))
	// curl -X POST http://127.0.0.1:8119/v1/sign/tx
}

//
//func TestReverseDeclare(t *testing.T) {
//	url := TestUrl + "/reverse/declare"
//	var req handle.ReqReverseDeclare
//	req.ChainType = common.ChainTypeMixin
//	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"
//	req.Account = "1234567871.bit"
//	req.EvmChainId = 5
//
//	var data handle.RespReverseDeclare
//	if err := doReq(url, req, &data); err != nil {
//		t.Fatal(err)
//	}
//	var signReq handle.ReqSignTx
//	signReq.SignInfo = data.SignInfo
//	signReq.ChainId = 5
//	signReq.Private = ""
//	fmt.Println(toolib.JsonString(signReq))
//}
//
//func TestReverseRedeclare(t *testing.T) {
//	url := TestUrl + "/reverse/redeclare"
//	var req handle.ReqReverseRedeclare
//	req.ChainType = common.ChainTypeMixin
//	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"
//	req.Account = "1234567872.bit"
//	req.EvmChainId = 5
//
//	var data handle.RespReverseRedeclare
//	if err := doReq(url, req, &data); err != nil {
//		t.Fatal(err)
//	}
//	var signReq handle.ReqSignTx
//	signReq.SignInfo = data.SignInfo
//	signReq.ChainId = 5
//	signReq.Private = ""
//	fmt.Println(toolib.JsonString(signReq))
//}
//
//func TestReverseRetract(t *testing.T) {
//	var req handle.ReqReverseRetract
//	req.ChainType = common.ChainTypeMixin
//	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"
//	req.EvmChainId = 5
//	url := TestUrl + "/reverse/retract"
//
//	var data handle.RespReverseRetract
//	if err := doReq(url, req, &data); err != nil {
//		t.Fatal(err)
//	}
//	var signReq handle.ReqSignTx
//	signReq.SignInfo = data.SignInfo
//	signReq.ChainId = 5
//	signReq.Private = ""
//	fmt.Println(toolib.JsonString(signReq))
//}

func TestBalanceInfo(t *testing.T) {
	url := TestUrl + "/balance/info"
	var req handle.ReqBalanceInfo
	req.ChainType = common.ChainTypeMixin
	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"

	var data handle.RespBalanceInfo
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(data))
}

func TestReverseLatest(t *testing.T) {
	url := TestUrl + "/reverse/latest"
	var req handle.ReqReverseLatest
	req.ChainType = common.ChainTypeMixin
	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"
	var data handle.RespReverseLatest
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(data))
}

func TestAccountList(t *testing.T) {
	url := TestUrl + "/account/mine"
	var req handle.ReqAccountMine
	req.ChainType = common.ChainTypeEth
	req.Address = "0x15a33588908cf8edb27d1abe3852bf287abd3891"
	req.Keyword = "zz"
	req.Size = 10
	var data handle.RespAccountMine
	for i := 0; i < 3; i++ {
		go func(index int) {
			req.Keyword = fmt.Sprintf("%d", index)
			if err := doReq(url, req, &data); err != nil {
				t.Fatal(err)
			}
			fmt.Println(toolib.JsonString(data))
		}(i)
	}
	time.Sleep(time.Second * 2)
}

//curl -X POST http://127.0.0.1:8119/v1/account/register -d'{"chain_type":1,"address":"0xc9f53b1d85356B60453F867610888D89a0B667Ad","account":"tzh20220620-01.bit","register_years":1,"inviter_account":"","channel_account":""}'

func TestTimestamp(t *testing.T) {
	fmt.Println(time.Now().Unix(), time.Now().UnixNano())
}

// curl -X POST http://127.0.0.1:8119/v1/sign/tx -d'{"chain_id":5,"private":"",}'

func TestEditScript(t *testing.T) {
	args := common.Bytes2Hex(make([]byte, 33))
	args = "0x01f15f519ecb226cd763b2bcbcab093e63f89100c07ac0caebc032c788b187ec99"
	fmt.Println(args)
	url := TestUrl + "/account/edit/script"
	req := handle.ReqEditScript{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				ChainId:  "",
				Key:      "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
			},
		},
		Account:          "00acc2022042902.bit",
		CustomScriptArgs: args,
		EvmChainId:       5,
	}
	var data handle.RespEditScript
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestOrderRegister2(t *testing.T) {
	str := `{"chain_type":1,"address":"0x020881e3f5b7832e752d16fe2710ee855a6977dc","account":"faksjfd3lakj.bit","account_char_str":[{"char_set_name":2,"char":"f"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"k"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"j"},{"char_set_name":2,"char":"f"},{"char_set_name":2,"char":"d"},{"char_set_name":1,"char":"3"},{"char_set_name":2,"char":"l"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"k"},{"char_set_name":2,"char":"j"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}],"register_years":1,"inviter_account":"xiand.bit","channel_account":"cryptofans.bit","pay_chain_type":0,"pay_address":"0x020881e3f5b7832e752d16fe2710ee855a6977dc","pay_token_id":"ckb_das","pay_type":""}`
	var req handle.ReqOrderRegister
	json.Unmarshal([]byte(str), &req)
	fmt.Println(req.ChannelAccount)

	orderContent := tables.TableOrderContent{
		AccountCharStr: req.AccountCharStr,
		InviterAccount: req.InviterAccount,
		ChannelAccount: req.ChannelAccount,
		RegisterYears:  req.RegisterYears,
		AmountTotalUSD: decimal.Zero,
		AmountTotalCKB: decimal.Zero,
	}

	contentDataStr, _ := json.Marshal(&orderContent)
	fmt.Println(string(contentDataStr))
}

func TestEndWithDotBitChar(t *testing.T) {
	str := `[{"char_set_name":9,"char":"บั๋"},{"char_set_name":9,"char":"บั๋"},{"char_set_name":9,"char":"บั๋"},{"char_set_name":9,"char":"บั๋"},{"char_set_name":9,"char":"บั๋"},{"char_set_name":9,"char":"บั๋"},{"char_set_name":9,"char":"บั๋"}]`

	var list []common.AccountCharSet
	json.Unmarshal([]byte(str), &list)
	fmt.Println(tables.EndWithDotBitChar(list))
}
