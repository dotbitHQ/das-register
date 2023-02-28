package handle

import (
	"crypto/md5"
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/tables"
	"errors"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ReqReverseUpdate struct {
	core.ChainTypeAddress
	Action  string `json:"action"`
	Account string `json:"account"`
}

type RespReverseUpdate struct {
	SignMsg  string                `json:"sign_msg"`
	SignType common.DasAlgorithmId `json:"sign_type"`
	SignKey  string                `json:"sign_key"`
}

type ReverseSmtSignCache struct {
	Action         string                `json:"action"`
	Account        string                `json:"account"`
	SignMsg        string                `json:"sign_msg"`
	ChainType      common.ChainType      `json:"chain_type"`
	DasAlgorithmId common.DasAlgorithmId `json:"das_algorithm_id"`
	AddressHex     string                `json:"address_hex"`
	Nonce          uint32                `json:"nonce"`
}

func (h *HttpHandle) ReverseUpdate(ctx *gin.Context) {
	var (
		funcName = "ReverseUpdate"
		clientIp = GetClientIp(ctx)
		req      ReqReverseUpdate
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

	if err = h.doReverseUpdate(&req, &apiResp); err != nil {
		log.Error("doReverseUpdate err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseUpdate(req *ReqReverseUpdate, apiResp *api_code.ApiResp) error {
	if req.Action != tables.SubActionUpdate && req.Action != tables.SubActionRemove {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return fmt.Errorf("invalid action, must one of: [%s,%s]", tables.SubActionUpdate, tables.SubActionRemove)
	}

	res := checkReqKeyInfo(h.dasCore.Daf(), &req.ChainTypeAddress, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		log.Error("checkReqReverseRecord:", apiResp.ErrMsg)
		return nil
	}
	address := strings.ToLower(res.AddressHex)

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}
	if exi := h.rc.ApiLimitExist(res.ChainType, res.AddressHex, common.DasActionUpdateReverseRecordRoot); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "The operation is too frequent")
		return fmt.Errorf("api limit: %d %s", res.ChainType, res.AddressHex)
	}

	lockDone := make(chan struct{})
	defer close(lockDone)

	lockKey := fmt.Sprintf("lock:doReverseUpdate:%s", res.AddressHex)
	if err := h.rc.Lock(lockKey, time.Second*10, func(lockFn func()) {
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

	reverse, err := h.dbDao.SearchLatestReverse(res.ChainType, address)
	if err != nil {
		return fmt.Errorf("SearchLatestReverse err: %s", err)
	}
	if req.Action == tables.SubActionRemove {
		if reverse.Id == 0 {
			err = fmt.Errorf("invalid param, reverse no exist")
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
			return err
		} else if reverse.ReverseType == tables.ReverseTypeOld {
			err = fmt.Errorf("invalid param, remove old reverse please use /v1/reverse/retract")
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
			return err
		}
		if reverse.ReverseType == tables.ReverseTypeSmt && reverse.Account != req.Account {
			err = fmt.Errorf("invalid param, reverse no exist")
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
			return err
		}
	}
	if req.Action == tables.SubActionUpdate && reverse.Id > 0 && reverse.Account == req.Account {
		err = fmt.Errorf("invalid param, reverse exist")
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
		return err
	}

	// account check
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
			return fmt.Errorf("SearchAccount err: %s", err.Error())
		}
	}
	if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return fmt.Errorf("account not exist: %s", req.Account)
	}

	if !strings.EqualFold(acc.Owner, address) && !strings.EqualFold(acc.Manager, address) {
		record, err := h.dbDao.SearchAccountReverseRecords(acc.Account, address)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				*apiResp = api_code.ApiRespErr(api_code.ApiCodeDbError, "search account err")
				return fmt.Errorf("SearchAccount err: %s", err.Error())
			}
		}
		if record.Id == 0 {
			err = errors.New("you not this account permissions")
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
			return err
		}
	}

	var resp RespReverseUpdate

	resp.SignType = res.DasAlgorithmId

	// nonce
	nonce, err := h.getReverseSmtNonce(res, req, apiResp)
	if err != nil {
		return err
	}

	dataCache := &ReverseSmtSignCache{
		Action:         req.Action,
		Account:        req.Account,
		SignMsg:        resp.SignMsg,
		ChainType:      res.ChainType,
		DasAlgorithmId: res.DasAlgorithmId,
		AddressHex:     address,
		Nonce:          nonce,
	}
	resp.SignMsg = dataCache.GenSignMsg()

	signKey := dataCache.CacheKey()
	cacheStr := toolib.JsonString(dataCache)
	if err = h.rc.SetSignTxCache(signKey, cacheStr); err != nil {
		return fmt.Errorf("SetSignTxCache err: %s", err.Error())
	}
	resp.SignKey = signKey
	apiResp.Data = resp
	return nil
}

func (cache *ReverseSmtSignCache) CacheKey() string {
	key := fmt.Sprintf("reverse:smt:%s:%d:%s:%s:%s:%d:%d", common.DasActionUpdateReverseRecordRoot, cache.ChainType, cache.AddressHex, cache.Account, cache.Action, cache.Nonce, time.Now().Unix())
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

func (cache *ReverseSmtSignCache) GenSignMsg() string {
	data := make([]byte, 0)
	data = append(data, molecule.GoU32ToBytes(cache.Nonce)...)
	data = append(data, []byte(cache.Account)...)
	bys, _ := blake2b.Blake256(data)
	signMsg := common.Bytes2Hex(append([]byte("from did: "), bys...))

	cache.SignMsg = signMsg
	return signMsg
}

// getReverseSmtNonce
func (h *HttpHandle) getReverseSmtNonce(res *core.DasAddressHex, req *ReqReverseUpdate, apiResp *api_code.ApiResp) (uint32, error) {
	var nonce uint32 = 1

	reverseSmtRecord, err := h.dbDao.GetReverseSmtRecordByAddress(res.AddressHex, uint8(res.DasAlgorithmId))
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "db err")
		return nonce, fmt.Errorf("GetReverseSmtRecordByAddress err: %s address: %s account: %s", err, res.AddressHex, req.Account)
	}

	if reverseSmtRecord.ID == 0 {
		return nonce, nil
	}
	// reverse and task
	if reverseSmtRecord.TaskID == "" {
		apiResp.ApiRespErr(api_code.ApiCodeReverseSmtOnReverse, "reverse is pending")
		return nonce, fmt.Errorf("address: %s account: %s is pending", res.AddressHex, req.Account)
	}
	reverseSmtTaskInfo, err := h.dbDao.GetLatestReverseSmtTaskByTaskID(reverseSmtRecord.TaskID)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "db err")
		return nonce, fmt.Errorf("GetLatestReverseSmtInfoByTaskID err: %s address: %s account: %s", err, res.AddressHex, req.Account)
	}
	if reverseSmtTaskInfo.ID == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "")
		return nonce, fmt.Errorf("GetLatestReverseSmtInfoByTaskID, task_id: %s ,taskInfo can't be null", reverseSmtRecord.TaskID)
	}
	if (reverseSmtTaskInfo.SmtStatus != tables.ReverseSmtStatusConfirm ||
		reverseSmtTaskInfo.TxStatus != tables.ReverseSmtTxStatusConfirm) &&
		reverseSmtTaskInfo.SmtStatus != tables.ReverseSmtStatusRollbackConfirm {
		apiResp.ApiRespErr(api_code.ApiCodeReverseSmtOnReverse, "reverse is pending")
		return nonce, fmt.Errorf("address: %s account: %s is pending", res.AddressHex, req.Account)
	}
	if reverseSmtTaskInfo.SmtStatus == tables.ReverseSmtStatusRollbackConfirm {
		nonce = reverseSmtRecord.Nonce
	} else {
		nonce = reverseSmtRecord.Nonce + 1
	}
	return nonce, nil
}

// checkReqKeyInfo
func checkReqKeyInfo(daf *core.DasAddressFormat, req *core.ChainTypeAddress, apiResp *api_code.ApiResp) *core.DasAddressHex {
	if req.Type != "blockchain" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("type [%s] is invalid", req.Type))
		return nil
	}
	if req.KeyInfo.Key == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "key is invalid")
		return nil
	}
	dasChainType := common.FormatCoinTypeToDasChainType(req.KeyInfo.CoinType)
	if dasChainType == -1 {
		dasChainType = common.FormatChainIdToDasChainType(config.Cfg.Server.Net, req.KeyInfo.ChainId)
	}
	if dasChainType == -1 {
		if !strings.HasPrefix(req.KeyInfo.Key, "0x") {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("coin_type [%s] and chain_id [%s] is invalid", req.KeyInfo.CoinType, req.KeyInfo.ChainId))
			return nil
		}

		ok, err := regexp.MatchString("^0x[0-9a-fA-F]{40}$", req.KeyInfo.Key)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
			return nil
		}

		if ok {
			dasChainType = common.ChainTypeEth
		} else {
			ok, err = regexp.MatchString("^0x[0-9a-fA-F]{64}$", req.KeyInfo.Key)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
				return nil
			}
			if !ok {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "key is invalid")
				return nil
			}
			dasChainType = common.ChainTypeMixin
		}
	}
	addrHex, err := daf.NormalToHex(core.DasAddressNormal{
		ChainType:     dasChainType,
		AddressNormal: req.KeyInfo.Key,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, err.Error())
		return nil
	}
	if addrHex.DasAlgorithmId == common.DasAlgorithmIdEth712 {
		addrHex.DasAlgorithmId = common.DasAlgorithmIdEth
	}
	return &addrHex
}
