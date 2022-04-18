package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"regexp"
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

	chainType := common.ChainTypeEth
	is712 := false
	switch req.AlgorithmId {
	case common.DasAlgorithmIdEth:
		if ok, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", req.Address); !ok {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "address invalid")
			return nil
		}
		chainType = common.ChainTypeEth
	case common.DasAlgorithmIdEth712:
		if ok, _ := regexp.MatchString("^0x[0-9a-fA-F]{40}$", req.Address); !ok {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "address invalid")
			return nil
		}
		chainType = common.ChainTypeEth
		is712 = true
	case common.DasAlgorithmIdTron:
		if tronAddr, err := common.TronBase58ToHex(req.Address); err != nil {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "address invalid")
			return nil
		} else {
			log.Info("tronAddr:", tronAddr)
		}
		chainType = common.ChainTypeTron
	case common.DasAlgorithmIdEd25519:
		if ok, _ := regexp.MatchString("^0x[0-9a-fA-F]{64}$", req.Address); !ok {
			apiResp.ApiRespErr(api_code.ApiCodeError500, "address invalid")
			return nil
		}
		chainType = common.ChainTypeMixin
	default:
		apiResp.ApiRespErr(api_code.ApiCodeError500, "algorithm_id invalid")
		return nil
	}

	lockScript, _, err := h.dasCore.FormatAddressToDasLockScript(chainType, req.Address, is712)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("FormatAddressToDasLockScript err: %s", err.Error())
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
