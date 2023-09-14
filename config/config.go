package config

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"time"
)

var (
	Cfg CfgServer
	log = logger.NewLogger("config", logger.LevelDebug)
)

func InitCfg(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "../config/config.yaml"
	}
	log.Info("config file：", configFilePath)
	if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
		return fmt.Errorf("UnmarshalYamlFile err:%s", err.Error())
	}
	private := Cfg.Server.PayPrivate
	Cfg.Server.PayPrivate = ""
	log.Info("config file：", toolib.JsonString(Cfg))
	Cfg.Server.PayPrivate = private
	return nil
}

func AddCfgFileWatcher(configFilePath string) (*fsnotify.Watcher, error) {
	if configFilePath == "" {
		configFilePath = "../config/config.yaml"
	}
	return toolib.AddFileWatcher(configFilePath, func() {
		log.Info("update config file：", configFilePath)
		if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
			log.Error("UnmarshalYamlFile err:", err.Error())
		}
		private := Cfg.Server.PayPrivate
		Cfg.Server.PayPrivate = ""
		log.Info("update config file：", toolib.JsonString(Cfg))
		Cfg.Server.PayPrivate = private
	})
}

type CfgServer struct {
	Server struct {
		IsUpdate                bool              `json:"is_update" yaml:"is_update"`
		Net                     common.DasNetType `json:"net" yaml:"net"`
		HttpServerAddr          string            `json:"http_server_addr" yaml:"http_server_addr"`
		HttpServerInternalAddr  string            `json:"http_server_internal_addr" yaml:"http_server_internal_addr"`
		PayServerAddress        string            `json:"pay_server_address" yaml:"pay_server_address"`
		PayPrivate              string            `json:"pay_private" yaml:"pay_private"`
		RemoteSignApiUrl        string            `json:"remote_sign_api_url" yaml:"remote_sign_api_url"`
		PushLogUrl              string            `json:"push_log_url" yaml:"push_log_url"`
		PushLogIndex            string            `json:"push_log_index" yaml:"push_log_index"`
		ParserUrl               string            `json:"parser_url" yaml:"parser_url"`
		TxToolSwitch            bool              `json:"tx_tool_switch" yaml:"tx_tool_switch"`
		SplitCkb                uint64            `json:"split_ckb" yaml:"split_ckb"`
		RecoverCkb              uint64            `json:"recover_ckb" yaml:"recover_ckb"`
		RecoverTime             time.Duration     `json:"recover_time" yaml:"recover_time"`
		RecycleAllPre           bool              `json:"recycle_all_pre" yaml:"recycle_all_pre"`
		RecyclePreEarly         bool              `json:"recycle_pre_early" yaml:"recycle_pre_early"`
		RecyclePreEarlyCronSpec string            `json:"recycle_pre_early_cron_spec" yaml:"recycle_pre_early_cron_spec"`
		NotExit                 bool              `json:"not_exit" yaml:"not_exit"`
		CouponFilePath          string            `json:"coupon_file_path" yaml:"coupon_file_path"`
		CouponEncrySalt         string            `json:"coupon_encry_salt" yaml:"coupon_encry_salt"`
		CouponQrcodePrefix      string            `json:"coupon_qrcode_prefix" yaml:"coupon_qrcode_prefix"`
		CouponCodeLength        uint8             `json:"coupon_code_length" yaml:"coupon_code_length"`
		UniPayUrl               string            `json:"uni_pay_url" yaml:"uni_pay_url"`
		UniPayRefundSwitch      bool              `json:"uni_pay_refund_switch" yaml:"uni_pay_refund_switch"`
		HedgeUrl                string            `json:"hedge_url" yaml:"hedge_url"`
	} `json:"server" yaml:"server"`
	Origins          []string          `json:"origins" yaml:"origins"`
	InviterWhitelist map[string]string `json:"inviter_whitelist" yaml:"inviter_whitelist"`
	Notify           struct {
		LarkErrorKey      string `json:"lark_error_key" yaml:"lark_error_key"`
		LarkRegisterKey   string `json:"lark_register_key" yaml:"lark_register_key"`
		LarkRegisterOkKey string `json:"lark_register_ok_key" yaml:"lark_register_ok_key"`
		LarkDasInfoKey    string `json:"lark_das_info_key" yaml:"lark_das_info_key"`
		DiscordWebhook    string `json:"discord_webhook" yaml:"discord_webhook"`
		SentryDsn         string `json:"sentry_dsn" yaml:"sentry_dsn"`
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
		Discount             decimal.Decimal `json:"discount" yaml:"discount"`
	} `json:"das" yaml:"das"`
	Stripe struct {
		PremiumPercentage decimal.Decimal `json:"premium_percentage" yaml:"premium_percentage"`
		PremiumBase       decimal.Decimal `json:"premium_base" yaml:"premium_base"`
	} `json:"stripe" yaml:"stripe"`
}

type DbMysql struct {
	Addr        string `json:"addr" yaml:"addr"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"`
	DbName      string `json:"db_name" yaml:"db_name"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"max_open_conn"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"max_idle_conn"`
}

func GetUnipayAddress(tokenId tables.PayTokenId) string {
	switch tokenId {
	case tables.TokenIdEth, tables.TokenIdErc20USDT:
		return Cfg.PayAddressMap["eth"]
	case tables.TokenIdBnb, tables.TokenIdBep20USDT:
		return Cfg.PayAddressMap["bsc"]
	case tables.TokenIdMatic:
		return Cfg.PayAddressMap["polygon"]
	case tables.TokenIdTrx, tables.TokenIdTrc20USDT:
		return Cfg.PayAddressMap["tron"]
	case tables.TokenIdCkb, tables.TokenIdDas:
		return Cfg.PayAddressMap["ckb"]
	case tables.TokenIdDoge:
		return Cfg.PayAddressMap["doge"]
	case tables.TokenIdStripeUSD:
		return "stripe"
	}
	log.Error("GetUnipayAddress not supported:", tokenId)
	return ""
}
