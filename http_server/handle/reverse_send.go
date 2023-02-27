package handle

import (
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

type ReqReverseSend struct {
	SignKey   string `json:"sign_key"`
	Signature string `json:"signature"`
}

func (h *HttpHandle) ReverseSend(ctx *gin.Context) {
	var (
		funcName = "ReverseSend"
		clientIp = GetClientIp(ctx)
		req      ReqReverseSend
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doReverseSend(&req, &apiResp); err != nil {
		log.Error("doReverseSend err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseSend(req *ReqReverseSend, apiResp *api_code.ApiResp) error {
	cacheData := &ReverseSmtSignCache{}
	if err := h.rc.GetSignTxCacheData(req.SignKey, cacheData); err != nil {
		if err == redis.Nil {
			apiResp.ApiRespErr(api_code.ApiCodeTxExpired, "tx expired err")
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeCacheError, "cache err")
		}
		return fmt.Errorf("GetSignTxCache err: %s", err.Error())
	}

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	lockDone := make(chan struct{})
	lockKey := fmt.Sprintf("lock:doReverseSend:%s", cacheData.AddressHex)
	if err := h.rc.Lock(lockKey, time.Second*3, func(lockFn func()) {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for range t.C {
			select {
			case <-lockDone:
				return
			default:
				lockFn()
			}
		}
	}); err != nil {
		if err == cache.ErrDistributedLockPreemption {
			apiResp.ApiRespErr(api_code.ApiCodeCacheError, err.Error())
			return fmt.Errorf("cache err: %s", err.Error())
		}
		apiResp.ApiRespErr(api_code.ApiCodeCacheError, "cache error")
		return fmt.Errorf("cache err: %s", err.Error())
	}
	defer func() {
		if err := h.rc.DelSignTxCache(req.SignKey); err != nil {
			log.Errorf("doReverseSend DelSignTxCache err: %s", err)
		}
		close(lockDone)
	}()

	// check user sign msg
	if _, err := doReverseSmtSignCheck(cacheData, req.Signature, apiResp); err != nil {
		return err
	}
	if err := h.dbDao.CreateReverseSmtRecord(&tables.ReverseSmtRecordInfo{
		Address:     cacheData.AddressHex,
		AlgorithmID: uint8(cacheData.DasAlgorithmId),
		Nonce:       cacheData.Nonce,
		Account:     cacheData.Account,
		Sign:        req.Signature,
		SubAction:   cacheData.Action,
	}); err != nil {
		return err
	}
	return nil
}

// doReverseSmtSignCheck
func doReverseSmtSignCheck(cacheData *ReverseSmtSignCache, signMsg string, apiResp *api_code.ApiResp) (string, error) {
	signOk := false
	var err error

	signAddress := cacheData.AddressHex

	switch cacheData.DasAlgorithmId {
	case common.DasAlgorithmIdEth:
		signMsg = fixSignature(signMsg)
		signOk, err = sign.VerifyPersonalSignature(common.Hex2Bytes(signMsg), common.Hex2Bytes(cacheData.SignMsg), signAddress)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeSigErr, "eth sign error")
			return "", fmt.Errorf("VerifyPersonalSignature err: %s", err)
		}
	case common.DasAlgorithmIdTron:
		signMsg = fixSignature(signMsg)
		if signAddress, err = common.TronHexToBase58(signAddress); err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeSigErr, "TronHexToBase58 error")
			return "", fmt.Errorf("TronHexToBase58 err: %s [%s]", err, signAddress)
		}
		signOk = sign.TronVerifySignature(true, common.Hex2Bytes(signMsg), common.Hex2Bytes(cacheData.SignMsg), signAddress)
	case common.DasAlgorithmIdEd25519:
		signOk = sign.VerifyEd25519Signature(common.Hex2Bytes(signAddress), common.Hex2Bytes(cacheData.SignMsg), common.Hex2Bytes(signMsg))
	default:
		apiResp.ApiRespErr(api_code.ApiCodeNotExistSignType, fmt.Sprintf("not exist sign type[%d]", cacheData.DasAlgorithmId))
		return "", nil
	}
	if !signOk {
		apiResp.ApiRespErr(api_code.ApiCodeSigErr, "res sign error")
	}
	return signMsg, nil
}

// fixSignature
func fixSignature(signMsg string) string {
	if len(signMsg) >= 132 && signMsg[130:132] == "1b" {
		signMsg = signMsg[0:130] + "00" + signMsg[132:]
	}
	if len(signMsg) >= 132 && signMsg[130:132] == "1c" {
		signMsg = signMsg[0:130] + "01" + signMsg[132:]
	}
	return signMsg
}
