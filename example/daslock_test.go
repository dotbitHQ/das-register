package example

import (
	"das_register_server/http_server/handle"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestDidCellDasLockEditOwner(t *testing.T) {
	req := handle.ReqDidCellDasLockEditOwner{
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: common.CoinTypeCKB,
				//Key:      "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qgrzk3ntzys3nuwmvnar2lrs54l9pat6wy3qv26xdvgjzx03mdj05dtuwzjhu5840fcjyg4hdja",
				Key: "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qg9zk3ntzys3nuwmvnar2lrs54l9pat6wy3q526xdvgjzx03mdj05dtuwzjhu5840fcjy2c9u8d",
			},
		},
		//Account: "2023100802.bit",
		Account: "20240620.bit",
		RawParam: struct {
			ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
			ReceiverAddress  string          `json:"receiver_address"`
		}{
			ReceiverCoinType: common.CoinTypeCKB,
			ReceiverAddress:  "ckt1qrfrwcdnvssswdwpn3s9v8fp87emat306ctjwsm3nmlkjg8qyza2cqgqqxjg99grmvgl0sljs3essy47l8tthsxp9sumhp20",
		},
	}

	//addrHexFrom, err := req.FormatChainTypeAddress(common.DasNetTypeTestnet2, true)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(addrHexFrom.AddressHex, addrHexFrom.DasAlgorithmId, addrHexFrom.AddressPayload)

	url := TestUrl + "/did/cell/daslock/edit/owner"
	var data handle.RespDidCellDasLockEditOwner
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	//fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
	//fmt.Println(data.CKBTx)

	if err := doSig(&data.SignInfo); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
	fmt.Println("===========================")
	if err := sendTx2(data.SignInfo); err != nil {
		t.Fatal(err)
	}
}
