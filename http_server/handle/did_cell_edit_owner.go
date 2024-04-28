package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqDidCellEditOwner struct {
	core.ChainTypeAddress
	Account  string `json:"account"`
	RawParam struct {
		ReceiverCoinType common.CoinType `json:"receiver_coin_type"`
		ReceiverAddress  string          `json:"receiver_address"`
	} `json:"raw_param"`
}

type RespDidCellEditOwner struct {
	SignInfo
}

func (h *HttpHandle) RpcDidCellEditOwner(p json.RawMessage, apiResp *http_api.ApiResp) {
	var req []ReqDidCellEditOwner
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

	if err = h.doDidCellEditOwner(&req[0], apiResp); err != nil {
		log.Error("doDidCellEditOwner err:", err.Error())
	}
}

func (h *HttpHandle) DidCellEditOwner(ctx *gin.Context) {
	var (
		funcName = "DidCellEditOwner"
		clientIp = GetClientIp(ctx)
		req      ReqDidCellEditOwner
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doDidCellEditOwner(&req, &apiResp); err != nil {
		log.Error("doDidCellEditOwner err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDidCellEditOwner(req *ReqDidCellEditOwner, apiResp *http_api.ApiResp) error {
	var resp RespDidCellEditOwner

	isFromAnyLock, fromParseAddr, err := h.isAnyLock(req.ChainTypeAddress, apiResp)
	if err != nil {
		return fmt.Errorf("isAnyLock err: %s", err.Error())
	}
	isToAnyLock, toParseAddr, err := h.isAnyLock(core.ChainTypeAddress{
		Type: "blockchain",
		KeyInfo: core.KeyInfo{
			CoinType: req.RawParam.ReceiverCoinType,
			Key:      req.RawParam.ReceiverAddress,
		},
	}, apiResp)
	if err != nil {
		return fmt.Errorf("isAnyLock err: %s", err.Error())
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))

	var txParams *txbuilder.BuildTransactionParams
	if isFromAnyLock && isToAnyLock {
		// did cell -> did cell
		fromArgs := common.Bytes2Hex(fromParseAddr.Script.Args)
		didAccount, err := h.dbDao.GetDidAccountByAccountId(accountId, fromArgs)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeDbError, "Failed to get did cell info")
			return fmt.Errorf("GetDidAccountByAccountId err: %s", err.Error())
		} else if didAccount.Id == 0 {
			apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "did cell not exist")
			return nil
		} else if didAccount.IsExpired() {
			apiResp.ApiRespErr(http_api.ApiCodeAccountIsExpired, "did cell expired")
			return nil
		}

		txParams, err = txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
			DasCore:             h.dasCore,
			Action:              common.DidCellActionEditOwner,
			DidCellOutPoint:     didAccount.GetOutpoint(),
			AccountCellOutPoint: nil,
			EditRecords:         nil,
			EditOwnerLock:       toParseAddr.Script,
			NormalCkbLiveCell:   nil,
			RenewYears:          0,
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build tx")
			return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
		}
	} else if !isFromAnyLock {
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
		}

		var editOwnerLock *types.Script
		var normalCkbLiveCell []*indexer.LiveCell
		if !isToAnyLock {
			// account cell -> account cell
			chainTypeAddress := core.ChainTypeAddress{
				Type: "blockchain",
				KeyInfo: core.KeyInfo{
					CoinType: req.RawParam.ReceiverCoinType,
					Key:      req.RawParam.ReceiverAddress,
				},
			}
			ownerHex, err := chainTypeAddress.FormatChainTypeAddress(config.Cfg.Server.Net, true)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "receiver address is invalid")
				return nil
			}
			editOwnerLock, _, err = h.dasCore.Daf().HexToScript(*ownerHex)
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "receiver address is invalid")
				return nil
			}
		} else {
			// account cell -> did cell
			editOwnerLock = toParseAddr.Script
			// todo normalCkbLiveCell
		}

		txParams, err = txbuilder.BuildDidCellTx(txbuilder.DidCellTxParams{
			DasCore:             h.dasCore,
			Action:              common.DidCellActionEditOwner,
			DidCellOutPoint:     nil,
			AccountCellOutPoint: acc.GetOutpoint(),
			EditRecords:         nil,
			EditOwnerLock:       editOwnerLock,
			NormalCkbLiveCell:   normalCkbLiveCell,
			RenewYears:          0,
		})
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "Failed to build tx")
			return fmt.Errorf("BuildDidCellTx err: %s", err.Error())
		}
	} else {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid")
		return nil
	}
	reqBuild := reqBuildTx{
		Action:     common.DasActionTransferAccount,
		ChainType:  0,
		Address:    req.KeyInfo.Key,
		Account:    req.Account,
		EvmChainId: req.GetChainId(config.Cfg.Server.Net),
	}
	if si, err := h.buildTx(&reqBuild, txParams); err != nil {
		checkBuildTxErr(err, apiResp)
		//apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err")
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) isAnyLock(cta core.ChainTypeAddress, apiResp *http_api.ApiResp) (bool, *address.ParsedAddress, error) {
	if cta.KeyInfo.CoinType == common.CoinTypeCKB {
		addrParse, err := address.Parse(cta.KeyInfo.Key)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address is invalid")
			return false, nil, fmt.Errorf("address.Parse err: %s", err.Error())
		}
		contractDispatch, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to get dispatch contract")
			return false, nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
		} else if !contractDispatch.IsSameTypeId(addrParse.Script.CodeHash) {
			return true, addrParse, nil
		}
	}
	return false, nil, nil
}
