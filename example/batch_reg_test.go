package example

import (
	"context"
	"das_register_server/http_server/handle"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/chain/chain_evm"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"strings"
	"testing"
	"time"
)

type RegUser struct {
	DC         *core.DasCore
	KeyInfo    core.KeyInfo      `json:"key_info"`
	PrivateKey string            `json:"private_key"`
	PayTokenId tables.PayTokenId `json:"pay_token_id"`
	EthNonce   uint64            `json:"eth_nonce"`
}

func (r *RegUser) doReg(acc string) error {
	//fmt.Println(r.KeyInfo.CoinType, r.KeyInfo.Key)
	req := handle.ReqOrderRegister{
		ReqAccountSearch: handle.ReqAccountSearch{
			Account:        acc,
			AccountCharStr: nil,
		},
		ReqOrderRegisterBase: handle.ReqOrderRegisterBase{
			RegisterYears: 1,
		},
		ChainTypeAddress: core.ChainTypeAddress{
			Type: "blockchain",
			KeyInfo: core.KeyInfo{
				CoinType: r.KeyInfo.CoinType,
				Key:      r.KeyInfo.Key,
			},
		},
		PayTokenId: r.PayTokenId,
	}

	accountChars, err := r.DC.GetAccountCharSetList(acc)
	if err != nil {
		return fmt.Errorf("GetAccountCharSetList err: %s", err.Error())
	}
	req.AccountCharStr = accountChars
	//fmt.Println(toolib.JsonString(&req))

	url := TestUrl + "/account/order/register"
	var data handle.RespOrderRegister
	if err = doReq(url, req, &data); err != nil {
		return fmt.Errorf("doReq err: %s", err.Error())
	}
	//fmt.Println(data)
	if err := r.doPay(data); err != nil {
		return fmt.Errorf("doPay err: %s", err.Error())
	}

	return nil
}

func (r *RegUser) doPay(data handle.RespOrderRegister) error {
	switch r.PayTokenId {
	case tables.TokenIdDas:
		url := TestUrl + "/balance/pay"
		req := handle.ReqBalancePay{
			ChainTypeAddress: core.ChainTypeAddress{
				Type: "blockchain",
				KeyInfo: core.KeyInfo{
					CoinType: r.KeyInfo.CoinType,
					Key:      r.KeyInfo.Key,
				},
			},
			OrderId:    data.OrderId,
			EvmChainId: 17000,
		}
		var res handle.RespBalancePay
		if err := doReq(url, req, &res); err != nil {
			return fmt.Errorf("doReq err: %s", err.Error())
		}
		//fmt.Println(res)
		if err := r.doSendTx(res.SignInfo); err != nil {
			return fmt.Errorf("doSendTx err: %s", err.Error())
		}

		return nil
	case tables.TokenIdEth:
		node := "https://rpc.ankr.com/eth_holesky"
		addFee := float64(2)
		chainEvm, err := chain_evm.NewChainEvm(context.Background(), node, addFee)
		if err != nil {
			return fmt.Errorf("NewChainEvm err: %s", err.Error())
		}
		from := r.KeyInfo.Key
		to := data.ReceiptAddress
		value := decimal.NewFromInt(data.Amount.IntPart())
		dataBys := []byte(data.OrderId)

		nonce, err := chainEvm.NonceAt(from)
		if err != nil {
			return fmt.Errorf("NonceAt err: %s", err.Error())
		}
		if nonce > r.EthNonce {
			r.EthNonce = nonce
		} else {
			nonce = r.EthNonce
		}
		gasPrice, gasLimit, err := chainEvm.EstimateGas(from, to, value, dataBys, addFee)
		if err != nil {
			return fmt.Errorf("EstimateGas err: %s", err.Error())
		}
		tx, err := chainEvm.NewTransaction(from, to, value, dataBys, nonce, gasPrice, gasLimit)
		if err != nil {
			return fmt.Errorf("NewTransaction err: %s", err.Error())
		}
		tx, err = chainEvm.SignWithPrivateKey(r.PrivateKey, tx)
		if err != nil {
			return fmt.Errorf("SignWithPrivateKey err: %s", err.Error())
		}
		if err = chainEvm.SendTransaction(tx); err != nil {
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		fmt.Println("eth hash:", tx.Hash().String())
		r.EthNonce++
		return nil
	}

	return fmt.Errorf("unsupport token id: %s", r.PayTokenId)
}

func (r *RegUser) doSendTx(data handle.SignInfo) error {
	//  sign
	for i, v := range data.SignList {
		switch v.SignType {
		case common.DasAlgorithmIdEth712:
			chainId := 17000
			var obj3 apitypes.TypedData
			mmJson := data.MMJson.String()
			oldChainId := fmt.Sprintf("chainId\":%d", chainId)
			newChainId := fmt.Sprintf("chainId\":\"%d\"", chainId)
			mmJson = strings.ReplaceAll(mmJson, oldChainId, newChainId)
			oldDigest := "\"digest\":\"\""
			newDigest := fmt.Sprintf("\"digest\":\"%s\"", v.SignMsg)
			mmJson = strings.ReplaceAll(mmJson, oldDigest, newDigest)
			_ = json.Unmarshal([]byte(mmJson), &obj3)
			var mmHash, signature []byte
			mmHash, signature, err := sign.EIP712Signature(obj3, r.PrivateKey)
			if err != nil {
				return fmt.Errorf("sign.EIP712Signature err: %s", err.Error())
			}
			signData := append(signature, mmHash...)
			hexChainId := fmt.Sprintf("%x", chainId)
			chainIdData := common.Hex2Bytes(fmt.Sprintf("%016s", hexChainId))
			signData = append(signData, chainIdData...)
			data.SignList[i].SignMsg = common.Bytes2Hex(signData)
		}
	}
	//fmt.Println(toolib.JsonString(&data))

	//
	url := TestUrl + "/transaction/send"
	req := handle.ReqTransactionSend{SignInfo: data}
	var res handle.RespTransactionSend
	if err := doReq(url, req, &res); err != nil {
		return fmt.Errorf("doReq err: %s", err.Error())
	}
	fmt.Println("hash:", res.Hash)
	return nil
}

func TestBatchReg(t *testing.T) {
	dc, _ := getNewDasCoreTestnet2()
	u1 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
		},
		PrivateKey: "",
		//PayTokenId: tables.TokenIdDas,
		PayTokenId: tables.TokenIdEth,
		DC:         dc,
	}

	// test1
	//acc := "batchtest008.bit"
	//if err := u1.doReg(acc); err != nil {
	//	t.Fatal(err)
	//}

	// test2
	for i := 0; i < 3; i++ {
		acc := fmt.Sprintf("batchtest06%02d.bit", i)
		fmt.Println(acc)
		if err := u1.doReg(acc); err != nil {
			t.Fatal(err)
		}
	}

	// test3
	//var ch = make(chan string, 50)
	//group := &errgroup.Group{}
	//group.Go(func() error {
	//	for acc := range ch {
	//		if err := u1.doReg(acc); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	return nil
	//})
	//
	//for i := 0; i < 5; i++ {
	//	acc := fmt.Sprintf("batchtest02%02d.bit", i)
	//	fmt.Println(acc)
	//	ch <- acc
	//}
	//close(ch)
	//
	//if err := group.Wait(); err != nil {
	//	t.Fatal(err)
	//}

	// test4
	//u2 := RegUser{
	//	KeyInfo: core.KeyInfo{
	//		CoinType: common.CoinTypeEth,
	//		Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
	//	},
	//	PrivateKey: "",
	//	PayTokenId: tables.TokenIdDas,
	//	DC:         dc,
	//}
	//u3 := RegUser{
	//	KeyInfo: core.KeyInfo{
	//		CoinType: common.CoinTypeEth,
	//		Key:      "0x7038595295Ae464bEFEA0030a283D0f481a3be73",
	//	},
	//	PrivateKey: "",
	//	PayTokenId: tables.TokenIdDas,
	//	DC:         dc,
	//}
	//var ch1 = make(chan string, 50)
	////var ch2 = make(chan string, 50)
	////var ch3 = make(chan string, 50)
	//group := &errgroup.Group{}
	//group.Go(func() error {
	//	for acc := range ch1 {
	//		if err := u1.doReg(acc); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	return nil
	//})
	//group.Go(func() error {
	//	for acc := range ch1 {
	//		if err := u2.doReg(acc); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	return nil
	//})
	//group.Go(func() error {
	//	for acc := range ch1 {
	//		if err := u3.doReg(acc); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	//	return nil
	//})
	//
	//for i := 0; i < 50; i++ {
	//	acc := fmt.Sprintf("batchtest04%02d.bit", i)
	//	fmt.Println(acc)
	//	ch1 <- acc
	//	//ch2 <- acc
	//	//ch3 <- acc
	//}
	//close(ch1)
	////close(ch2)
	////close(ch3)
	//
	//if err := group.Wait(); err != nil {
	//	t.Fatal(err)
	//}

}

func TestBatchReg2(t *testing.T) {
	dc, _ := getNewDasCoreTestnet2()
	u1 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x7038595295Ae464bEFEA0030a283D0f481a3be73",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u2 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u3 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}

	group := &errgroup.Group{}
	tic := time.NewTicker(time.Second * 10)
	i := 75
	var ch1 = make(chan string, 10)
	var ch2 = make(chan string, 10)
	var ch3 = make(chan string, 10)
	group.Go(func() error {
		for {
			select {
			case <-tic.C:
				acc1 := fmt.Sprintf("batchtest05%03d.bit", i)
				i++
				acc2 := fmt.Sprintf("batchtest05%03d.bit", i)
				i++
				acc3 := fmt.Sprintf("batchtest05%03d.bit", i)
				i++
				ch1 <- acc1
				ch2 <- acc2
				ch3 <- acc3
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch1 {
			fmt.Println("ch1:", acc)
			if err := u1.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch2 {
			fmt.Println("ch2:", acc)
			if err := u2.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch3 {
			fmt.Println("ch3:", acc)
			if err := u3.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestCkbChange(t *testing.T) {
	dc, _ := getNewDasCoreTestnet2()
	amount := uint64(20000 * common.OneCkb)
	//fromAddr := "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qg9e86nk8v9x44kq3flsemppzyd3xstveadqhyl2wcas56kkcz987r8vyyg3ky6pdn845uxtt3h"
	//privateFrom := ""
	//fromAddr := "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qg9zk3ntzys3nuwmvnar2lrs54l9pat6wy3q526xdvgjzx03mdj05dtuwzjhu5840fcjy2c9u8d"
	//privateFrom := ""
	fromAddr := "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qg9wqu9j5544eryhml2qqc29q7s7jq680nnq4crsk2jjkhyvjl0agqrpg5r6r6grga7wvw84mev"
	privateFrom := ""
	fromPA, err := address.Parse(fromAddr)
	if err != nil {
		t.Fatal(err)
	}

	txBuilderBase := getTxBuilderBase(dc, "", privateFrom)
	liveCells, total, err := dc.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        fromPA.Script,
		CapacityNeed:      amount,
		CapacityForChange: common.DasLockWithBalanceTypeMinCkbCapacity + common.OneCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		t.Fatal(err)
	}
	var txParams txbuilder.BuildTransactionParams
	var changeLock, changeType *types.Script
	for _, v := range liveCells {
		changeLock = v.Output.Lock
		changeType = v.Output.Type
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			Since:          0,
			PreviousOutput: v.OutPoint,
		})
	}
	// outputs
	splitCkb := 1000 * common.OneCkb
	changeList, err := core.SplitOutputCell2(total, splitCkb, 20, changeLock, changeType, indexer.SearchOrderAsc)
	if err != nil {
		t.Fatal("SplitOutputCell2:", err.Error())
	}
	for i := 0; i < len(changeList); i++ {
		txParams.Outputs = append(txParams.Outputs, changeList[i])
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionWithdrawFromWallet, nil)
	if err != nil {
		t.Fatal(err)
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	//
	txBuilder := txbuilder.NewDasTxBuilderFromBase(txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(&txParams); err != nil {
		t.Fatal(err)
	}

	sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
	txFee := sizeInBlock + 1000
	changeCapacity := txBuilder.Transaction.Outputs[0].Capacity - txFee
	txBuilder.Transaction.Outputs[0].Capacity = changeCapacity

	// sign
	signList, err := txBuilder.GenerateDigestListFromTx([]int{})
	if err != nil {
		t.Fatal(err)
	}
	mmJsonObj, err := txBuilder.BuildMMJsonObj(17000)
	if err != nil {
		t.Fatal(err)
	}

	for i, v := range signList {
		switch v.SignType {
		case common.DasAlgorithmIdEth712:
			chainId := 17000
			var obj3 apitypes.TypedData
			mmJson := mmJsonObj.String()
			oldChainId := fmt.Sprintf("chainId\":%d", chainId)
			newChainId := fmt.Sprintf("chainId\":\"%d\"", chainId)
			mmJson = strings.ReplaceAll(mmJson, oldChainId, newChainId)
			oldDigest := "\"digest\":\"\""
			newDigest := fmt.Sprintf("\"digest\":\"%s\"", v.SignMsg)
			mmJson = strings.ReplaceAll(mmJson, oldDigest, newDigest)
			_ = json.Unmarshal([]byte(mmJson), &obj3)
			var mmHash, signature []byte
			mmHash, signature, err := sign.EIP712Signature(obj3, privateFrom)
			if err != nil {
				t.Fatal(err)
			}
			signData := append(signature, mmHash...)
			hexChainId := fmt.Sprintf("%x", chainId)
			chainIdData := common.Hex2Bytes(fmt.Sprintf("%016s", hexChainId))
			signData = append(signData, chainIdData...)
			signList[i].SignMsg = common.Bytes2Hex(signData)
		}
	}
	if err := txBuilder.AddSignatureForTx(signList); err != nil {
		t.Fatal(err)
	}

	if hash, err := txBuilder.SendTransactionWithCheck(false); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("hash:", hash)
	}
}

func getTxBuilderBase(dasCore *core.DasCore, args, privateKey string) *txbuilder.DasTxBuilderBase {
	handleSign := sign.LocalSign(privateKey)
	txBuilderBase := txbuilder.NewDasTxBuilderBase(context.Background(), dasCore, handleSign, args)
	return txBuilderBase
}
