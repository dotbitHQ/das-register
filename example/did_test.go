package example

import (
	"das_register_server/http_server/handle"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"testing"
)

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
			ReceiverCoinType: common.CoinTypeEth,
			ReceiverAddress:  "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
		},
		PayTokenId: "",
	}
	url := TestUrl + "/did/cell/edit/owner"
	var data handle.RespDidCellEditOwner
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
