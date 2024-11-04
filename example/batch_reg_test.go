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
	"io/ioutil"
	"strconv"
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
	u0 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u9 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
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
			Key:      "0x1A5CD1c976b846695633caC0307DB418E4472a5c",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u3 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x911399D06AE2aA323B203e6bAFA28397c2495173",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u4 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0xe5589D9d1c2D1D46e3f8B77f8b82E0eE16D33BCa",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u5 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0xF37302B4A3A665B99FCAD06eD5cfEbf85207Da8f",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u6 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x79c5ebd2dBE02e7A0C08F38437D39c82B1DD39E2",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u7 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0x2E9a84eB6676EF2B813b7E83A596a49B6dF689f5",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}
	u8 := RegUser{
		KeyInfo: core.KeyInfo{
			CoinType: common.CoinTypeEth,
			Key:      "0xF766FA6a9dB8F85eAae68a2B638485eDF771C65E",
		},
		PrivateKey: "",
		PayTokenId: tables.TokenIdDas,
		DC:         dc,
	}

	group := &errgroup.Group{}
	tic := time.NewTicker(time.Second * 10)
	i := 0
	var ch0 = make(chan string, 50)
	var ch1 = make(chan string, 50)
	var ch2 = make(chan string, 50)
	var ch3 = make(chan string, 50)
	var ch4 = make(chan string, 50)
	var ch5 = make(chan string, 50)
	var ch6 = make(chan string, 50)
	var ch7 = make(chan string, 50)
	var ch8 = make(chan string, 50)
	var ch9 = make(chan string, 50)

	group.Go(func() error {
		for {
			select {
			case <-tic.C:
				acc0 := fmt.Sprintf("batchtest000%03d.bit", i)
				acc1 := fmt.Sprintf("batchtest001%03d.bit", i)
				acc2 := fmt.Sprintf("batchtest002%03d.bit", i)
				acc3 := fmt.Sprintf("batchtest003%03d.bit", i)
				acc4 := fmt.Sprintf("batchtest004%03d.bit", i)
				acc5 := fmt.Sprintf("batchtest005%03d.bit", i)
				acc6 := fmt.Sprintf("batchtest006%03d.bit", i)
				acc7 := fmt.Sprintf("batchtest007%03d.bit", i)
				acc8 := fmt.Sprintf("batchtest008%03d.bit", i)
				acc9 := fmt.Sprintf("batchtest009%03d.bit", i)
				i++
				ch0 <- acc0
				ch1 <- acc1
				ch2 <- acc2
				ch3 <- acc3
				ch4 <- acc4
				ch5 <- acc5
				ch6 <- acc6
				ch7 <- acc7
				ch8 <- acc8
				ch9 <- acc9
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch0 {
			fmt.Println("ch0:", acc)
			if err := u0.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
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
	group.Go(func() error {
		for acc := range ch4 {
			fmt.Println("ch4:", acc)
			if err := u4.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch5 {
			fmt.Println("ch5:", acc)
			if err := u5.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch6 {
			fmt.Println("ch6:", acc)
			if err := u6.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch7 {
			fmt.Println("ch7:", acc)
			if err := u7.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch8 {
			fmt.Println("ch8:", acc)
			if err := u8.doReg(acc); err != nil {
				fmt.Println("doReg err: ", err.Error())
			}
		}
		return nil
	})
	group.Go(func() error {
		for acc := range ch9 {
			fmt.Println("ch9:", acc)
			if err := u9.doReg(acc); err != nil {
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

func TestBatchCsv(t *testing.T) {
	content, err := ioutil.ReadFile("")

	if err != nil {
		t.Fatal(err)
	}
	str := string(content[4:])
	str = strings.ReplaceAll(str, "\"", "")
	//fmt.Println(str)
	list := strings.Split(str, "\n")
	//fmt.Println(len(list))
	var mapBatchCsv = make(map[string][]BatchCsv)
	for _, v := range list {
		if strings.TrimSpace(v) == "" {
			break
		}

		res := strings.Split(strings.TrimSpace(v), ",")
		timestamp, _ := strconv.ParseInt(res[2], 10, 64)
		startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", res[3], time.Local)
		endTime, _ := time.ParseInLocation("2006-01-02 15:04:05", res[4], time.Local)
		tmp := BatchCsv{
			OrderId:   res[0],
			Action:    res[1],
			Timestamp: timestamp,
			StartTime: startTime.Unix(),
			EndTime:   endTime.Unix(),
		}
		if tmp.OrderId == "008eaf1deffafbd1a89fcc1c2f11ec0b" {
			fmt.Println(tmp.OrderId)
		}
		mapBatchCsv[tmp.OrderId] = append(mapBatchCsv[tmp.OrderId], tmp)
	}
	totalAcc := int64(len(mapBatchCsv))
	fmt.Println(totalAcc)
	var mapAction = make(map[string]ActionCount)
	for _, v := range mapBatchCsv {
		var preAction, appAction BatchCsv
		for _, action := range v {
			if action.OrderId == "008eaf1deffafbd1a89fcc1c2f11ec0b" {
				fmt.Println(action.OrderId)
			}
			switch action.Action {
			case "propose", "confirm_proposal":
				txTime := action.EndTime - preAction.Timestamp/1000
				if item, ok := mapAction[action.Action]; ok {
					item.AvgTime += txTime
					if item.MinTime > txTime {
						item.MinTime = txTime
						item.MinOrderId = action.OrderId
					}
					if item.MaxTime < txTime {
						item.MaxTime = txTime
						item.MaxOrderId = action.OrderId
					}
					mapAction[action.Action] = item
				} else {
					mapAction[action.Action] = ActionCount{
						MinTime:    txTime,
						MaxTime:    txTime,
						AvgTime:    txTime,
						MinOrderId: action.OrderId,
						MaxOrderId: action.OrderId,
					}
				}
				if action.Action == "confirm_proposal" {
					txTime = action.EndTime - appAction.StartTime
					if item, ok := mapAction["avg"]; ok {
						item.AvgTime += txTime
						if item.MinTime > txTime {
							item.MinTime = txTime
							item.MinOrderId = action.OrderId
						}
						if item.MaxTime < txTime {
							item.MaxTime = txTime
							item.MaxOrderId = action.OrderId
						}
						mapAction["avg"] = item
					} else {
						mapAction["avg"] = ActionCount{
							MinTime:    txTime,
							MaxTime:    txTime,
							AvgTime:    txTime,
							MinOrderId: action.OrderId,
							MaxOrderId: action.OrderId,
						}
					}
				}
			case "apply_register", "pre_register":
				if action.Action == "pre_register" {
					preAction = action
				} else {
					appAction = action
				}

				txTime := action.EndTime - action.StartTime
				if item, ok := mapAction[action.Action]; ok {
					item.AvgTime += txTime
					if item.MinTime > txTime {
						item.MinTime = txTime
						item.MinOrderId = action.OrderId
					}
					if item.MaxTime < txTime {
						item.MaxTime = txTime
						item.MaxOrderId = action.OrderId
					}
					mapAction[action.Action] = item
				} else {
					mapAction[action.Action] = ActionCount{
						MinTime:    txTime,
						MaxTime:    txTime,
						AvgTime:    txTime,
						MinOrderId: action.OrderId,
						MaxOrderId: action.OrderId,
					}
				}
			}
		}
	}
	for k, v := range mapAction {
		fmt.Println(k, v.MinOrderId, v.MaxOrderId, v.MinTime, v.MaxTime, v.AvgTime/totalAcc)
	}
}

type BatchCsv struct {
	OrderId   string `json:"order_id"`
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type ActionCount struct {
	MinOrderId string `json:"min_order_id"`
	MinTime    int64  `json:"min_time"`
	MaxOrderId string `json:"max_order_id"`
	MaxTime    int64  `json:"max_time"`
	AvgTime    int64  `json:"avg_time"`
}
