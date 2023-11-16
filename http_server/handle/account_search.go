package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/compatible"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
	"time"
)

type ReqAccountSearch struct {
	core.ChainTypeAddress
	ChainType      common.ChainType        `json:"chain_type"`
	Address        string                  `json:"address"`
	Account        string                  `json:"account"`
	AccountCharStr []common.AccountCharSet `json:"account_char_str"`
}

type RespAccountSearch struct {
	Status            tables.SearchStatus                  `json:"status"`
	Account           string                               `json:"account"`
	AccountPrice      decimal.Decimal                      `json:"account_price"`
	BaseAmount        decimal.Decimal                      `json:"base_amount"`
	IsSelf            bool                                 `json:"is_self"`
	RegisterTxMap     map[tables.RegisterStatus]RegisterTx `json:"register_tx_map"`
	OpenTimestamp     int64                                `json:"open_timestamp"`
	PremiumPercentage decimal.Decimal                      `json:"premium_percentage"`
	PremiumBase       decimal.Decimal                      `json:"premium_base"`
}

type RegisterTx struct {
	ChainType common.ChainType  `json:"chain_id"`
	TokenId   tables.PayTokenId `json:"token_id"`
	Hash      string            `json:"hash"`
}

func (h *HttpHandle) RpcAccountSearch(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountSearch
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

	if err = h.doAccountSearch(&req[0], apiResp); err != nil {
		log.Error("doAccountSearch err:", err.Error())
	}
}

func (h *HttpHandle) AccountSearch(ctx *gin.Context) {
	var (
		funcName = "AccountSearch"
		clientIp = GetClientIp(ctx)
		req      ReqAccountSearch
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

	if err = h.doAccountSearch(&req, &apiResp); err != nil {
		log.Error("doAccountSearch err:", err.Error(), funcName, clientIp, ctx)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountSearch(req *ReqAccountSearch, apiResp *api_code.ApiResp) error {
	var resp RespAccountSearch
	resp.RegisterTxMap = make(map[tables.RegisterStatus]RegisterTx)
	resp.PremiumPercentage = config.Cfg.Stripe.PremiumPercentage
	resp.PremiumBase = config.Cfg.Stripe.PremiumBase

	if req.ChainType == common.ChainTypeCkb || (req.Address == "" && req.KeyInfo.Key == "") {

	} else {
		addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
			return err
		}
		req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex
	}

	resp.Account = req.Account

	// check sub account
	isSubAccount := false
	resp.Status, resp.IsSelf, isSubAccount = h.checkSubAccount(req, apiResp)
	if isSubAccount {
		if apiResp.ErrNo != api_code.ApiCodeSuccess {
			return nil
		}
		apiResp.ApiRespOK(resp)
		return nil
	}

	// account char set check
	h.checkAccountCharSet(req, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	}

	confirmProposalHash := ""
	confirmProposalHash, resp.Status, resp.IsSelf, resp.OpenTimestamp = h.checkAccountBase(req, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if resp.Status != tables.SearchStatusRegisterAble && !resp.IsSelf {
		apiResp.ApiRespOK(resp)
		return nil
	}
	// account price
	argsStr := ""
	if req.Address == "" {

	} else {
		hexAddress := core.DasAddressHex{
			DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
			AddressHex:     req.Address,
			IsMulti:        false,
			ChainType:      req.ChainType,
		}
		args, err := h.dasCore.Daf().HexToArgs(hexAddress, hexAddress)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
			return fmt.Errorf("HexToArgs err: %s", err.Error())
		}
		argsStr = common.Bytes2Hex(args)
	}

	accLen := uint8(len(req.AccountCharStr))
	if tables.EndWithDotBitChar(req.AccountCharStr) {
		accLen -= 4
	}
	baseAmount, accountPrice, err := h.getAccountPrice(accLen, argsStr, req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	resp.BaseAmount, resp.AccountPrice = baseAmount, accountPrice
	// address order
	status, registerTxMap := h.checkAddressOrder(req, apiResp, true)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if status != tables.SearchStatusRegisterAble {
		if resp.Status == tables.SearchStatusRegisterAble {
			resp.Status = status
		}
		resp.RegisterTxMap = registerTxMap
		if _, ok := resp.RegisterTxMap[tables.RegisterStatusConfirmProposal]; !ok && confirmProposalHash != "" {
			resp.RegisterTxMap[tables.RegisterStatusConfirmProposal] = RegisterTx{
				ChainType: common.ChainTypeCkb,
				Hash:      confirmProposalHash,
			}
		}
		resp.IsSelf = true
		apiResp.ApiRespOK(resp)
		return nil
	}
	// other register
	status = h.checkOtherAddressOrder(req, apiResp)
	if apiResp.ErrNo != api_code.ApiCodeSuccess {
		return nil
	} else if status != tables.SearchStatusRegisterAble {
		if resp.Status == tables.SearchStatusRegisterAble {
			resp.Status = status
		}
		apiResp.ApiRespOK(resp)
		return nil
	}

	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) checkAccountCharSet(req *ReqAccountSearch, apiResp *api_code.ApiResp) {
	if !strings.HasSuffix(req.Account, common.DasAccountSuffix) {
		apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "not has suffix .bit")
		return
	} else if len(req.Account) <= 4 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "Invalid account name")
		return
	}

	accountName := strings.TrimSuffix(req.Account, common.DasAccountSuffix)
	if strings.Contains(accountName, ".") {
		apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "char invalid.")
		return
	}
	var accountCharStr string
	for _, v := range req.AccountCharStr {
		if v.Char == "" {
			apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "char invalid nil")
			return
		}
		switch v.CharSetName {
		case common.AccountCharTypeEmoji:
			if _, ok := common.CharSetTypeEmojiMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "emoji char invalid")
				return
			}
		case common.AccountCharTypeDigit:
			if _, ok := common.CharSetTypeDigitMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "digit char invalid")
				return
			}
		case common.AccountCharTypeEn:
			if _, ok := common.CharSetTypeEnMap[v.Char]; v.Char != "." && !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "en char invalid")
				return
			}
		case common.AccountCharTypeJa:
			if _, ok := common.CharSetTypeJaMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "ja char invalid")
				return
			}
		case common.AccountCharTypeRu:
			if _, ok := common.CharSetTypeRuMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "ru char invalid")
				return
			}
		case common.AccountCharTypeTr:
			if _, ok := common.CharSetTypeTrMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "tr char invalid")
				return
			}
		case common.AccountCharTypeVi:
			if _, ok := common.CharSetTypeViMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "vi char invalid")
				return
			}
		case common.AccountCharTypeTh:
			if _, ok := common.CharSetTypeThMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "th char invalid")
				return
			}
		case common.AccountCharTypeKo:
			if _, ok := common.CharSetTypeKoMap[v.Char]; !ok {
				apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "ko char invalid")
				return
			}
		default:
			apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, fmt.Sprintf("%d char invalid", v.CharSetName))
			return
		}
		accountCharStr += v.Char
	}

	if !strings.HasSuffix(accountCharStr, common.DasAccountSuffix) {
		accountCharStr += common.DasAccountSuffix
	}
	if !strings.EqualFold(req.Account, accountCharStr) {
		apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, fmt.Sprintf("diff account chars[%s]!=[%s]", accountCharStr, req.Account))
		return
	}

	if isDiff := common.CheckAccountCharTypeDiff(req.AccountCharStr); isDiff {
		apiResp.ApiRespErr(api_code.ApiCodeAccountContainsInvalidChar, "diff char type")
		return
	}
	return
}

var OpenCharTypeMap = map[common.AccountCharType]struct{}{
	common.AccountCharTypeEmoji: {},
	common.AccountCharTypeDigit: {},
	common.AccountCharTypeKo:    {},
	common.AccountCharTypeTh:    {},
	//common.AccountCharTypeTr:    {},
	//common.AccountCharTypeVi:    {},
}

func (h *HttpHandle) checkAccountBase(req *ReqAccountSearch, apiResp *api_code.ApiResp) (confirmProposalHash string, status tables.SearchStatus, isSelf bool, openTs int64) {
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		log.Error("GetAccountInfoByAccountId err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account fail")
		return
	} else if acc.Id > 0 {
		status = acc.FormatAccountStatus()
		confirmProposalHash = acc.ConfirmProposalHash
		if req.ChainType == acc.OwnerChainType && strings.EqualFold(req.Address, acc.Owner) {
			isSelf = true
		}
		return
	} else {
		accountName := strings.ToLower(strings.TrimSuffix(req.Account, common.DasAccountSuffix))
		accountName = common.Bytes2Hex(common.Blake2b([]byte(accountName))[:20])
		// unavailable
		if _, ok := h.mapUnAvailableAccounts[accountName]; ok {
			status = tables.SearchStatusUnAvailableAccount
			return
		}
		// reserved
		if _, ok := h.mapReservedAccounts[accountName]; ok {
			status = tables.SearchStatusReservedAccount
			return
		}
		// accLen
		//accLen := common.GetAccountLength(req.Account)
		accLen := uint8(len(req.AccountCharStr))
		if tables.EndWithDotBitChar(req.AccountCharStr) {
			accLen -= 4
		}
		log.Info("account len:", accLen, req.Account)
		if accLen < config.Cfg.Das.AccountMinLength || accLen > config.Cfg.Das.AccountMaxLength {
			apiResp.ApiRespErr(api_code.ApiCodeAccountLenInvalid, fmt.Sprintf("account len err:%d [%s]", accLen, accountName))
			return
		} else if accLen >= config.Cfg.Das.OpenAccountMinLength && accLen <= config.Cfg.Das.OpenAccountMaxLength {
			// check time cell
			tc, err := h.dasCore.GetTimeCell()
			if err != nil {
				apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("get time cell err: %s", err.Error()))
				return
			}
			tcTimestamp := tc.Timestamp()
			openTimestamp := int64(1666094400)
			if config.Cfg.Server.Net != common.DasNetTypeMainNet {
				//openTimestamp = 1666094400
				openTimestamp = 1665712800
			}
			// check dao char type
			isSameDaoCharType := true
			for i, v := range req.AccountCharStr {
				if v.Char == "." {
					break
				}
				if i == 0 {
					continue
				}
				if _, ok := OpenCharTypeMap[req.AccountCharStr[i].CharSetName]; !ok {
					isSameDaoCharType = false
					break
				}
				if req.AccountCharStr[i].CharSetName != req.AccountCharStr[i-1].CharSetName {
					isSameDaoCharType = false
					break
				}
			}
			if tcTimestamp >= openTimestamp && isSameDaoCharType {
				return
			}

			configRelease, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsRelease)
			if err != nil {
				log.Error("GetDasConfigCellInfo err:", err.Error())
				apiResp.ApiRespErr(api_code.ApiCodeError500, "search config release fail")
				return
			}
			luckyNumber, _ := configRelease.LuckyNumber()
			log.Info("config release lucky number: ", luckyNumber)
			if resNum, _ := Blake256AndFourBytesBigEndian([]byte(req.Account)); resNum > luckyNumber {
				status = tables.SearchStatusRegisterNotOpen
				if isSameDaoCharType {
					openTs = openTimestamp
				}
				return
			}
		}
	}
	return
}

func (h *HttpHandle) checkAddressOrder(req *ReqAccountSearch, apiResp *api_code.ApiResp, isGetOrderTx bool) (status tables.SearchStatus, mapTx map[tables.RegisterStatus]RegisterTx) {
	mapTx = make(map[tables.RegisterStatus]RegisterTx)
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))

	var txList []tables.TableDasOrderTxInfo
	var order tables.TableDasOrderInfo
	order, err := h.dbDao.GetLatestRegisterOrderByAddress(req.ChainType, req.Address, accountId)
	if err != nil {
		log.Error("GetLatestRegisterOrderByAddress err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order fail")
		return
	}
	timeCheck := time.Now().Add(-time.Hour*24*365).UnixNano() / 1e6
	log.Info("checkAddressOrder:", timeCheck, order.Timestamp)
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		log.Error("GetAccountInfoByAccountId err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account fail")
		return
	}
	if acc.Id == 0 && timeCheck > order.Timestamp {
		status = tables.SearchStatusRegisterAble
		return
	}

	if (order.Id > 0 && order.OrderStatus == tables.OrderStatusDefault) || (order.Id > 0 && order.RegisterStatus == tables.RegisterStatusRegistered) {
		status = tables.FormatRegisterStatusToSearchStatus(order.RegisterStatus)
		if !isGetOrderTx {
			return
		}
		if order.OrderType == tables.OrderTypeSelf {
			if order.RegisterStatus == tables.RegisterStatusRegistered && order.CrossCoinType != "" {
				status = tables.SearchStatusOnCross
			}
			payInfo, err := h.dbDao.GetPayInfoByOrderId(order.OrderId)
			if err != nil {
				log.Error("GetPayInfoByOrderId err:", err.Error())
				apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order pay fail")
				return
			} else if payInfo.Id > 0 {
				chainType := payInfo.ChainType
				switch order.PayTokenId {
				case tables.TokenIdDas, tables.TokenIdCkb, tables.TokenIdCkbInternal:
					chainType = common.ChainTypeCkb
				case tables.TokenIdEth, tables.TokenIdBnb, tables.TokenIdMatic:
					chainType = common.ChainTypeEth
				case tables.TokenIdTrx:
					chainType = common.ChainTypeTron
				}
				mapTx[tables.RegisterStatusConfirmPayment] = RegisterTx{
					ChainType: chainType,
					TokenId:   order.PayTokenId,
					Hash:      payInfo.Hash,
				}
				if order.PayTokenId == tables.TokenIdStripeUSD && payInfo.Status != tables.OrderTxStatusConfirm {
					delete(mapTx, tables.RegisterStatusConfirmPayment)
				}
			}
		}
		txList, err = h.dbDao.GetOrderTxListByOrderId(order.OrderId)
		if err != nil {
			log.Error("GetOrderTxListByOrderId err:", err.Error())
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order tx fail")
			return
		}
		for _, v := range txList {
			switch v.Action {
			case tables.TxActionApplyRegister:
				mapTx[tables.RegisterStatusApplyRegister] = RegisterTx{
					ChainType: common.ChainTypeCkb,
					Hash:      v.Hash,
				}
			case tables.TxActionPreRegister:
				mapTx[tables.RegisterStatusPreRegister] = RegisterTx{
					ChainType: common.ChainTypeCkb,
					Hash:      v.Hash,
				}
			case tables.TxActionPropose:
				mapTx[tables.RegisterStatusProposal] = RegisterTx{
					ChainType: common.ChainTypeCkb,
					Hash:      v.Hash,
				}
			case tables.TxActionConfirmProposal:
				mapTx[tables.RegisterStatusConfirmProposal] = RegisterTx{
					ChainType: common.ChainTypeCkb,
					Hash:      v.Hash,
				}
			}
		}
	}
	return
}

func (h *HttpHandle) checkOtherAddressOrder(req *ReqAccountSearch, apiResp *api_code.ApiResp) (status tables.SearchStatus) {
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	order, err := h.dbDao.GetLatestRegisterOrderByLatest(accountId)
	if err != nil {
		log.Error("GetLatestRegisterOrderByLatest err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order fail")
		return
	} else if order.Id > 0 {
		status = tables.FormatRegisterStatusToSearchStatus(order.RegisterStatus)
	}
	return
}

func (h *HttpHandle) checkSubAccount(req *ReqAccountSearch, apiResp *api_code.ApiResp) (status tables.SearchStatus, isSelf, isSubAccount bool) {
	count := strings.Count(req.Account, ".")
	if count > 1 {
		isSubAccount = true
		accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
		acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
		if err != nil {
			log.Error("GetAccountInfoByAccountId err:", err.Error())
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account fail")
			return
		} else if acc.Id > 0 {
			status = acc.FormatAccountStatus()
			if req.ChainType == acc.OwnerChainType && strings.EqualFold(req.Address, acc.Owner) {
				isSelf = true
			}
			return
		} else {
			status = tables.SearchStatusSubAccountUnRegister
		}
	}
	return
}
