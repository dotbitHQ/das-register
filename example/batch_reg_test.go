package example

import (
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/scorpiotzh/toolib"
	"testing"
)

type RegUser struct {
	DC         *core.DasCore
	KeyInfo    core.KeyInfo      `json:"key_info"`
	PrivateKey string            `json:"private_key"`
	PayTokenId tables.PayTokenId `json:"pay_token_id"`
}

func (r *RegUser) doReg(acc string) error {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			ChainTypeAddress: core.ChainTypeAddress{
				Type: "blockchain",
				KeyInfo: core.KeyInfo{
					CoinType: r.KeyInfo.CoinType,
					Key:      r.KeyInfo.Key,
				},
			},
			Account:        acc,
			AccountCharStr: nil,
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears: 1,
		},
		PayTokenId: r.PayTokenId,
	}

	accountChars, err := r.DC.GetAccountCharSetList(acc)
	if err != nil {
		return fmt.Errorf("GetAccountCharSetList err: %s", err.Error())
	}
	req.AccountCharStr = accountChars
	log.Info(toolib.JsonString(&req))

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err = doReq(url, req, &data); err != nil {
		return fmt.Errorf("doReq err: %s", err.Error())
	}
	fmt.Println(data)
	return nil
}

func (r *RegUser) doPay() error {
	switch r.PayTokenId {
	case tables.TokenIdDas:
		// todo
		return nil
	case tables.TokenIdEth:
		// todo
		return nil
	}
	return fmt.Errorf("unsupport token id: %s", r.PayTokenId)
}

func TestBatchReg(t *testing.T) {
	dc, _ := getNewDasCoreTestnet2()
	u1 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}

	acc := "batch001.bit"
	if err := u1.doReg(acc); err != nil {
		t.Fatal(err)
	}
}
