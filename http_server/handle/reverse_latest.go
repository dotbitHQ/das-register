package handle

import (
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type ReqReverseLatest struct {
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
}

type RespReverseLatest struct {
	Account string `json:"account"`
	IsValid bool   `json:"is_valid"`
}

func (h *HttpHandle) RpcReverseLatest(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqReverseLatest
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

	if err = h.doReverseLatest(&req[0], apiResp); err != nil {
		log.Error("doReverseLatest err:", err.Error())
	}
}

func (h *HttpHandle) ReverseLatest(ctx *gin.Context) {
	var (
		funcName = "ReverseLatest"
		clientIp = GetClientIp(ctx)
		req      ReqReverseLatest
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

	if err = h.doReverseLatest(&req, &apiResp); err != nil {
		log.Error("doReverseLatest err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doReverseLatest(req *ReqReverseLatest, apiResp *api_code.ApiResp) error {
	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.ChainType,
		AddressNormal: req.Address,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address NormalToHex err")
		return fmt.Errorf("NormalToHex err: %s", err.Error())
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex
	var resp RespReverseLatest

	reverse, err := h.dbDao.SearchLatestReverse(req.ChainType, req.Address)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search reverse err")
			return fmt.Errorf("SearchLatestReverse err: %s", err.Error())
		}
	}
	if reverse.Id == 0 {
		apiResp.ApiRespOK(resp)
		return nil
	} else {
		resp.Account = reverse.Account
	}

	// account

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(reverse.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
			return fmt.Errorf("SearchAccount err: %s", err.Error())
		}
	}
	if acc.Id == 0 {
		apiResp.ApiRespOK(resp)
		return nil
	} else {
		if strings.EqualFold(req.Address, acc.Owner) {
			resp.IsValid = true
			apiResp.ApiRespOK(resp)
			return nil
		}
		if strings.EqualFold(req.Address, acc.Manager) {
			resp.IsValid = true
			apiResp.ApiRespOK(resp)
			return nil
		}
	}

	// records

	record, err := h.dbDao.SearchAccountReverseRecords(acc.Account, req.Address)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			*apiResp = api_code.ApiRespErr(api_code.ApiCodeDbError, "search account err")
			return fmt.Errorf("SearchAccount err: %s", err.Error())
		}
	}
	if record.Id > 0 {
		resp.IsValid = true
	}

	apiResp.ApiRespOK(resp)
	return nil
}
