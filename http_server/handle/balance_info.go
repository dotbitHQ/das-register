package handle

import (
	"context"
	"das_register_server/config"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqBalanceInfo struct {
	core.ChainTypeAddress
	ChainType       common.ChainType `json:"chain_type"`
	Address         string           `json:"address"`
	TransferAddress string           `json:"transfer_address"`
}

type RespBalanceInfo struct {
	TransferAddressAmount uint64 `json:"transfer_address_amount"`
	DasLockAmount         uint64 `json:"das_lock_amount"`
}

func (h *HttpHandle) RpcBalanceInfo(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqBalanceInfo
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

	if err = h.doBalanceInfo(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doBalanceInfo err:", err.Error())
	}
}

func (h *HttpHandle) BalanceInfo(ctx *gin.Context) {
	var (
		funcName = "BalanceInfo"
		clientIp = GetClientIp(ctx)
		req      ReqBalanceInfo
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

	if err = h.doBalanceInfo(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doBalanceInfo err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalanceInfo(ctx context.Context, req *ReqBalanceInfo, apiResp *api_code.ApiResp) error {
	//addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	//if err != nil {
	//	apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
	//	return err
	//}

	addressHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex
	var resp RespBalanceInfo

	if req.TransferAddress != "" {
		parseAddr, err := address.Parse(req.TransferAddress)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address.Parse err")
			return fmt.Errorf("address.Parse err: %s", err.Error())
		}
		searchKey := &indexer.SearchKey{
			Script:     parseAddr.Script,
			ScriptType: indexer.ScriptTypeLock,
			Filter: &indexer.CellsFilter{
				OutputDataLenRange: &[2]uint64{0, 1},
			},
		}
		res, err := h.dasCore.Client().GetCellsCapacity(h.ctx, searchKey)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "get capacity err")
			return fmt.Errorf("GetCellsCapacity err: %s", err.Error())
		}
		resp.TransferAddressAmount = res.Capacity
	}
	// not 712
	if req.ChainType == common.ChainTypeEth {
		dasLockScript, _, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
			DasAlgorithmId: req.ChainType.ToDasAlgorithmId(false),
			AddressHex:     req.Address,
			IsMulti:        false,
			ChainType:      req.ChainType,
		})
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "get das lock err")
			return fmt.Errorf("HexToScript not 712 err: %s", err.Error())
		}
		_, dasLockAmount, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
			DasCache:          nil,
			LockScript:        dasLockScript,
			CapacityNeed:      0,
			CapacityForChange: 0,
			SearchOrder:       indexer.SearchOrderDesc,
		})
		if err != nil {
			*apiResp = api_code.ApiRespErr(api_code.ApiCodeError500, "get das balance err")
			return fmt.Errorf("GetBalanceCells not 712 err: %s", err.Error())
		}
		resp.TransferAddressAmount += dasLockAmount
	}

	// 712
	dasLockScript, _, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get das lock err")
		return fmt.Errorf("FormatAddressToDasLockScript err: %s", err.Error())
	}
	_, dasLockAmount, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        dasLockScript,
		CapacityNeed:      0,
		CapacityForChange: 0,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		*apiResp = api_code.ApiRespErr(api_code.ApiCodeError500, "get 712 das balance err")
		return fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	resp.DasLockAmount = dasLockAmount

	apiResp.ApiRespOK(resp)
	return nil
}
