package timer

import (
	"context"
	"das_register_server/block_parser"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/txtool"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"sync"
	"testing"
	"time"
)

var (
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}

	txTimer *TxTimer
	txTool  *txtool.TxTool
	bp      *block_parser.BlockParser
)

func TestMain(m *testing.M) {
	configPath := "../config/config.yaml"
	// config file
	if err := config.InitCfg(configPath); err != nil {
		panic(err)
	}

	// db
	dbDao, err := dao.NewGormDB(config.Cfg.DB.Mysql, config.Cfg.DB.ParserMysql)
	if err != nil {
		panic(err)
	}

	// das core
	dasCore, dasCache, err := initDasCore()
	if err != nil {
		panic(err)
	}

	// tx builder
	txBuilderBase, serverScript, err := initTxBuilder(dasCore)
	if err != nil {
		panic(err)
	}

	// reverse smt tx builder
	reverseSmtTxBuilder, reverseSmtServerScript, err := initReverseSmtTxBuilder(dasCore)
	if err != nil {
		panic(err)
	}

	// service timer
	txTimer = NewTxTimer(TxTimerParam{
		Ctx:                    ctxServer,
		Wg:                     &wgServer,
		DbDao:                  dbDao,
		DasCore:                dasCore,
		DasCache:               dasCache,
		TxBuilderBase:          txBuilderBase,
		ReverseSmtTxBuilder:    reverseSmtTxBuilder,
		ReverseSmtServerScript: reverseSmtServerScript,
	})

	txTool = &txtool.TxTool{
		Ctx:           ctxServer,
		Wg:            &wgServer,
		DbDao:         dbDao,
		DasCore:       dasCore,
		DasCache:      dasCache,
		TxBuilderBase: txBuilderBase,
		ServerScript:  serverScript,
		RebootTime:    time.Now(),
	}

	// block parser
	bp = &block_parser.BlockParser{
		DasCore:            dasCore,
		CurrentBlockNumber: config.Cfg.Chain.CurrentBlockNumber,
		DbDao:              dbDao,
		ConcurrencyNum:     config.Cfg.Chain.ConcurrencyNum,
		ConfirmNum:         config.Cfg.Chain.ConfirmNum,
		Ctx:                ctxServer,
		Cancel:             cancel,
		Wg:                 &wgServer,
	}
	m.Run()
}

func initDasCore() (*core.DasCore, *dascache.DasCache, error) {
	// ckb node
	ckbClient, err := rpc.DialWithIndexer(config.Cfg.Chain.CkbUrl, config.Cfg.Chain.IndexUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("rpc.DialWithIndexer err: %s", err.Error())
	}
	log.Info("ckb node ok")

	// das init
	env := core.InitEnvOpt(config.Cfg.Server.Net, common.DasContractNameConfigCellType, common.DasContractNameAccountCellType,
		common.DasContractNameBalanceCellType, common.DasContractNameDispatchCellType, common.DasContractNameApplyRegisterCellType,
		common.DasContractNamePreAccountCellType, common.DasContractNameProposalCellType, common.DasContractNameReverseRecordCellType,
		common.DasContractNameIncomeCellType, common.DasContractNameAlwaysSuccess, common.DASContractNameEip712LibCellType,
		common.DASContractNameSubAccountCellType)
	ops := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(config.Cfg.Server.Net),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dasCore := core.NewDasCore(ctxServer, &wgServer, ops...)
	dasCore.InitDasContract(env.MapContract)
	if err := dasCore.InitDasConfigCell(); err != nil {
		return nil, nil, fmt.Errorf("InitDasConfigCell err: %s", err.Error())
	}
	if err := dasCore.InitDasSoScript(); err != nil {
		return nil, nil, fmt.Errorf("InitDasSoScript err: %s", err.Error())
	}
	dasCore.RunAsyncDasContract(time.Minute * 3)   // contract outpoint
	dasCore.RunAsyncDasConfigCell(time.Minute * 5) // config cell outpoint
	dasCore.RunAsyncDasSoScript(time.Minute * 7)   // so

	log.Info("das contract ok")

	// das cache
	dasCache := dascache.NewDasCache(ctxServer, &wgServer)
	dasCache.RunClearExpiredOutPoint(time.Minute * 15)
	log.Info("das cache ok")

	return dasCore, dasCache, nil
}

func initTxBuilder(dasCore *core.DasCore) (*txbuilder.DasTxBuilderBase, *types.Script, error) {
	payServerAddressArgs := ""
	var serverScript *types.Script
	if config.Cfg.Server.PayServerAddress != "" {
		parseAddress, err := address.Parse(config.Cfg.Server.PayServerAddress)
		if err != nil {
			log.Error("pay server address.Parse err: ", err.Error())
		} else {
			payServerAddressArgs = common.Bytes2Hex(parseAddress.Script.Args)
			serverScript = parseAddress.Script
		}
	}
	var handleSign sign.HandleSignCkbMessage
	if config.Cfg.Server.RemoteSignApiUrl != "" && payServerAddressArgs != "" {
		remoteSignClient, err := sign.NewClient(ctxServer, config.Cfg.Server.RemoteSignApiUrl)
		if err != nil {
			return nil, nil, fmt.Errorf("sign.NewClient err: %s", err.Error())
		}
		handleSign = sign.RemoteSign(remoteSignClient, config.Cfg.Server.Net, payServerAddressArgs)
	} else if config.Cfg.Server.PayPrivate != "" {
		handleSign = sign.LocalSign(config.Cfg.Server.PayPrivate)
	}
	txBuilderBase := txbuilder.NewDasTxBuilderBase(ctxServer, dasCore, handleSign, payServerAddressArgs)
	log.Info("tx builder ok")

	return txBuilderBase, serverScript, nil
}

func initReverseSmtTxBuilder(dasCore *core.DasCore) (*txbuilder.DasTxBuilderBase, *types.Script, error) {
	parseAddress, err := address.Parse(config.Cfg.Server.ReverseSmtPayServerAddress)
	if err != nil {
		log.Error("initReverseSmtTxBuilder address.Parse err: ", err.Error())
		return nil, nil, err
	}
	payServerAddressArgs := common.Bytes2Hex(parseAddress.Script.Args)
	serverScript := parseAddress.Script

	remoteSignClient, err := sign.NewClient(ctxServer, config.Cfg.Server.RemoteSignApiUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("sign.NewClient err: %s", err.Error())
	}
	handleSign := sign.RemoteSign(remoteSignClient, config.Cfg.Server.Net, payServerAddressArgs)

	txBuilderBase := txbuilder.NewDasTxBuilderBase(ctxServer, dasCore, handleSign, payServerAddressArgs)
	log.Info("reverse smt tx builder ok")

	return txBuilderBase, serverScript, nil
}
