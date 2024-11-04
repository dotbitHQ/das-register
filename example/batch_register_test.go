package example

import (
	"context"
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"testing"
	"time"
)

func doRegister(dc *core.DasCore, account, address string) error {
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			Account:        account,
			AccountCharStr: nil,
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears:  1,
			InviterAccount: "tzh20221009-01.bit",
			ChannelAccount: "tzh20221009-01.bit",
		},
		PayChainType: 0,
		PayAddress:   "",
		PayTokenId:   tables.TokenIdDas,
		PayType:      "",
	}

	accountChars, err := dc.GetAccountCharSetList(account)
	if err != nil {
		return fmt.Errorf("doRegister err: %s", err.Error())
	}
	req.AccountCharStr = accountChars

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err = doReq(url, req, &data); err != nil {
		return fmt.Errorf("doReq err: %s", err.Error())
	}
	fmt.Println(data)
	return nil
}

func TestBatchRegister(t *testing.T) {
	dc, _ := getNewDasCoreTestnet2()
	address1 := "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	address2 := "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
	address3 := "0xD43B906Be6FbfFFFF60977A0d75EC93696e01dC7"

	var acc1 = make(chan string, 2)
	var acc2 = make(chan string, 2)
	var acc3 = make(chan string, 2)
	ctxServer, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case account1 := <-acc1:
				if err := doRegister(dc, account1, address1); err != nil {
					fmt.Println(err.Error())
				}
			case <-ctxServer.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case account2 := <-acc2:
				if err := doRegister(dc, account2, address2); err != nil {
					fmt.Println(err.Error())
				}
			case <-ctxServer.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case account3 := <-acc3:
				if err := doRegister(dc, account3, address3); err != nil {
					fmt.Println(err.Error())
				}
			case <-ctxServer.Done():
				return
			}
		}
	}()

	fmt.Println("begin", time.Now().String())
	for i := 0; i < 2000; i++ {
		account := fmt.Sprintf("test20221018-%d.bit", i)
		acc1 <- account
		acc2 <- account
		acc3 <- account
	}
	fmt.Println("ok", time.Now().String())
	time.Sleep(time.Second * 5)
	cancel()
}
