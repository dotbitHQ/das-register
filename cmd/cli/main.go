package main

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"flag"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/scorpiotzh/mylog"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

var (
	log               = mylog.NewLogger("main", mylog.LevelDebug)
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

type RewardData struct {
	Account   string           `json:"account"`
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Amount    decimal.Decimal  `json:"amount"`
}

func main() {
	// xxxx.bit 50
	var conf = flag.String("conf", "", "conf file path")
	var file = flag.String("file", "", "reward file path")
	flag.Parse()

	if *conf == "" {
		log.Error("conf is nil")
		return
	}
	if *file == "" {
		log.Error("file is nil")
		return
	}
	if err := config.InitCfg(*conf); err != nil {
		log.Error("InitCfg err: ", err.Error())
		return
	}
	rewardStr, err := readFile(*file)
	if err != nil {
		log.Error("readFile err: %s", err.Error())
		return
	}
	log.Info("readFile:", rewardStr)

	// db
	dbDao, err := dao.NewGormDB(config.Cfg.DB.Mysql, config.Cfg.DB.ParserMysql)
	if err != nil {
		log.Error("dao.NewGormDB err: ", err.Error())
		return
	}
	log.Info("db ok")

	// ckb 节点
	ckbClient, err := rpc.DialWithIndexer(config.Cfg.Chain.CkbUrl, config.Cfg.Chain.IndexUrl)
	if err != nil {
		log.Error("rpc.DialWithIndexer err: ", err.Error())
		return
	}
	log.Info("ckb node ok")

	// das 合约初始化
	env := core.InitEnvOpt(config.Cfg.Server.Net, common.DasContractNameConfigCellType, common.DasContractNameAccountCellType,
		common.DasContractNameBalanceCellType, common.DasContractNameDispatchCellType, common.DasContractNameApplyRegisterCellType,
		common.DasContractNamePreAccountCellType, common.DasContractNameProposalCellType, common.DasContractNameReverseRecordCellType,
		common.DasContractNameIncomeCellType, common.DasContractNameAlwaysSuccess)
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
		log.Error("InitDasConfigCell err: ", err.Error())
		return
	}
	if err := dasCore.InitDasSoScript(); err != nil {
		log.Error("InitDasSoScript err: ", err.Error())
		return
	}
	dasCore.RunAsyncDasContract(time.Minute)   // contract outpoint
	dasCore.RunAsyncDasConfigCell(time.Minute) // config cell outpoint
	dasCore.RunAsyncDasSoScript(time.Minute)   // so

	log.Info("das contract ok")

	// 交易构造器
	//	txBuilderBase := txbuilder.NewDasTxBuilderBase(ctxServer, dasCore, nil, "")

	// 查询待发奖列表
	mapReward, err := getRewardList(dbDao, rewardStr)
	if err != nil {
		log.Error("getRewardList err: %s", err.Error())
		return
	}
	fmt.Println(mapReward)

	// 构建交易
	// 签名
	// 发交易

	cancel()
}
func readFile(fPath string) (string, error) {
	data, err := ioutil.ReadFile(fPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func getRewardList(db *dao.DbDao, rewardStr string) (map[string]RewardData, error) {
	strList := strings.Split(rewardStr, "\n")
	var mapReward = make(map[string]RewardData)
	var accounts []string
	for _, v := range strList {
		if v == "" {
			continue
		}
		arr := strings.Split(v, " ")
		amount, _ := decimal.NewFromString(arr[1])
		mapReward[arr[0]] = RewardData{
			Account:   arr[0],
			ChainType: 0,
			Address:   "",
			Amount:    amount,
		}
		accounts = append(accounts, arr[0])
	}
	// 查询库
	list, err := db.GetAccounts(accounts)
	if err != nil {
		return nil, fmt.Errorf("GetAccounts err: %s", err.Error())
	}
	for _, v := range list {
		if item, ok := mapReward[v.Account]; ok {
			item.ChainType = v.OwnerChainType
			item.Address = v.Owner
			mapReward[v.Account] = item
		}
	}
	return mapReward, nil
}
