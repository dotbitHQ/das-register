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

func TestDidCellList(t *testing.T) {
	req := handle.ReqDidCellList{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				Key:      "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgpzk3ntzys3nuwmvnar2lrs54l9pat6wy3qqcmu76w",
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

func TestDidCellEditRecord(t *testing.T) {
	req := handle.ReqDidCellEditRecord{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
			},
		},
		Account: "20240507.bit",
		RawParam: struct {
			Records []handle.ReqRecord `json:"records"`
		}{},
	}
	var records []handle.ReqRecord
	records = append(records, handle.ReqRecord{
		Key:   "60",
		Type:  "address",
		Label: "",
		Value: "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
		TTL:   "300",
	})
	req.RawParam.Records = records
	url := TestUrl + "/did/cell/edit/record"
	var data handle.RespDidCellEditRecord
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

func TestDidCellRenew(t *testing.T) {
	req := handle.ReqDidCellRenew{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			},
		},
		Account:    "20240507.bit",
		PayTokenId: tables.TokenIdDas,
		RenewYears: 1,
	}

	url := TestUrl + "/did/cell/renew"
	var data handle.RespDidCellRenew
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")

	if len(data.SignInfo.SignList) > 0 {
		if err := doSig(&data.SignInfo); err != nil {
			t.Fatal(err)
		}
		fmt.Println(toolib.JsonString(&data))
		fmt.Println("===========================")

		if err := sendTx2(data.SignInfo); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDidCellEditOwner(t *testing.T) {
	req := handle.ReqDidCellEditOwner{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeEth,
				Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
			},
		},
		Account: "20240507.bit",
		RawParam: struct {
			ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
			ReceiverAddress  string          `json:"receiver_address"`
		}{
			ReceiverCoinType: common.CoinTypeCKB,
			ReceiverAddress:  "ckt1qrejnmlar3r452tcg57gvq8patctcgy8acync0hxfnyka35ywafvkqgpzk3ntzys3nuwmvnar2lrs54l9pat6wy3qqcmu76w",
		},
		PayTokenId: tables.TokenIdDas,
	}
	url := TestUrl + "/did/cell/edit/owner"
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

func doSig(sigInfo *handle.SignInfo) error {
	private := ""
	chainId := 17000
	var err error
	for i, v := range sigInfo.SignList {
		var signData []byte
		sigMsg := []byte(v.SignMsg)
		switch v.SignType {
		case common.DasAlgorithmIdCkb:
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
		OrderId:    "7de94ff90e3ce2bdb75fc3172c751982",
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
	mmJson := `{
        "types": {
            "EIP712Domain": [
                {
                    "name": "name",
                    "type": "string"
                },
                {
                    "name": "version",
                    "type": "string"
                },
                {
                    "name": "chainId",
                    "type": "uint256"
                },
                {
                    "name": "verifyingContract",
                    "type": "address"
                }
            ],
            "Action": [
                {
                    "name": "action",
                    "type": "string"
                },
                {
                    "name": "params",
                    "type": "string"
                }
            ],
            "Cell": [
                {
                    "name": "capacity",
                    "type": "string"
                },
                {
                    "name": "lock",
                    "type": "string"
                },
                {
                    "name": "type",
                    "type": "string"
                },
                {
                    "name": "data",
                    "type": "string"
                },
                {
                    "name": "extraData",
                    "type": "string"
                }
            ],
            "Transaction": [
                {
                    "name": "DAS_MESSAGE",
                    "type": "string"
                },
                {
                    "name": "inputsCapacity",
                    "type": "string"
                },
                {
                    "name": "outputsCapacity",
                    "type": "string"
                },
                {
                    "name": "fee",
                    "type": "string"
                },
                {
                    "name": "action",
                    "type": "Action"
                },
                {
                    "name": "inputs",
                    "type": "Cell[]"
                },
                {
                    "name": "outputs",
                    "type": "Cell[]"
                },
                {
                    "name": "digest",
                    "type": "bytes32"
                }
            ]
        },
        "primaryType": "Transaction",
        "domain": {
            "name": "d.id",
            "version": "1",
            "chainId": 17000,
            "verifyingContract": "0x0000000000000000000000000000000020210722"
        },
        "message": {
            "DAS_MESSAGE": "TRANSFER THE ACCOUNT 20240507.bit TO 0x15a33588908cf8edb27d1abe3852bf287abd3891",
            "inputsCapacity": "1758.64161614 CKB",
            "outputsCapacity": "1758.64156958 CKB",
            "fee": "0.00004656 CKB",
            "digest": "",
            "action": {
                "action": "transfer_account",
                "params": "0x00"
            },
            "inputs": [
                {
                    "capacity": "218.99986273 CKB",
                    "lock": "das-lock,0x01,0x0515a33588908cf8edb27d1abe3852bf287abd38...",
                    "type": "account-cell-type,0x01,0x",
                    "data": "{ account: 20240507.bit, expired_at: 1778140699 }",
                    "extraData": "{ status: 0, records_hash: 0x55478d76900611eb079b22088081124ed6c8bae21a05dd1a0d197efcc7c114ce }"
                }
            ],
            "outputs": [
                {
                    "capacity": "218.99981617 CKB",
                    "lock": "das-lock,0x01,0x0515a33588908cf8edb27d1abe3852bf287abd38...",
                    "type": "account-cell-type,0x01,0x",
                    "data": "{ account: 20240507.bit, expired_at: 1778140699 }",
                    "extraData": "{ status: 153, records_hash: 0x55478d76900611eb079b22088081124ed6c8bae21a05dd1a0d197efcc7c114ce }"
                },
                {
                    "capacity": "160 CKB",
                    "lock": "",
                    "type": "did-cell-type,0x01,0x",
                    "data": "0x000000003c00000010000000240000002c000000...",
                    "extraData": ""
                }
            ]
        }
    }`
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
