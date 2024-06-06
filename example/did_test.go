package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"strings"
	"testing"
)

func sendTx2(sigInfo handle.SignInfo) error {
	var req handle.ReqTransactionSend
	req.SignInfo = sigInfo
	url := TestUrl + "/transaction/send"
	var data handle.RespTransactionSend
	if err := doReq(url, req, &data); err != nil {
		return fmt.Errorf("doReq err: %s", err.Error())
	}
	fmt.Println(toolib.JsonString(&data))
	return nil
}

func TestAccountDetail2(t *testing.T) {
	req := handle.ReqAccountDetail{Account: "web3max004.bit"}
	url := TestUrl + "/account/detail"
	var data handle.RespAccountDetail
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
}

func TestDidCellList(t *testing.T) {
	req := handle.ReqDidCellList{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgytmmrfg7aczevlxngqnr28npj2849erjyqqhe2guh",
			},
		},
		Pagination: handle.Pagination{
			Page: 1,
			Size: 20,
		},
	}
	url := TestUrl + "/did/cell/list"
	var data handle.RespDidCellList
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
}

func TestAccountMine2(t *testing.T) {
	req := handle.ReqAccountMine{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			},
		},
		Pagination: handle.Pagination{
			Page: 1,
			Size: 20,
		},
		Keyword:  "202405",
		Category: 0,
	}
	url := TestUrl + "/account/mine"
	var data handle.RespAccountMine
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestAccountRecords(t *testing.T) {
	req := handle.ReqAccountRecords{
		Account: "20240511.bit",
	}
	url := TestUrl + "/account/records"
	var data handle.RespAccountRecords
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDidCellEditRecord(t *testing.T) {
	req := handle.ReqDidCellEditRecord{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgytmmrfg7aczevlxngqnr28npj2849erjyqqhe2guh",
			},
		},
		Account: "20240604.bit",
		RawParam: struct {
			Records []handle.ReqRecord `json:"records"`
		}{},
	}
	var records []handle.ReqRecord
	records = append(records, handle.ReqRecord{
		Key:   "0",
		Type:  "address",
		Label: "",
		Value: "tb1pzl9nkuavvt303hly08u3ug0v55yd3a8x86d8g5jsrllsaell8j5s8gzedg",
		TTL:   "300",
	})
	req.RawParam.Records = records
	url := TestUrl + "/account/edit/records"
	var data handle.RespDidCellEditRecord
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
	fmt.Println(data.CKBTx)

	//if err := doSig(&data.SignInfo); err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(toolib.JsonString(&data))
	//fmt.Println("===========================")

	//if err := sendTx2(data.SignInfo); err != nil {
	//	t.Fatal(err)
	//}
}

func TestDidCellRenew(t *testing.T) {
	req := handle.ReqDidCellRenew{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgytmmrfg7aczevlxngqnr28npj2849erjyqqhe2guh",
				//CoinType: common.CoinTypeCKB,
				//Key:      "ckt1qrc77cdkja6s3k0v2mlyxwv6q8jhvzr2wm8s7lrg052psv6733qp7qgzt8h5fs",
			},
		},
		Account:    "20240604.bit",
		PayTokenId: tables.TokenIdDas,
		RenewYears: 1,
	}

	url := TestUrl + "/account/order/renew"
	var data handle.RespDidCellRenew
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")

	//if len(data.SignInfo.SignList) > 0 {
	//	if err := doSig(&data.SignInfo); err != nil {
	//		t.Fatal(err)
	//	}
	//	fmt.Println(toolib.JsonString(&data))
	//	fmt.Println("===========================")
	//
	//	if err := sendTx2(data.SignInfo); err != nil {
	//		t.Fatal(err)
	//	}
	//}
}

func TestDidCellEditOwner(t *testing.T) {
	req := handle.ReqDidCellEditOwner{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
				//Key: "ckt1qrc77cdkja6s3k0v2mlyxwv6q8jhvzr2wm8s7lrg052psv6733qp7qgp95zz80",
				//Key: "ckt1qrc77cdkja6s3k0v2mlyxwv6q8jhvzr2wm8s7lrg052psv6733qp7qgzt8h5fs",
			},
		},
		Account: "20230616.bit",
		RawParam: struct {
			ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
			ReceiverAddress  string          `json:"receiver_address"`
		}{
			ReceiverCoinType: common.CoinTypeCKB,
			ReceiverAddress:  "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgytmmrfg7aczevlxngqnr28npj2849erjyqqhe2guh",
			//ReceiverAddress: "ckt1qrc77cdkja6s3k0v2mlyxwv6q8jhvzr2wm8s7lrg052psv6733qp7qgp95zz80",
			//ReceiverAddress:  "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgpzk3ntzys3nuwmvnar2lrs54l9pat6wy3qqcmu76w",
		},
		PayTokenId: tables.TokenIdDas,
	}
	url := TestUrl + "/account/edit/owner"
	var data handle.RespDidCellEditOwner
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")

	if err := doSig(&data.SignInfo); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
	if err := sendTx2(data.SignInfo); err != nil {
		t.Fatal(err)
	}
}

func TestDidCellRecycle(t *testing.T) {
	req := handle.ReqDidCellRecycle{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrc77cdkja6s3k0v2mlyxwv6q8jhvzr2wm8s7lrg052psv6733qp7qgzt8h5fs",
			},
		},
		Account: "20240512.bit",
	}
	url := TestUrl + "/did/cell/recycle"
	var data handle.RespDidCellRenew
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
	if err := doSig(&data.SignInfo); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")

	if err := sendTx2(data.SignInfo); err != nil {
		t.Fatal(err)
	}
}

func TestDidCellUpgradableList(t *testing.T) {
	req := handle.ReqDidCellUpgradableList{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			},
		},
		Pagination: handle.Pagination{
			Page: 1,
			Size: 10,
		},
		Keyword: "",
	}
	url := TestUrl + "/did/cell/upgradable/list"
	var data handle.RespDidCellUpgradableList
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func TestDidCellUpgradePrice(t *testing.T) {
	req := handle.ReqDidCellUpgradePrice{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrc77cdkja6s3k0v2mlyxwv6q8jhvzr2wm8s7lrg052psv6733qp7qgzt8h5fs",
			},
		},
		Account: "20240512.bit",
	}
	url := TestUrl + "/did/cell/upgrade/price"
	var data handle.RespDidCellUpgradePrice
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}

func doSig(sigInfo *handle.SignInfo) error {
	private := ""
	chainId := 17000
	var err error
	for i, v := range sigInfo.SignList {
		if v.SignMsg == "" {
			continue
		}
		var signData []byte
		sigMsg := []byte(v.SignMsg)
		switch v.SignType {
		case common.DasAlgorithmIdCkb, common.DasAlgorithmIdAnyLock:
			continue
		case common.DasAlgorithmIdTron:
			signData, err = sign.TronSignature(true, sigMsg, private)
			if err != nil {
				return fmt.Errorf("sign.TronSignature err: %s", err.Error())
			}
		case common.DasAlgorithmIdEth:
			signData, err = sign.PersonalSignature(sigMsg, private)
			if err != nil {
				return fmt.Errorf("sign.PersonalSignature err: %s", err.Error())
			}
		case common.DasAlgorithmIdEth712:
			var obj3 apitypes.TypedData
			mmJson := sigInfo.MMJson.String()
			oldChainId := fmt.Sprintf("chainId\":%d", chainId)
			newChainId := fmt.Sprintf("chainId\":\"%d\"", chainId)
			mmJson = strings.ReplaceAll(mmJson, oldChainId, newChainId)
			oldDigest := "\"digest\":\"\""
			newDigest := fmt.Sprintf("\"digest\":\"%s\"", v.SignMsg)
			mmJson = strings.ReplaceAll(mmJson, oldDigest, newDigest)
			_ = json.Unmarshal([]byte(mmJson), &obj3)
			var mmHash, signature []byte
			mmHash, signature, err := sign.EIP712Signature(obj3, private)
			if err != nil {
				return fmt.Errorf("sign.EIP712Signature err: %s", err.Error())
			}
			signData = append(signature, mmHash...)
			hexChainId := fmt.Sprintf("%x", chainId)
			chainIdData := common.Hex2Bytes(fmt.Sprintf("%016s", hexChainId))
			signData = append(signData, chainIdData...)
		case common.DasAlgorithmIdDogeChain:
			signData, err = sign.DogeSignature(sigMsg, private, true)
			if err != nil {
				return fmt.Errorf("sign.DogeSignature err: %s", err.Error())
			}
		default:
			return fmt.Errorf("unsupport SignType: %d", v.SignType)
		}
		sigInfo.SignList[i].SignMsg = common.Bytes2Hex(signData)
	}
	return nil
}

func TestBalancePay2(t *testing.T) {
	req := handle.ReqBalancePay{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			},
		},
		OrderId:    "e444cad37351630038a5c90af9ca2634",
		EvmChainId: 17000,
	}
	url := TestUrl + "/balance/pay"
	var data handle.RespBalancePay
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	if err := doSig(&data.SignInfo); err != nil {
		t.Fatal(err)
	}
	fmt.Println("===========================")
	fmt.Println(toolib.JsonString(&data))
	if err := sendTx2(data.SignInfo); err != nil {
		t.Fatal(err)
	}
}

func Test712(t *testing.T) {
	chainId := 17000
	digest := "0x18665bb86d3789c1d27b229f78144d6090686e9bfa35ddf5c91d1e306af5b142"
	private := ""
	mmJson := ``
	var obj3 apitypes.TypedData
	oldChainId := fmt.Sprintf("chainId\": %d", chainId)
	newChainId := fmt.Sprintf("chainId\":\"%d\"", chainId)
	mmJson = strings.ReplaceAll(mmJson, oldChainId, newChainId)
	oldDigest := "\"digest\": \"\""
	newDigest := fmt.Sprintf("\"digest\":\"%s\"", digest)
	mmJson = strings.ReplaceAll(mmJson, oldDigest, newDigest)
	fmt.Println(mmJson)
	err1 := json.Unmarshal([]byte(mmJson), &obj3)
	if err1 != nil {
		t.Fatal(err1)
	}
	var mmHash, signature []byte
	mmHash, signature, err := sign.EIP712Signature(obj3, private)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("mmHash:", common.Bytes2Hex(mmHash))
	fmt.Println("signature:", common.Bytes2Hex(signature))
	signData := append(signature, mmHash...)
	hexChainId := fmt.Sprintf("%x", chainId)
	chainIdData := common.Hex2Bytes(fmt.Sprintf("%016s", hexChainId))
	signData = append(signData, chainIdData...)
	fmt.Println("signData:", common.Bytes2Hex(signData))
}

func TestAlwaysSuccessAddr(t *testing.T) {
	addr, err := address.ConvertScriptToAddress(address.Testnet, &types.Script{
		CodeHash: types.HexToHash("0xf1ef61b6977508d9ec56fe43399a01e576086a76cf0f7c687d1418335e8c401f"),
		HashType: "type",
		Args:     common.Hex2Bytes("0x2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(addr)
}
