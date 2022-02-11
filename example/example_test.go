package example

import (
	"context"
	"das_register_server/http_server/api_code"
	"das_register_server/http_server/handle"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/sign"
	"github.com/DeAccountSystems/das-lib/txbuilder"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"testing"
	"time"
)

func TestTronSign(t *testing.T) {
	signType := true
	data := common.Hex2Bytes("0xf841d39c7599b720e32453729eb24072956e5475425a4c188287136bba9fa4a4")
	privateKey := ""
	address := "TQoLh9evwUmZKxpD1uhFttsZk3EBs8BksV"
	signature, err := sign.TronSignature(signType, data, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(common.Bytes2Hex(signature))
	//curl -X POST http://127.0.0.1:8120/v1/transaction/send -d'{"sign_key":"3a8aac6d0ab9791f2b0de0207d29860e","sign_list":[{"sign_type":4,"sign_msg":"0xe087beaf79cdc97a8852ef302e918238e729a9a30413abeedae9ef16c482c4b21d14d2ec07be1dfd5f35c68a4dc4a88884104a919802525d867315db51b66b871c"}]}'

	fmt.Println(sign.TronVerifySignature(signType, signature, data, address))
}

func TestPersonalSignature(t *testing.T) {
	data := common.Hex2Bytes("0x07f495e2f611979835f2735eb78bcee409726c12f51f01aa6b5e903fdedea538")
	privateKey := ""
	address := "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	signature, err := sign.PersonalSignature(data, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(common.Bytes2Hex(signature))

	fmt.Println(sign.VerifyPersonalSignature(signature, data, address))
}

func TestGetLiveCell(t *testing.T) {
	client, err := getClientTestnet2()
	if err != nil {
		t.Fatal(err)
	}
	outPoint := common.String2OutPointStruct("0x8adaddae4f1fd21d47d0924f01422be9fdbb4171f48214653adcc1b83cb7b84a-0")
	cell, err := client.GetLiveCell(context.Background(), outPoint, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cell.Cell.Output.Lock.Hash())
}

func getClientTestnet2() (rpc.Client, error) {
	ckbUrl := "http://100.77.204.22:8224"
	indexerUrl := "http://100.77.204.22:8226"
	return rpc.DialWithIndexer(ckbUrl, indexerUrl)
}

func TestEIP712Signature(t *testing.T) {
	mmjson := `{
        "types":{
            "EIP712Domain":[
                {
                    "name":"chainId",
                    "type":"uint256"
                },
                {
                    "name":"name",
                    "type":"string"
                },
                {
                    "name":"verifyingContract",
                    "type":"address"
                },
                {
                    "name":"version",
                    "type":"string"
                }
            ],
            "Action":[
                {
                    "name":"action",
                    "type":"string"
                },
                {
                    "name":"params",
                    "type":"string"
                }
            ],
            "Cell":[
                {
                    "name":"capacity",
                    "type":"string"
                },
                {
                    "name":"lock",
                    "type":"string"
                },
                {
                    "name":"type",
                    "type":"string"
                },
                {
                    "name":"data",
                    "type":"string"
                },
                {
                    "name":"extraData",
                    "type":"string"
                }
            ],
            "Transaction":[
                {
                    "name":"DAS_MESSAGE",
                    "type":"string"
                },
                {
                    "name":"inputsCapacity",
                    "type":"string"
                },
                {
                    "name":"outputsCapacity",
                    "type":"string"
                },
                {
                    "name":"fee",
                    "type":"string"
                },
                {
                    "name":"action",
                    "type":"Action"
                },
                {
                    "name":"inputs",
                    "type":"Cell[]"
                },
                {
                    "name":"outputs",
                    "type":"Cell[]"
                },
                {
                    "name":"digest",
                    "type":"bytes32"
                }
            ]
        },
        "primaryType":"Transaction",
        "domain":{
            "chainId":"5",
            "name":"da.systems",
            "verifyingContract":"0x0000000000000000000000000000000020210722",
            "version":"1"
        },
        "message":{
            "DAS_MESSAGE":"RETRACT REVERSE RECORDS ON 0x15a33588908cf8edb27d1abe3852bf287abd3891",
            "inputsCapacity":"200.9997 CKB",
            "outputsCapacity":"200.9996 CKB",
            "fee":"0.0001 CKB",
            "digest":"0x0d38bbc46caad651081216a9ba338b301ef64d0f20d7062cd986c7c1776eda3f",
            "action":{
                "action":"retract_reverse_record",
                "params":"0x00"
            },
            "inputs":[
                {
                    "capacity":"200.9997 CKB",
                    "lock":"das-lock,0x01,0x0515a33588908cf8edb27d1abe3852bf287abd38...",
                    "type":"reverse-record-cell-type,0x01,0x",
                    "data":"0x74616e677465737430342e626974",
                    "extraData":""
                }
            ],
            "outputs":[

            ]
        }
    }`
	privateKey := ""
	var obj3 core.TypedData
	_ = json.Unmarshal([]byte(mmjson), &obj3)
	mmHash, signature, err := sign.EIP712Signature(obj3, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(common.Bytes2Hex(signature))
	fmt.Println(common.Bytes2Hex(mmHash))
	signMsg := append(signature, mmHash...)

	fmt.Println(common.Bytes2Hex(signMsg))

	fmt.Println(sign.VerifyEIP712Signature(obj3, signature, "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"))

	var req handle.ReqTransactionSend
	req.SignKey = "ca159eb4fb2d7890ec46cb3d2231928b"
	req.SignList = []txbuilder.SignData{txbuilder.SignData{
		SignType: common.DasAlgorithmIdEth712,
		SignMsg:  common.Bytes2Hex(signMsg) + "0000000000000005",
	}}

	url := TestUrl + "/transaction/send"

	var data handle.RespTransactionSend
	var resp api_code.ApiResp
	resp.Data = &data
	_, _, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&resp)
	if errs != nil {
		t.Fatal(errs)
	}
	fmt.Println(toolib.JsonString(data))
}

func Test(t *testing.T) {
	//str := "0000000000000061"
	d := math.NewDecimal256(97)
	fmt.Println(d.String())
	hex := fmt.Sprintf("%x", 97)

	data := fmt.Sprintf("%016s", hex)
	fmt.Println(data)
	//fmt.Println(common.Hex2Bytes(data))
}

func TestSelect(t *testing.T) {
	ticker := time.NewTicker(time.Second * 2)
	ticker2 := time.NewTicker(time.Second * 6)
	for {
		select {
		case <-ticker.C:
			fmt.Println("1111")
		case <-ticker2.C:
			fmt.Println("2222")
			time.Sleep(time.Second * 10)
		}
	}
}
