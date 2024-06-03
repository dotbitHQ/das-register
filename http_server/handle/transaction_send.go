package handle

import (
	"das_register_server/tables"
	"das_register_server/txtool"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
	"time"
)

type ReqTransactionSend struct {
	SignInfo
}

type RespTransactionSend struct {
	Hash string `json:"hash"`
}

func (h *HttpHandle) RpcTransactionSend(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqTransactionSend
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doTransactionSend(&req[0], apiResp); err != nil {
		log.Error("doVersion err:", err.Error())
	}
}

func (h *HttpHandle) TransactionSend(ctx *gin.Context) {
	var (
		funcName = "TransactionSend"
		clientIp = GetClientIp(ctx)
		req      ReqTransactionSend
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doTransactionSend(&req, &apiResp); err != nil {
		log.Error("doTransactionSend err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTransactionSend(req *ReqTransactionSend, apiResp *api_code.ApiResp) error {
	var resp RespTransactionSend

	var sic SignInfoCache
	// get tx by cache
	if txStr, err := h.rc.GetSignTxCache(req.SignKey); err != nil {
		if err == redis.Nil {
			apiResp.ApiRespErr(api_code.ApiCodeTxExpired, "tx expired err")
		} else {
			apiResp.ApiRespErr(api_code.ApiCodeCacheError, "cache err")
		}
		return fmt.Errorf("GetSignTxCache err: %s", err.Error())
	} else if err = json.Unmarshal([]byte(txStr), &sic); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "json.Unmarshal err")
		return fmt.Errorf("json.Unmarshal err: %s", err.Error())
	}

	hasWebAuthn := false
	for _, v := range req.SignList {
		if v.SignType == common.DasAlgorithmIdWebauthn && v.SignMsg != "" {
			hasWebAuthn = true
			break
		}
	}

	if hasWebAuthn {
		loginAddrHex := core.DasAddressHex{
			DasAlgorithmId:    common.DasAlgorithmIdWebauthn,
			DasSubAlgorithmId: common.DasWebauthnSubAlgorithmIdES256,
			AddressHex:        sic.Address,
			AddressPayload:    common.Hex2Bytes(sic.Address),
			ChainType:         common.ChainTypeWebauthn,
		}

		if req.SignAddress == "" {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "SignAddress err")
			return nil
		}
		log.Info("req.signaddress: ", req.SignAddress)
		signAddressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
			ChainType:     common.ChainTypeWebauthn,
			AddressNormal: req.SignAddress,
		})
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "sign address NormalToHex err")
			return err
		}
		log.Info("loginAddrHex.AddressHex: ", loginAddrHex.AddressHex)
		log.Info("signAddressHex.AddressHex: ", signAddressHex.AddressHex)
		idx, err := h.dasCore.GetIdxOfKeylist(loginAddrHex, signAddressHex)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "GetIdxOfKeylist err: "+err.Error())
			return fmt.Errorf("GetIdxOfKeylist err: %s", err.Error())
		}
		if idx == -1 {
			apiResp.ApiRespErr(api_code.ApiCodePermissionDenied, "permission denied")
			return fmt.Errorf("permission denied")
		}

		for i, v := range req.SignList {
			if v.SignType == common.DasAlgorithmIdWebauthn && v.SignMsg != "" {
				h.dasCore.AddPkIndexForSignMsg(&req.SignList[i].SignMsg, idx)
			}
		}
	}

	// sign
	if req.CKBTx != "" {
		userTx, err := rpc.TransactionFromString(req.CKBTx)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("rpc.TransactionFromString err: %s", err.Error()))
			return fmt.Errorf("rpc.TransactionFromString err: %s", err.Error())
		}
		cacheTxHash, err := sic.BuilderTx.Transaction.ComputeHash()
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "ComputeHash err")
			return fmt.Errorf("ComputeHash err: %s", err.Error())
		}
		//userTx := txbuilder.TransactionToTx(req.CKBTx)
		userTxHash, err := userTx.ComputeHash()
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "ComputeHash err")
			return fmt.Errorf("ComputeHash err: %s", err.Error())
		}
		if userTxHash.String() != cacheTxHash.String() {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "ckb tx invalid")
			return nil
		}
		sic.BuilderTx.Transaction = userTx
	}

	txBuilder := txbuilder.NewDasTxBuilderFromBase(h.txBuilderBase, sic.BuilderTx)
	if err := txBuilder.AddSignatureForTx(req.SignList); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "add signature fail")
		return fmt.Errorf("AddSignatureForTx err: %s", err.Error())
	}
	//
	if sic.Action == common.DasActionEditRecords {
		builder, err := witness.AccountCellDataBuilderFromTx(txBuilder.Transaction, common.DataTypeNew)
		if err != nil {
			log.Warn("AccountCellDataBuilderFromTx err: ", err.Error())
		} else {
			log.Info("edit records:", sic.Account, sic.ChainType, sic.Address, toolib.JsonString(builder.Records))
		}
	}

	// account cell -> did cell
	if (sic.Action == common.DasActionTransferAccount || sic.Action == common.DasActionRenewAccount) && sic.OrderId != "" {
		didCellTxCache := txtool.DidCellTxCache{
			BuilderTx: txBuilder.DasTxBuilderTransaction,
		}
		didCellTxCacheStr := toolib.JsonString(&didCellTxCache)
		if err := h.rc.SetCache(sic.OrderId, didCellTxCacheStr, time.Hour*24); err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "Failed to cache did cell tx")
			return fmt.Errorf("SetCache err: %s", err.Error())
		}
		var outpoints []string
		for _, v := range txBuilder.Transaction.Inputs {
			outpoints = append(outpoints, common.OutPointStruct2String(v.PreviousOutput))
		}
		h.dasCache.AddOutPointWithDuration(outpoints, time.Hour*12)
		apiResp.ApiRespOK(resp)
		return nil
	}

	// send tx
	if hash, err := txBuilder.SendTransaction(); err != nil {
		if strings.Contains(err.Error(), "PoolRejectedDuplicatedTransaction") ||
			strings.Contains(err.Error(), "Dead(OutPoint(") ||
			strings.Contains(err.Error(), "Unknown(OutPoint(") ||
			(strings.Contains(err.Error(), "getInputCell") && strings.Contains(err.Error(), "not live")) {
			apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, err.Error())
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		if strings.Contains(err.Error(), "-102 in the page") {
			apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "account frequency limit")
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		if strings.Contains(err.Error(), "error code 65 in") {
			apiResp.ApiRespErr(api_code.ApiCodeTooManyRecords, "too many records")
			return fmt.Errorf("SendTransaction err: %s", err.Error())
		}
		apiResp.ApiRespErr(api_code.ApiCodeError500, "send tx err:"+err.Error())
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		resp.Hash = hash.Hex()
		if sic.Address != "" {

			// operate limit
			_ = h.rc.SetApiLimit(sic.ChainType, sic.Address, sic.Action)
			_ = h.rc.SetAccountLimit(sic.Account, time.Minute*2)

			// cache tx inputs
			h.dasCache.AddCellInputByAction("", sic.BuilderTx.Transaction.Inputs)
			// pending tx
			pending := tables.TableRegisterPendingInfo{
				Account:        sic.Account,
				Action:         sic.Action,
				ChainType:      sic.ChainType,
				Address:        sic.Address,
				Capacity:       sic.Capacity,
				Outpoint:       common.OutPoint2String(hash.Hex(), 0),
				BlockTimestamp: uint64(time.Now().UnixNano() / 1e6),
			}
			if err = h.dbDao.CreatePending(&pending); err != nil {
				log.Error("CreatePending err: ", err.Error(), toolib.JsonString(pending))
			}

			if sic.Action == common.DasActionBidExpiredAccountAuction {

				auctionOrder := tables.TableAuctionOrder{
					Account:      sic.Account,
					AccountId:    common.Bytes2Hex(common.GetAccountIdByAccount(sic.Account)),
					Address:      sic.Address,
					BasicPrice:   sic.AuctionInfo.BasicPrice,
					PremiumPrice: sic.AuctionInfo.PremiumPrice,
					BidTime:      sic.AuctionInfo.BidTime,
					ChainType:    sic.ChainType,
					Outpoint:     pending.Outpoint,
				}
				auctionOrder.CreateOrderId()
				if err = h.dbDao.CreateAuctionOrder(auctionOrder); err != nil {
					log.Error("CreateAuctionOrder err: ", err.Error(), toolib.JsonString(auctionOrder))
				}
			}
		}
	}

	apiResp.ApiRespOK(resp)
	return nil
}
