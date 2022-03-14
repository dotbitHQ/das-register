package config

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/fsnotify/fsnotify"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"strings"
)

var (
	Cfg                  CfgServer
	AccountCharSetEmoji  string
	AccountCharSetNumber = "0123456789"
	AccountCharSetEn     = "abcdefghijklmnopqrstuvwxyz."
	log                  = mylog.NewLogger("config", mylog.LevelDebug)
)

func InitCfg(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "./conf/config.yaml"
	}
	log.Info("config file：", configFilePath)
	if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
		return fmt.Errorf("UnmarshalYamlFile err:%s", err.Error())
	}
	log.Info("config file：", toolib.JsonString(Cfg))
	return nil
}

func AddCfgFileWatcher(configFilePath string) (*fsnotify.Watcher, error) {
	if configFilePath == "" {
		configFilePath = "./conf/config.yaml"
	}
	return toolib.AddFileWatcher(configFilePath, func() {
		log.Info("update config file：", configFilePath)
		if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
			log.Error("UnmarshalYamlFile err:", err.Error())
		}
		log.Info("update config file：", toolib.JsonString(Cfg))
	})
}

type CfgServer struct {
	Server struct {
		IsUpdate               bool              `json:"is_update" yaml:"is_update"`
		Net                    common.DasNetType `json:"net" yaml:"net"`
		HttpServerAddr         string            `json:"http_server_addr" yaml:"http_server_addr"`
		HttpServerInternalAddr string            `json:"http_server_internal_addr" yaml:"http_server_internal_addr"`
		PayServerAddress       string            `json:"pay_server_address" yaml:"pay_server_address"`
		PayPrivate             string            `json:"pay_private" yaml:"pay_private"`
		RemoteSignApiUrl       string            `json:"remote_sign_api_url" yaml:"remote_sign_api_url"`
		PushLogUrl             string            `json:"push_log_url" yaml:"push_log_url"`
		PushLogIndex           string            `json:"push_log_index" yaml:"push_log_index"`
		ParserUrl              string            `json:"parser_url" yaml:"parser_url"`
		TxToolSwitch           bool              `json:"tx_tool_switch" yaml:"tx_tool_switch"`
	} `json:"server" yaml:"server"`
	Origins []string `json:"origins" yaml:"origins"`
	Notify  struct {
		LarkErrorKey      string `json:"lark_error_key" yaml:"lark_error_key"`
		LarkRegisterKey   string `json:"lark_register_key" yaml:"lark_register_key"`
		LarkRegisterOkKey string `json:"lark_register_ok_key" yaml:"lark_register_ok_key"`
		DiscordWebhook    string `json:"discord_webhook" yaml:"discord_webhook"`
	} `json:"notify" yaml:"notify"`
	PayAddressMap map[string]string `json:"pay_address_map" yaml:"pay_address_map"`
	Chain         struct {
		CkbUrl             string `json:"ckb_url" yaml:"ckb_url"`
		IndexUrl           string `json:"index_url" yaml:"index_url"`
		CurrentBlockNumber uint64 `json:"current_block_number" yaml:"current_block_number"`
		ConfirmNum         uint64 `json:"confirm_num" yaml:"confirm_num"`
		ConcurrencyNum     uint64 `json:"concurrency_num" yaml:"concurrency_num"`
	} `json:"chain" yaml:"chain"`
	DB struct {
		Mysql       DbMysql `json:"mysql" yaml:"mysql"`
		ParserMysql DbMysql `json:"parser_mysql" yaml:"parser_mysql"`
	} `json:"db" yaml:"db"`
	Cache struct {
		Redis struct {
			Addr     string `json:"addr" yaml:"addr"`
			Password string `json:"password" yaml:"password"`
			DbNum    int    `json:"db_num" yaml:"db_num"`
		} `json:"redis" yaml:"redis"`
	} `json:"cache" yaml:"cache"`
	Das struct {
		AccountMinLength     uint8           `json:"account_min_length" yaml:"account_min_length"`
		AccountMaxLength     uint8           `json:"account_max_length" yaml:"account_max_length"`
		OpenAccountMinLength uint8           `json:"open_account_min_length" yaml:"open_account_min_length"`
		OpenAccountMaxLength uint8           `json:"open_account_max_length" yaml:"open_account_max_length"`
		MaxRegisterYears     int             `json:"max_register_years" yaml:"max_register_years"`
		Premium              decimal.Decimal `json:"premium" yaml:"premium"`
	} `json:"das" yaml:"das"`
	DasLib struct {
		DasArgs             string                            `json:"das_args" yaml:"das_args"`
		THQCodeHash         string                            `json:"thq_code_hash" yaml:"thq_code_hash"`
		DasContractArgs     string                            `json:"das_contract_args" yaml:"das_contract_args"`
		DasContractCodeHash string                            `json:"das_contract_code_hash" yaml:"das_contract_code_hash"`
		MapDasContract      map[common.DasContractName]string `json:"map_das_contract" yaml:"map_das_contract"`
	} `json:"das_lib" yaml:"das_lib"`
}

type DbMysql struct {
	Addr        string `json:"addr" yaml:"addr"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"`
	DbName      string `json:"db_name" yaml:"db_name"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"max_open_conn"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"max_idle_conn"`
}

func InitAccountCharSetEmoji(dc *core.DasCore) {
	if dc == nil {
		return
	}
	builder, err := dc.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsCharSetEmoji)
	if err != nil {
		log.Error("ConfigCellDataBuilderByTypeArgsList err: ", err.Error())
	} else {
		AccountCharSetEmoji = strings.Join(builder.ConfigCellEmojis, "")
	}
}
