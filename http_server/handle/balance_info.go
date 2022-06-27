package handle

import (
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqBalanceInfo struct {
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

	if err = h.doBalanceInfo(&req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doBalanceInfo(&req, &apiResp); err != nil {
		log.Error("doBalanceInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doBalanceInfo(req *ReqBalanceInfo, apiResp *api_code.ApiResp) error {
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
		dasLockScript, dasTypeScript, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
			DasAlgorithmId: req.ChainType.ToDasAlgorithmId(false),
			AddressHex:     req.Address,
			IsMulti:        false,
			ChainType:      req.ChainType,
		})
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "get das lock err")
			return fmt.Errorf("HexToScript not 712 err: %s", err.Error())
		}
		_, dasLockAmount, err := core.GetSatisfiedCapacityLiveCell(h.dasCore.Client(), nil, dasLockScript, dasTypeScript, 0, 0)
		if err != nil {
			*apiResp = api_code.ApiRespErr(api_code.ApiCodeError500, "get das balance err")
			return fmt.Errorf("GetSatisfiedCapacityLiveCell not 712 err: %s", err.Error())
		}
		resp.TransferAddressAmount += dasLockAmount
	}

	// 712
	dasLockScript, dasTypeScript, err := h.dasCore.Daf().HexToScript(core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "get das lock err")
		return fmt.Errorf("FormatAddressToDasLockScript err: %s", err.Error())
	}
	_, dasLockAmount, err := core.GetSatisfiedCapacityLiveCell(h.dasCore.Client(), nil, dasLockScript, dasTypeScript, 0, 0)
	if err != nil {
		*apiResp = api_code.ApiRespErr(api_code.ApiCodeError500, "get 712 das balance err")
		return fmt.Errorf("GetSatisfiedCapacityLiveCell err: %s", err.Error())
	}
	resp.DasLockAmount = dasLockAmount

	apiResp.ApiRespOK(resp)
	return nil
}
