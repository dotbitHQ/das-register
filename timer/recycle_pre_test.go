package timer

import (
	"context"
	"das_register_server/config"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"sync"
	"testing"
)

func TestRecyclePre(t *testing.T) {
	dc, err := getNewDasCoreTestnet2()
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
	if err := txTimer.doRecyclePre(); err != nil {
		t.Fatal(err)
	}
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

func initTxBuilder(dasCore *core.DasCore) (*txbuilder.DasTxBuilderBase, error) {
	var handleSign sign.HandleSignCkbMessage
	if config.Cfg.Server.PayPrivate != "" {
		handleSign = sign.LocalSign(config.Cfg.Server.PayPrivate)
	}
	txBuilderBase := txbuilder.NewDasTxBuilderBase(context.Background(), dasCore, handleSign, "")
	log.Info("tx builder ok")

	return txBuilderBase, nil
}
