package handle

import (
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqSignTx struct {
	ChainId  int    `json:"chain_id"`
	Private  string `json:"private"`
	Compress bool   `json:"compress"`
	SignInfo
}

type RespSignTx struct {
	SignInfo
}

func (h *HttpHandle) RpcSignTx(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqSignTx
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

	if err = h.doSignTx(&req[0], apiResp); err != nil {
		log.Error("doSignTx err:", err.Error())
	}
}

func (h *HttpHandle) SignTx(ctx *gin.Context) {
	var (
		funcName = "SignTx"
		clientIp = GetClientIp(ctx)
		req      ReqSignTx
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

	if err = h.doSignTx(&req, &apiResp); err != nil {
		log.Error("doSignTx err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doSignTx(req *ReqSignTx, apiResp *api_code.ApiResp) error {
	var resp RespSignTx
	resp.SignKey = req.SignKey
	var signData []byte
	var err error
	for _, v := range req.SignList {
		switch v.SignType {
		case common.DasAlgorithmIdTron:
			signData, err = sign.TronSignature(true, common.Hex2Bytes(v.SignMsg), req.Private)
			if err != nil {
				err = fmt.Errorf("sign.TronSignature err: %s", err.Error())
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return err
			}
		case common.DasAlgorithmIdEth:
			signData, err = sign.PersonalSignature(common.Hex2Bytes(v.SignMsg), req.Private)
			if err != nil {
				err = fmt.Errorf("sign.PersonalSignature err: %s", err.Error())
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return err
			}
		case common.DasAlgorithmIdEd25519:
			signData = sign.Ed25519Signature(common.Hex2Bytes(req.Private), common.Hex2Bytes(v.SignMsg))
			signData = append(signData, []byte{1}...)
		case common.DasAlgorithmIdEth712:
			var obj3 apitypes.TypedData
			mmJson := req.MMJson.String()

			log.Info("old mmJson:", mmJson)
			oldChainId := fmt.Sprintf("chainId\":%d", req.ChainId)
			newChainId := fmt.Sprintf("chainId\":\"%d\"", req.ChainId)
			mmJson = strings.ReplaceAll(mmJson, oldChainId, newChainId)
			oldDigest := "\"digest\":\"\""
			newDigest := fmt.Sprintf("\"digest\":\"%s\"", v.SignMsg)
			mmJson = strings.ReplaceAll(mmJson, oldDigest, newDigest)
			log.Info("new mmJson:", mmJson)
			_ = json.Unmarshal([]byte(mmJson), &obj3)
			var mmHash, signature []byte
			mmHash, signature, err = sign.EIP712Signature(obj3, req.Private)
			if err != nil {
				err = fmt.Errorf("sign.EIP712Signature err: %s", err.Error())
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return err
			}
			log.Info("EIP712Signature mmHash:", common.Bytes2Hex(mmHash))
			log.Info("EIP712Signature signature:", common.Bytes2Hex(signature))
			signData = append(signature, mmHash...)

			hexChainId := fmt.Sprintf("%x", req.ChainId)
			chainIdData := common.Hex2Bytes(fmt.Sprintf("%016s", hexChainId))
			signData = append(signData, chainIdData...)
		case common.DasAlgorithmIdDogeChain:
			signData, err = sign.DogeSignature(common.Hex2Bytes(v.SignMsg), req.Private, req.Compress)
			if err != nil {
				err = fmt.Errorf("sign.DogeSignature err: %s", err.Error())
				apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
				return err
			}
		default:
			//apiResp.ApiRespErr(api_code.ApiCodeError500, fmt.Sprintf("not support sign type [%d]", v.SignType))
			//return nil
			signData = []byte{}
		}
		resp.SignList = append(resp.SignList, txbuilder.SignData{
			SignType: v.SignType,
			SignMsg:  common.Bytes2Hex(signData),
		})
	}
	apiResp.ApiRespOK(resp)
	return nil
}
