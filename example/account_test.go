package example

import (
	"das_register_server/http_server/api_code"
	"das_register_server/http_server/handle"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"testing"
)

const (
	TestUrl = "https://test-register-api.da.systems/v1"
)

func TestTransactionSend(t *testing.T) {
	str := `{"sign_key":"e8bf9ce3dbba27764d4d2bd65ebf6eda","sign_list":[{"sign_type":5,"sign_msg":"0xb808f090c058173531c641e3602aa2c508e00be7efe9cb5e312e835f5ec201550ed522964e487761e8dd5621072408bef39a31af4dac26902a6019bb4e2dd6861c685e8bfd4df7271d62517c22bf19123638c8d95653b09e3fc72ccad3cec6c0400000000000000005"}],"mm_json":null}`

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
	req.Address = "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	req.Account = "tangzhihong007.bit"
	req.RawParam.ManagerChainType = common.ChainTypeTron
	req.RawParam.ManagerAddress = "TQoLh9evwUmZKxpD1uhFttsZk3EBs8BksV"
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
	req.ChainType = common.ChainTypeEth
	req.Address = "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	req.Account = "tang0001.bit"
	req.RawParam.ReceiverChainType = common.ChainTypeTron
	req.RawParam.ReceiverAddress = "TQoLh9evwUmZKxpD1uhFttsZk3EBs8BksV"
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
	req.ChainType = common.ChainTypeTron
	req.Address = "TQoLh9evwUmZKxpD1uhFttsZk3EBs8BksV"
	req.Account = "tangzhihong007.bit"
	req.RawParam.Records = []handle.ReqRecord{{
		Key:   "twitter",
		Type:  "profile",
		Label: "22",
		Value: "111",
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
	req.OrderId = "bb4a9ed3eb80dcfbc3b28b8028fad362"
	req.ChainType = common.ChainTypeEth
	req.Address = "0xc9f53b1d85356b60453f867610888d89a0b667ad"

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
