package timer

import (
	"context"
	"das_register_server/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
	"sync"
	"testing"
)

func TestSince(t *testing.T) {
	fmt.Println(utils.SinceFromRelativeTimestamp(5 * 60))
	fmt.Println(utils.SinceFromRelativeTimestamp(60 * 60))
	fmt.Println(utils.SinceFromRelativeTimestamp(24 * 60 * 60))
}

func TestRecyclePre(t *testing.T) {
	config.Cfg.Server.Net = common.DasNetTypeTestnet2
	config.Cfg.Server.PayPrivate = ""
	config.Cfg.Server.PayServerAddress = ""
	config.Cfg.Server.RecyclePreEarly = true
	dc, err := getNewDasCoreTestnet2() //getNewDasCoreTestnet2()
	if err != nil {
		t.Fatal(err)
	}
	txBuilderBase, err := initTxBuilder(dc)
	if err != nil {
		t.Fatal(err)
	}
	txTimer := NewTxTimer(TxTimerParam{
		Ctx:           context.Background(),
		Wg:            nil,
		DbDao:         nil,
		DasCore:       dc,
		TxBuilderBase: txBuilderBase,
	})
	//if err := txTimer.doRecyclePre(); err != nil {
	//	t.Fatal(err)
	//}

	if err := txTimer.doRecyclePreEarly(); err != nil {
		t.Fatal(err)
	}
}

func getClientTestnet2() (rpc.Client, error) {
	ckbUrl := "https://testnet.ckb.dev"
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
		common.DasContractNameDispatchCellType,
		common.DasContractNameBalanceCellType,
		common.DasContractNameAlwaysSuccess,
		//common.DasContractNameIncomeCellType,
		//common.DASContractNameSubAccountCellType,
		common.DasContractNamePreAccountCellType,
		//common.DASContractNameEip712LibCellType,
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
		//common.DasContractNameAccountCellType,
		common.DasContractNameDispatchCellType,
		common.DasContractNameBalanceCellType,
		common.DasContractNameAlwaysSuccess,
		//common.DasContractNameIncomeCellType,
		//common.DASContractNameSubAccountCellType,
		common.DasContractNamePreAccountCellType,
		//common.DASContractNameEip712LibCellType,
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

func initTxBuilder(dasCore *core.DasCore) (*txbuilder.DasTxBuilderBase, error) {
	var handleSign sign.HandleSignCkbMessage
	if config.Cfg.Server.PayPrivate != "" {
		handleSign = sign.LocalSign(config.Cfg.Server.PayPrivate)
	}
	payServerAddressArgs := ""
	if config.Cfg.Server.PayServerAddress != "" {
		parseAddress, err := address.Parse(config.Cfg.Server.PayServerAddress)
		if err != nil {
			log.Error("pay server address.Parse err: ", err.Error())
		} else {
			payServerAddressArgs = common.Bytes2Hex(parseAddress.Script.Args)
		}
	}
	txBuilderBase := txbuilder.NewDasTxBuilderBase(context.Background(), dasCore, handleSign, payServerAddressArgs)
	log.Info("tx builder ok")

	return txBuilderBase, nil
}

func TestAddress(t *testing.T) {
	res, err := address.GenerateAddress(address.Testnet)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res.Address, res.PrivateKey)
}
