module das_register_server

go 1.16

require (
	github.com/dotbitHQ/das-lib v1.0.2-0.20230414093428-df86fbc5e925
	github.com/ethereum/go-ethereum v1.10.17
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gin-gonic/gin v1.7.7
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/nervosnetwork/ckb-sdk-go v0.101.3
	github.com/parnurzeal/gorequest v0.2.16
	github.com/robfig/cron/v3 v3.0.1
	github.com/scorpiotzh/mylog v1.0.10
	github.com/scorpiotzh/toolib v1.1.6-0.20230210123015-9770bc1afe72
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.4.4
	gorm.io/gorm v1.23.6
)

replace github.com/ethereum/go-ethereum v1.9.14 => github.com/ethereum/go-ethereum v1.10.17
