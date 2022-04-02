package handle

import (
	"bytes"
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/minio/blake2b-simd"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type ReqAccountDetail struct {
	Account string `json:"account"`
}

type RespAccountDetail struct {
	Account              string                  `json:"account"`
	Owner                string                  `json:"owner"`
	OwnerChainType       common.ChainType        `json:"owner_chain_type"`
	Manager              string                  `json:"manager"`
	ManagerChainType     common.ChainType        `json:"manager_chain_type"`
	RegisteredAt         int64                   `json:"registered_at"`
	ExpiredAt            int64                   `json:"expired_at"`
	Status               tables.SearchStatus     `json:"status"`
	AccountPrice         decimal.Decimal         `json:"account_price"`
	BaseAmount           decimal.Decimal         `json:"base_amount"`
	ConfirmProposalHash  string                  `json:"confirm_proposal_hash"`
	EnableSubAccount     tables.EnableSubAccount `json:"enable_sub_account"`
	RenewSubAccountPrice uint64                  `json:"renew_sub_account_price"`
	Nonce                uint64                  `json:"nonce"`
}

func (h *HttpHandle) RpcAccountDetail(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqAccountDetail
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

	if err = h.doAccountDetail(&req[0], apiResp); err != nil {
		log.Error("doAccountDetail err:", err.Error())
	}
}

func (h *HttpHandle) AccountDetail(ctx *gin.Context) {
	var (
		funcName = "AccountDetail"
		clientIp = GetClientIp(ctx)
		req      ReqAccountDetail
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

	if err = h.doAccountDetail(&req, &apiResp); err != nil {
		log.Error("doAccountDetail err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) getAccountPrice(args, account string, isRenew bool) (baseAmount, accountPrice decimal.Decimal, err error) {
	builder, err := h.dasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsPrice, common.ConfigCellTypeArgsAccount)
	if err != nil {
		err = fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
		return
	}
	accLen := common.GetAccountLength(account)
	newPrice, renewPrice, err := builder.AccountPrice(accLen)
	if err != nil {
		err = fmt.Errorf("AccountPrice err: %s", err.Error())
		return
	}

	quoteCell, err := h.dasCore.GetQuoteCell()
	if err != nil {
		err = fmt.Errorf("GetQuoteCell err: %s", err.Error())
		return
	}
	quote := quoteCell.Quote()

	if args == "" {
		args = "0x03"
	}
	basicCapacity, err := builder.BasicCapacityFromOwnerDasAlgorithmId(args)
	if err != nil {
		err = fmt.Errorf("BasicCapacity err: %s", err.Error())
		return
	}
	preparedFeeCapacity, err := builder.PreparedFeeCapacity()
	if err != nil {
		err = fmt.Errorf("PreparedFeeCapacity err: %s", err.Error())
		return
	}
	log.Info("BasicCapacity:", basicCapacity, "PreparedFeeCapacity:", preparedFeeCapacity, "Quote:", quote, "Price:", newPrice, renewPrice)

	basicCapacity = basicCapacity/common.OneCkb + uint64(len([]byte(account))) + preparedFeeCapacity/common.OneCkb
	baseAmount, _ = decimal.NewFromString(fmt.Sprintf("%d", basicCapacity))
	decQuote, _ := decimal.NewFromString(fmt.Sprintf("%d", quote))
	decUsdRateBase := decimal.NewFromInt(common.UsdRateBase)
	baseAmount = baseAmount.Mul(decQuote).DivRound(decUsdRateBase, 2)

	if isRenew {
		accountPrice, _ = decimal.NewFromString(fmt.Sprintf("%d", renewPrice))
		accountPrice = accountPrice.DivRound(decUsdRateBase, 2)
	} else {
		accountPrice, _ = decimal.NewFromString(fmt.Sprintf("%d", newPrice))
		accountPrice = accountPrice.DivRound(decUsdRateBase, 2)
	}
	return
}

func (h *HttpHandle) doAccountDetail(req *ReqAccountDetail, apiResp *api_code.ApiResp) error {
	var resp RespAccountDetail
	resp.Account = req.Account
	resp.Status = tables.SearchStatusRegisterAble

	// check sub account
	count := strings.Count(req.Account, ".")
	if count == 1 {
		// price
		var err error
		resp.BaseAmount, resp.AccountPrice, err = h.getAccountPrice("", req.Account, true)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "get account price err")
			return fmt.Errorf("getAccountPrice err: %s", err.Error())
		}
	}

	// acc
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}

	if acc.Id > 0 {
		resp.Status = acc.FormatAccountStatus()
		resp.ExpiredAt = int64(acc.ExpiredAt) * 1e3
		resp.RegisteredAt = int64(acc.RegisteredAt) * 1e3
		resp.OwnerChainType = acc.OwnerChainType
		resp.Owner = core.FormatHexAddressToNormal(acc.OwnerChainType, acc.Owner)
		resp.ManagerChainType = acc.ManagerChainType
		resp.Manager = core.FormatHexAddressToNormal(acc.ManagerChainType, acc.Manager)
		resp.ConfirmProposalHash = acc.ConfirmProposalHash
		resp.EnableSubAccount = acc.EnableSubAccount
		resp.RenewSubAccountPrice = acc.RenewSubAccountPrice
		resp.Nonce = acc.Nonce
		apiResp.ApiRespOK(resp)
		return nil
	}

	if count == 1 {
		// reserve account
		accountName := strings.ToLower(strings.TrimSuffix(req.Account, common.DasAccountSuffix))
		accountName = common.Bytes2Hex(common.Blake2b([]byte(accountName))[:20])

		if _, ok := h.mapReservedAccounts[accountName]; ok {
			resp.Status = tables.SearchStatusReservedAccount
			apiResp.ApiRespOK(resp)
			return nil
		}

		// unavailable account
		if _, ok := h.mapUnAvailableAccounts[accountName]; ok {
			resp.Status = tables.SearchStatusUnAvailableAccount
			apiResp.ApiRespOK(resp)
			return nil
		}
	}

	// account not exist
	apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, fmt.Sprintf("account [%s] not exist", req.Account))

	// not open
	//accLen := common.GetAccountLength(req.Account)
	//if accLen < config.Cfg.Das.AccountMinLength || accLen > config.Cfg.Das.AccountMaxLength {
	//	resp.Status = tables.SearchStatusRegisterNotOpen
	//	apiResp.ApiRespOK(resp)
	//	return nil
	//} else if accLen >= config.Cfg.Das.OpenAccountMinLength && accLen <= config.Cfg.Das.OpenAccountMaxLength {
	//	configRelease, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsRelease)
	//	if err != nil {
	//		resp.Status = tables.SearchStatusRegisterNotOpen
	//		apiResp.ApiRespOK(resp)
	//		return nil
	//	}
	//	luckyNumber, _ := configRelease.LuckyNumber()
	//	if resNum, _ := Blake256AndFourBytesBigEndian([]byte(req.Account)); resNum > luckyNumber {
	//		resp.Status = tables.SearchStatusRegisterNotOpen
	//		apiResp.ApiRespOK(resp)
	//		return nil
	//	}
	//}
	//apiResp.ApiRespOK(resp)
	return nil
}

func Blake256AndFourBytesBigEndian(data []byte) (uint32, error) {
	bys, err := Blake256(data)
	if err != nil {
		return 0, err
	}
	bytesBuffer := bytes.NewBuffer(bys[0:4])
	var res uint32
	if err = binary.Read(bytesBuffer, binary.BigEndian, &res); err != nil {
		return 0, err
	}
	return res, nil
}

func Blake256(data []byte) ([]byte, error) {
	tmpConfig := &blake2b.Config{
		Size:   32,
		Person: []byte("2021-07-22 12:00"),
	}
	hash, err := blake2b.New(tmpConfig)
	if err != nil {
		return nil, err
	}
	hash.Write(data)
	return hash.Sum(nil), nil
}
