package handle

import (
	"encoding/json"
	"github.com/dotbitHQ/das-lib/common"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqCharacterSetList struct {
	CharType common.AccountCharType `json:"char_type"`
}

type RespCharacterSetList struct {
	EmojiList []string `json:"emoji_list"`
	DigitList []string `json:"digit_list"`
	EnList    []string `json:"en_list"`
	KoList    []string `json:"ko_list"`
	ViList    []string `json:"vi_list"`
	ThList    []string `json:"th_list"`
	TrList    []string `json:"tr_list"`
}

func (h *HttpHandle) RpcCharacterSetList(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqCharacterSetList
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

	if err = h.doCharacterSetList(&req[0], apiResp); err != nil {
		log.Error("doCharacterSetList err:", err.Error())
	}
}

func (h *HttpHandle) CharacterSetList(ctx *gin.Context) {
	var (
		funcName = "CharacterSetList"
		clientIp = GetClientIp(ctx)
		req      ReqCharacterSetList
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

	if err = h.doCharacterSetList(&req, &apiResp); err != nil {
		log.Error("doCharacterSetList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCharacterSetList(req *ReqCharacterSetList, apiResp *api_code.ApiResp) error {
	var resp RespCharacterSetList

	for k, _ := range common.CharSetTypeEmojiMap {
		resp.EmojiList = append(resp.EmojiList, k)
	}
	for k, _ := range common.CharSetTypeDigitMap {
		resp.DigitList = append(resp.DigitList, k)
	}
	for k, _ := range common.CharSetTypeEnMap {
		resp.EnList = append(resp.EnList, k)
	}
	for k, _ := range common.CharSetTypeKoMap {
		resp.KoList = append(resp.KoList, k)
	}
	for k, _ := range common.CharSetTypeViMap {
		resp.ViList = append(resp.ViList, k)
	}
	for k, _ := range common.CharSetTypeThMap {
		resp.ThList = append(resp.ThList, k)
	}
	for k, _ := range common.CharSetTypeTrMap {
		resp.TrList = append(resp.TrList, k)
	}

	apiResp.ApiRespOK(resp)
	return nil
}
