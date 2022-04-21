package example

import (
	"das_register_server/http_server/handle"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/secp256k1"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestAddressDeposit(t *testing.T) {
	var req handle.ReqAddressDeposit
	req.AlgorithmId = 6
	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"

	url := TestUrl + "/address/deposit"
	var data handle.RespAddressDeposit
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(data))
}

func TestBalanceDeposit(t *testing.T) {
	var req handle.ReqBalanceDeposit
	req.FromCkbAddress = "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q"
	req.ToCkbAddress = "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qgxuyyse6pywn97pvvk68nzas6fasz6vyrkc6x3gy5jv5mseflq28zqdcgfpn5zgaxtuzced50x9mp5nmq95cg8d35dzsffyefhpjn7q5wyxgqtre"
	req.Amount = 116 * common.OneCkb

	url := TestUrl + "/balance/deposit"

	var data handle.RespBalanceDeposit
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(data))

	private := ""
	for i, v := range data.SignList {
		key, err := secp256k1.HexToKey(private)
		if err != nil {
			t.Fatal(err)
		}
		signed, err := key.Sign(common.Hex2Bytes(v.SignMsg))
		if err != nil {
			t.Fatal(err)
		}
		//signData, err := sign.EthSignature(common.Hex2Bytes(v.SignMsg), private)
		//if err != nil {
		//	t.Fatal(err)
		//}
		data.SignList[i].SignMsg = common.Bytes2Hex(signed)
	}
	fmt.Println(toolib.JsonString(data))
}
