package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"sync"
	"testing"
)

func TestCheckAccountCharSet(t *testing.T) {
	_ = getDasCore(t)
	var h HttpHandle
	var req ReqAccountSearch
	reqStr := `{"chain_type":1,"address":"0xa0324794ff56ecb258220046034a363d0da98f51","account":"1111ぁぁぁぁ.bit","account_char_str":[{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"1"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}]}`

	json.Unmarshal([]byte(reqStr), &req)
	var apiResp api_code.ApiResp
	h.checkAccountCharSet(&req, &apiResp)
	fmt.Println(apiResp)
}

func getDasCore(t *testing.T) *core.DasCore {
	ckbUrl := "https://testnet.ckb.dev/"
	indexerUrl := "https://testnet.ckb.dev/indexer"
	// ckb node
	ckbClient, err := rpc.DialWithIndexer(ckbUrl, indexerUrl)
	if err != nil {
		t.Fatal(err)
	}

	// das init
	env := core.InitEnvOpt(common.DasNetTypeTestnet2, common.DasContractNameConfigCellType)
	ops := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(config.Cfg.Server.Net),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	var wg sync.WaitGroup
	dasCore := core.NewDasCore(context.Background(), &wg, ops...)
	dasCore.InitDasContract(env.MapContract)
	if err := dasCore.InitDasConfigCell(); err != nil {
		t.Fatal(err)
	}
	return dasCore
}

const (
	TestUrl = "https://test-register-api.did.id/v1"
)

func TestEditManager(t *testing.T) {
	url := TestUrl + "/account/edit/manager"

	var req ReqEditManager
	req.ChainType = common.ChainTypeDogeCoin
	req.Address = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.Account = "20230301.bit"
	req.RawParam.ManagerChainType = common.ChainTypeDogeCoin
	req.RawParam.ManagerAddress = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.EvmChainId = 5
	var data RespEditManager

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}

	var signReq ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 5
	signReq.Private = ""
	fmt.Println(toolib.JsonString(signReq))

	var handle HttpHandle
	var apiResp api_code.ApiResp
	if err := handle.doSignTx(&signReq, &apiResp); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(apiResp.Data))

	// curl -X POST http://127.0.0.1:8119/v1/sign/tx -d'{"chain_id":5,"private":"","sign_key":"87c906d6121d7eeea5670a247600cd4c","sign_list":[{"sign_type":5,"sign_msg":"0xad9bc80a25d3753354f085074c271477db309dff75b15f7ab4db9e192a7cf768"}],"mm_json":{"types":{"EIP712Domain":[{"name":"chainId","type":"uint256"},{"name":"name","type":"string"},{"name":"verifyingContract","type":"address"},{"name":"version","type":"string"}],"Action":[{"name":"action","type":"string"},{"name":"params","type":"string"}],"Cell":[{"name":"capacity","type":"string"},{"name":"lock","type":"string"},{"name":"type","type":"string"},{"name":"data","type":"string"},{"name":"extraData","type":"string"}],"Transaction":[{"name":"DAS_MESSAGE","type":"string"},{"name":"inputsCapacity","type":"string"},{"name":"outputsCapacity","type":"string"},{"name":"fee","type":"string"},{"name":"action","type":"Action"},{"name":"inputs","type":"Cell[]"},{"name":"outputs","type":"Cell[]"},{"name":"digest","type":"bytes32"}]},"primaryType":"Transaction","domain":{"chainId":5,"name":"da.systems","verifyingContract":"0x0000000000000000000000000000000020210722","version":"1"},"message":{"DAS_MESSAGE":"EDIT MANAGER OF ACCOUNT 0001.bit","inputsCapacity":"214.9995 CKB","outputsCapacity":"214.9994 CKB","fee":"0.0001 CKB","digest":"","action":{"action":"edit_manager","params":"0x00"},"inputs":[{"capacity":"214.9995 CKB","lock":"das-lock,0x01,0x05c9f53b1d85356b60453f867610888d89a0b667...","type":"account-cell-type,0x01,0x","data":"{ account: 0001.bit, expired_at: 1822199174 }","extraData":"{ status: 0, records_hash: 0x5376adbb69986cf8192a1ab94fe438920e2046f1b450ef9af5a8ad0902890e28 }"}],"outputs":[{"capacity":"214.9994 CKB","lock":"das-lock,0x01,0x05c9f53b1d85356b60453f867610888d89a0b667...","type":"account-cell-type,0x01,0x","data":"{ account: 0001.bit, expired_at: 1822199174 }","extraData":"{ status: 0, records_hash: 0x5376adbb69986cf8192a1ab94fe438920e2046f1b450ef9af5a8ad0902890e28 }"}]}}}'
}

func TestEditOwner(t *testing.T) {
	url := TestUrl + "/account/edit/owner"

	var req ReqEditOwner
	req.ChainType = common.ChainTypeEth
	req.Address = "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
	req.Account = "tzh2022070401.bit"
	req.RawParam.ReceiverChainType = common.ChainTypeDogeCoin
	req.RawParam.ReceiverAddress = "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6"
	req.EvmChainId = 5

	var data RespEditOwner

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}

	var signReq ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 5
	signReq.Private = ""
	var handle HttpHandle
	var apiResp api_code.ApiResp
	if err := handle.doSignTx(&signReq, &apiResp); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(apiResp.Data))
	// curl -X POST http://127.0.0.1:8119/v1/sign/tx
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

func TestEditRecords2(t *testing.T) {
	url := TestUrl + "/account/edit/records"

	var req ReqEditRecords
	req.ChainType = common.ChainTypeEth
	req.Address = "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
	req.Account = "20230301.bit"
	req.RawParam.Records = []ReqRecord{{
		Key:   "60",
		Type:  "address",
		Label: "",
		Value: "DMjVFBqbqZGAyTXgkt7fTuqihhCCVuLwZ6",
		TTL:   "300",
	}}
	req.EvmChainId = 0

	var data RespEditRecords

	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	var signReq ReqSignTx
	signReq.SignInfo = data.SignInfo
	signReq.ChainId = 5
	signReq.Private = ""

	var handle HttpHandle
	var apiResp api_code.ApiResp
	if err := handle.doSignTx(&signReq, &apiResp); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(apiResp.Data))
}

func TestReverse(t *testing.T) {
	type ReqReverseUpdate struct {
		core.ChainTypeAddress
		Action  string `json:"action"`
		Account string `json:"account"`
	}

	type RespReverseUpdate struct {
		SignMsg  string                `json:"sign_msg"`
		SignType common.DasAlgorithmId `json:"sign_type"`
		SignKey  string                `json:"sign_key"`
	}
	req := ReqReverseUpdate{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: "60",
				ChainId:  "",
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			},
		},
		Action:  "update",
		Account: "20230301.bit",
	}
	var data RespReverseUpdate

	url := "https://test-reverse-api.did.id/v1/reverse/update"
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println("data:", toolib.JsonString(&data))

	privateKey := ""
	res, err := sign.PersonalSignature(common.Hex2Bytes(data.SignMsg), privateKey) //sign.DogeSignature(common.Hex2Bytes(data.SignMsg), privateKey, false)
	if err != nil {
		t.Fatal(err)
	}

	type ReqReverseSend struct {
		SignKey   string `json:"sign_key"`
		Signature string `json:"signature"`
	}
	req2 := ReqReverseSend{
		SignKey:   data.SignKey,
		Signature: common.Bytes2Hex(res),
	}

	url = "https://test-reverse-api.did.id/v1/reverse/send"
	var str = ""
	if err := doReq(url, req2, str); err != nil {
		t.Fatal(err)
	}

}
