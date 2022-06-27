package example

import (
	"crypto/ed25519"
	"das_register_server/http_server/handle"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"testing"
)

func TestBalanceWithdraw(t *testing.T) {
	url := TestUrl + "/balance/withdraw"

	var req handle.ReqBalanceWithdraw
	req.ChainType = common.ChainTypeMixin
	req.Address = "0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"
	req.ReceiverAddress = "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qgxuyyse6pywn97pvvk68nzas6fasz6vyrkc6x3gy5jv5mseflq28zqdcgfpn5zgaxtuzced50x9mp5nmq95cg8d35dzsffyefhpjn7q5wyxgqtre"
	req.Amount = decimal.NewFromInt(50000000000)
	req.EvmChainId = 5

	var data handle.RespBalanceWithdraw

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

func TestSign(t *testing.T) {
	private := common.Hex2Bytes("0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4")
	msg := common.Hex2Bytes("0x6d31bda56835b9c2d4876a53d611dfb58238aacf26ca00d3c8f5a2165c3f70cf")
	sig := common.Hex2Bytes("0x4d0fff8474b060546d7cd5310ba317412e100c53bad5a15665052e344b615f979bf951666276d8e1548d2a39b899f518d2d6718ad4bfe4ee2b6bc988b049bd0d")
	fmt.Println(ed25519.Verify(private, msg, sig))
}
