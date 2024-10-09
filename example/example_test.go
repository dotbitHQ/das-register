package example

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"sync"
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

func TestAccount(t *testing.T) {
	fmt.Println(common.GetAccountLength("ให้บริการ.bit"))
}

func getClientTestnet2() (rpc.Client, error) {
	ckbUrl := "https://testnet.ckb.dev/"
	indexerUrl := "https://testnet.ckb.dev/indexer"
	return rpc.DialWithIndexer(ckbUrl, indexerUrl)
}

func getNewDasCoreTestnet2() (*core.DasCore, error) {
	client, err := getClientTestnet2()
	if err != nil {
		return nil, err
	}

	env := core.InitEnvOpt(common.DasNetTypeTestnet2,
		common.DasContractNameConfigCellType,
		//common.DasContractNameAccountCellType,
		//common.DasContractNameDispatchCellType,
		//common.DasContractNameBalanceCellType,
		//common.DasContractNameAlwaysSuccess,
		//common.DasContractNameIncomeCellType,
		//common.DASContractNameSubAccountCellType,
		//common.DasContractNamePreAccountCellType,
	)
	var wg sync.WaitGroup
	ops := []core.DasCoreOption{
		core.WithClient(client),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(common.DasNetTypeTestnet2),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dc := core.NewDasCore(context.Background(), &wg, ops...)
	// contract
	dc.InitDasContract(env.MapContract)
	// config cell
	if err = dc.InitDasConfigCell(); err != nil {
		return nil, err
	}
	// so script
	if err = dc.InitDasSoScript(); err != nil {
		return nil, err
	}
	return dc, nil
}

func getClientMainNet() (rpc.Client, error) {
	ckbUrl := "http://127.0.0.1:8114"
	indexerUrl := "http://127.0.0.1:8116"
	return rpc.DialWithIndexer(ckbUrl, indexerUrl)
}

func getNewDasCoreMainNet() (*core.DasCore, error) {
	client, err := getClientMainNet()
	if err != nil {
		return nil, err
	}

	env := core.InitEnvOpt(common.DasNetTypeMainNet,
		common.DasContractNameConfigCellType,
		common.DasContractNameAccountCellType,
		common.DasContractNameDispatchCellType,
		common.DasContractNameBalanceCellType,
		common.DasContractNameAlwaysSuccess,
		//common.DasContractNameIncomeCellType,
		//common.DASContractNameSubAccountCellType,
		common.DasContractNamePreAccountCellType,
	)
	var wg sync.WaitGroup
	ops := []core.DasCoreOption{
		core.WithClient(client),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(common.DasNetTypeMainNet),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dc := core.NewDasCore(context.Background(), &wg, ops...)
	// contract
	dc.InitDasContract(env.MapContract)
	// config cell
	if err = dc.InitDasConfigCell(); err != nil {
		return nil, err
	}
	// so script
	if err = dc.InitDasSoScript(); err != nil {
		return nil, err
	}
	return dc, nil
}
