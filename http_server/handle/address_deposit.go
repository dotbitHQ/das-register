package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqAddressDeposit struct {
	AlgorithmId common.DasAlgorithmId `json:"algorithm_id"`
	Address     string                `json:"address"`
}

type RespAddressDeposit struct {
	CkbAddress string `json:"ckb_address"`
}

func (h *HttpHandle) AddressDeposit(ctx *gin.Context) {
	var (
		funcName = "AddressDeposit"
		clientIp = GetClientIp(ctx)
		req      ReqAddressDeposit
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

	if err = h.doAddressDeposit(&req, &apiResp); err != nil {
		log.Error("doAddressDeposit err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAddressDeposit(req *ReqAddressDeposit, apiResp *api_code.ApiResp) error {
	var resp RespAddressDeposit

	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.AlgorithmId.ToChainType(),
		AddressNormal: req.Address,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "HexToArgs err")
		return fmt.Errorf("NormalToHex err: %s", err.Error())
	}
	addressHex.DasAlgorithmId = req.AlgorithmId

	lockScript, _, err := h.dasCore.Daf().HexToScript(addressHex)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("HexToScript err: %s", err.Error())
	}
	log.Info("doAddressDeposit:", req.Address, common.Bytes2Hex(lockScript.Args))

	if config.Cfg.Server.Net == common.DasNetTypeMainNet {
		addr, err := address.ConvertScriptToAddress(address.Mainnet, lockScript)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("ConvertScriptToAddress err: %s", err.Error())
		}
		resp.CkbAddress = addr
	} else {
		addr, err := address.ConvertScriptToAddress(address.Testnet, lockScript)
		if err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
			return fmt.Errorf("ConvertScriptToAddress err: %s", err.Error())
		}
		resp.CkbAddress = addr
	}

	apiResp.ApiRespOK(resp)
	return nil
}
