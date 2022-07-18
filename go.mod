module das_register_server

go 1.16

require (
	github.com/dotbitHQ/das-lib v0.0.0-20220718022750-cfcd2c5b88b7
	github.com/elazarl/goproxy v0.0.0-20220417044921-416226498f94 // indirect
	github.com/ethereum/go-ethereum v1.10.17
	github.com/fsnotify/fsnotify v1.5.3
	github.com/gin-gonic/gin v1.7.7
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/nervosnetwork/ckb-sdk-go v0.101.3
	github.com/parnurzeal/gorequest v0.2.16
	github.com/scorpiotzh/mylog v1.0.10
	github.com/scorpiotzh/toolib v1.1.3
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.4.4
	gorm.io/gorm v1.22.1
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/ethereum/go-ethereum v1.9.14 => github.com/ethereum/go-ethereum v1.10.17
