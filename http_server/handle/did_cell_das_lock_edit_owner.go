package handle

import (
	"bytes"
	"context"
	"das_register_server/config"
	"das_register_server/internal"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqDidCellDasLockEditOwner struct {
	core.ChainTypeAddress
	Account  string `json:"account"`
	RawParam struct {
		ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
		ReceiverAddress  string          `json:"receiver_address"`
	} `json:"raw_param"`
}

type RespDidCellDasLockEditOwner struct {
	SignInfo
}

func (h *HttpHandle) RpcDidCellDasLockEditOwner(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellDasLockEditOwner
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doDidCellDasLockEditOwner(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doDidCellDasLockEditOwner err:", err.Error())
	}
}

func (h *HttpHandle) DidCellDasLockEditOwner(ctx *gin.Context) {
	var (
		funcName = "DidCellDasLockEditOwner"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellDasLockEditOwner
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doDidCellDasLockEditOwner(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doDidCellDasLockEditOwner err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellDasLockEditOwner(ctx context.Context, req *ReqDidCellDasLockEditOwner, apiResp *http_api.ApiResp) error {
	var resp RespDidCellDasLockEditOwner

	req.Account = strings.ToLower(req.Account)
	if req.KeyInfo.CoinType != common.CoinTypeCKB {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}
	addrFrom, err := address.Parse(req.KeyInfo.Key)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}
	formHex, _, err := h.dasCore.Daf().ScriptToHex(addrFrom.Script)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}
	switch formHex.DasAlgorithmId {
	case common.DasAlgorithmIdEth, common.DasAlgorithmIdTron,
		common.DasAlgorithmIdEth712, common.DasAlgorithmIdWebauthn:
	default:
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}

	//addrHexFrom, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	//if err != nil {
	//	apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address is invalid")
	//	return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	//}
	//
	//switch addrHexFrom.DasAlgorithmId {
	//case common.DasAlgorithmIdBitcoin:
	//	apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "address invalid")
	//	return nil
	//}
	toCTA := core.ChainTypeAddress{
		Type: "blockchain",
		KeyInfo: core.KeyInfo{
			CoinType: req.RawParam.ReceiverCoinType,
			Key:      req.RawParam.ReceiverAddress,
		},
	}
	addrHexTo, err := toCTA.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeInvalidTargetAddress, "receiver address is invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	if strings.EqualFold(req.KeyInfo.Key, req.RawParam.ReceiverAddress) {
		apiResp.ApiRespErr(http_api.ApiCodeSameLock, "same address")
		return nil
	}
	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}
	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(http_api.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	var didCellOutPoint *types.OutPoint
	var editOwnerLock, normalCellScript *types.Script

	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get account info")
		return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
	} else if acc.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.IsExpired() {
		apiResp.ApiRespErr(http_api.ApiCodeAccountIsExpired, "account expired")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "not support sub account")
		return nil
	}
	if acc.Status != tables.AccountStatusOnUpgrade {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "account is not a dob")
		return nil
	}
	didAccount, err := h.dbDao.GetDidAccountByAccountIdWithoutArgs(accountId)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did cell info")
		return fmt.Errorf("GetDidAccountByAccountId err: %s", err.Error())
	} else if didAccount.Id == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "did cell not exist")
		return nil
	} else if bytes.Compare(common.Hex2Bytes(didAccount.Args), addrFrom.Script.Args) != 0 {
		apiResp.ApiRespErr(http_api.ApiCodeNoAccountPermissions, "transfer account permission denied")
		return nil
	}
	didCellOutPoint = didAccount.GetOutpoint()

	// owner check
	if addrHexTo.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "DasAlgorithmId address invalid")
		return nil
	}
	editOwnerLock = addrHexTo.ParsedAddress.Script

	parseSvrAddr, err := address.Parse(config.Cfg.Server.PayServerAddress)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	normalCellScript = parseSvrAddr.Script
	txParams, err := txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
		DasCore:             h.dasCore,
		DasCache:            h.dasCache,
		Action:              common.DidCellActionEditOwner,
		DidCellOutPoint:     didCellOutPoint,
		AccountCellOutPoint: nil,
		EditOwnerLock:       editOwnerLock,
		NormalCellScript:    normalCellScript,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build tx")
		return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
	}

	reqBuild := reqBuildTx{
		OrderId:    "",
		Action:     common.DasActionTransferAccount,
		ChainType:  formHex.ChainType,
		Address:    formHex.AddressHex,
		Account:    req.Account,
		EvmChainId: req.GetChainId(config.Cfg.Server.Net),
	}
	if _, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
		checkBuildTxErr(err, apiResp)
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
		//if acc.Status == tables.AccountStatusOnUpgrade {
		//	resp.SignInfo.CKBTx = didCellTx
		//}
	}
	log.Info("doDidCellDasLockEditOwner:", toolib.JsonString(&resp))

	apiResp.ApiRespOK(resp)
	return nil
}
