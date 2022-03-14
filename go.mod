module das_register_server

go 1.15

require (
	github.com/DeAccountSystems/das-lib v0.0.0-20220314090719-b2a743f77dab
	github.com/elazarl/goproxy v0.0.0-20211114080932-d06c3be7c11b // indirect
	github.com/ethereum/go-ethereum v1.10.13
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.7.2
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	github.com/nervosnetwork/ckb-sdk-go v0.101.3
	github.com/parnurzeal/gorequest v0.2.16
	github.com/scorpiotzh/mylog v1.0.9
	github.com/scorpiotzh/toolib v1.1.3
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.3.0
	gorm.io/gorm v1.22.1
	moul.io/http2curl v1.0.0 // indirect
)

replace github.com/ethereum/go-ethereum v1.10.13 => github.com/pranksteess/go-ethereum v1.10.15-0.20211214035109-e01bfb488ddb
