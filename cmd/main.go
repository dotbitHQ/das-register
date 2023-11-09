package main

import (
	"context"
	"das_register_server/block_parser"
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/http_server"
	"das_register_server/timer"
	"das_register_server/txtool"
	"das_register_server/unipay"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/remote_sign"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"github.com/urfave/cli/v2"
	"os"
	"sync"
	"time"
)

var (
	log               = logger.NewLogger("main", logger.LevelDebug)
	exit              = make(chan struct{})
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

func main() {
	log.Debugf("start：")
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	// config file
	configFilePath := ctx.String("config")
	if err := config.InitCfg(configFilePath); err != nil {
		return err
	}

	// config file watcher
	watcher, err := config.AddCfgFileWatcher(configFilePath)
	if err != nil {
		return err
	}
	// ============= service start =============

	//sentry
	if err := http_api.SentryInit(config.Cfg.Notify.SentryDsn); err != nil {
		return fmt.Errorf("SentryInit err: %s", err.Error())
	}
	defer http_api.RecoverPanic()

	// db
	dbDao, err := dao.NewGormDB(config.Cfg.DB.Mysql, config.Cfg.DB.ParserMysql)
	if err != nil {
		return fmt.Errorf("dao.NewGormDB err: %s", err.Error())
	}
	log.Info("db ok")

	// redis
	red, err := toolib.NewRedisClient(config.Cfg.Cache.Redis.Addr, config.Cfg.Cache.Redis.Password, config.Cfg.Cache.Redis.DbNum)
	if err != nil {
		log.Info("NewRedisClient err: %s", err.Error())
		//return fmt.Errorf("NewRedisClient err:%s", err.Error())
	} else {
		log.Info("redis ok")
	}
	rc := cache.Initialize(red)

	// das core
	dasCore, dasCache, err := initDasCore()
	if err != nil {
		return fmt.Errorf("initDasCore err: %s", err.Error())
	}

	// tx builder
	txBuilderBase, serverScript, err := initTxBuilder(dasCore)
	if err != nil {
		return fmt.Errorf("initTxBuilder err: %s", err.Error())
	}

	// service timer
	txTimer := timer.NewTxTimer(timer.TxTimerParam{
		Ctx:           ctxServer,
		Wg:            &wgServer,
		DbDao:         dbDao,
		DasCore:       dasCore,
		DasCache:      dasCache,
		TxBuilderBase: txBuilderBase,
	})
	if err = txTimer.Run(); err != nil {
		return fmt.Errorf("txTimer.Run() err: %s", err.Error())
	}
	txTimer.DoRecyclePreEarly()
	log.Info("timer ok")

	if config.Cfg.Server.UniPayUrl != "" {
		toolUniPay := unipay.ToolUniPay{
			Ctx:   ctxServer,
			Wg:    &wgServer,
			DbDao: dbDao,
		}
		toolUniPay.RunConfirmStatus()
		toolUniPay.RunOrderRefund()
		toolUniPay.RunDoOrderHedge()
		toolUniPay.RunRegisterInfo()
	}

	// tx timer
	txTool := &txtool.TxTool{
		Ctx:           ctxServer,
		Wg:            &wgServer,
		DbDao:         dbDao,
		DasCore:       dasCore,
		DasCache:      dasCache,
		TxBuilderBase: txBuilderBase,
		ServerScript:  serverScript,
		RebootTime:    time.Now(),
	}
	txTool.Run()

	// prometheus
	//prometheus.Init()
	//prometheus.Tools.Run()

	// block parser
	bp := block_parser.BlockParser{
		DasCore:            dasCore,
		CurrentBlockNumber: config.Cfg.Chain.CurrentBlockNumber,
		DbDao:              dbDao,
		ConcurrencyNum:     config.Cfg.Chain.ConcurrencyNum,
		ConfirmNum:         config.Cfg.Chain.ConfirmNum,
		Ctx:                ctxServer,
		Cancel:             cancel,
		Wg:                 &wgServer,
	}
	if err := bp.Run(); err != nil {
		return fmt.Errorf("block parser err: %s", err.Error())
	}
	log.Info("block parser ok")

	//
	builderConfigCell, err := dasCore.ConfigCellDataBuilderByTypeArgsList(
		common.ConfigCellTypeArgsPreservedAccount00,
		common.ConfigCellTypeArgsPreservedAccount01,
		common.ConfigCellTypeArgsPreservedAccount02,
		common.ConfigCellTypeArgsPreservedAccount03,
		common.ConfigCellTypeArgsPreservedAccount04,
		common.ConfigCellTypeArgsPreservedAccount05,
		common.ConfigCellTypeArgsPreservedAccount06,
		common.ConfigCellTypeArgsPreservedAccount07,
		common.ConfigCellTypeArgsPreservedAccount08,
		common.ConfigCellTypeArgsPreservedAccount09,
		common.ConfigCellTypeArgsPreservedAccount10,
		common.ConfigCellTypeArgsPreservedAccount11,
		common.ConfigCellTypeArgsPreservedAccount12,
		common.ConfigCellTypeArgsPreservedAccount13,
		common.ConfigCellTypeArgsPreservedAccount14,
		common.ConfigCellTypeArgsPreservedAccount15,
		common.ConfigCellTypeArgsPreservedAccount16,
		common.ConfigCellTypeArgsPreservedAccount17,
		common.ConfigCellTypeArgsPreservedAccount18,
		common.ConfigCellTypeArgsPreservedAccount19,
		common.ConfigCellTypeArgsUnavailable,
	)
	if err != nil {
		return fmt.Errorf("unavailable account and preserved account init err: %s", err.Error())
	}

	// http service
	hs, err := http_server.Initialize(http_server.HttpServerParams{
		Address:                config.Cfg.Server.HttpServerAddr,
		InternalAddress:        config.Cfg.Server.HttpServerInternalAddr,
		DbDao:                  dbDao,
		Rc:                     rc,
		Ctx:                    ctxServer,
		DasCore:                dasCore,
		DasCache:               dasCache,
		TxBuilderBase:          txBuilderBase,
		ServerScript:           serverScript,
		MapReservedAccounts:    builderConfigCell.ConfigCellPreservedAccountMap,
		MapUnAvailableAccounts: builderConfigCell.ConfigCellUnavailableAccountMap,
	})
	if err != nil {
		return fmt.Errorf("http server Initialize err:%s", err.Error())
	}
	hs.Run()
	log.Info("httpserver ok")

	//nameDaoTimer := timer.NameDaoTimer{DbDao: dbDao}
	//nameDaoTimer.RunCheckNameDaoMember()

	// ============= service end =============
	toolib.ExitMonitoring(func(sig os.Signal) {
		log.Warn("ExitMonitoring:", sig.String())
		if watcher != nil {
			log.Warn("close watcher ... ")
			_ = watcher.Close()
		}
		cancel()
		hs.Shutdown()
		//nameDaoTimer.CloseCron()
		txTimer.CloseCron()

		wgServer.Wait()
		log.Warn("success exit server. bye bye!")
		time.Sleep(time.Second)
		exit <- struct{}{}
	})

	<-exit

	return nil
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
		common.DASContractNameSubAccountCellType, common.DasKeyListCellType, common.DasContractNameDpCellType)
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
		//remoteSignClient, err := sign.NewClient(ctxServer, config.Cfg.Server.RemoteSignApiUrl)
		//if err != nil {
		//	return nil, nil, fmt.Errorf("sign.NewClient err: %s", err.Error())
		//}
		//handleSign = sign.RemoteSign(remoteSignClient, config.Cfg.Server.Net, payServerAddressArgs)
		handleSign = remote_sign.SignTxForCKBHandle(config.Cfg.Server.RemoteSignApiUrl, config.Cfg.Server.PayServerAddress)
	} else if config.Cfg.Server.PayPrivate != "" {
		handleSign = sign.LocalSign(config.Cfg.Server.PayPrivate)
	}
	txBuilderBase := txbuilder.NewDasTxBuilderBase(ctxServer, dasCore, handleSign, payServerAddressArgs)
	log.Info("tx builder ok")

	return txBuilderBase, serverScript, nil
}
