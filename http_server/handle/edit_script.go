package handle

import (
	"das_register_server/http_server/api_code"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqEditScript struct {
}

type RespEditScript struct {
}

func (h *HttpHandle) RpcEditScript(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqEditScript
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

	if err = h.doEditScript(&req[0], apiResp); err != nil {
		log.Error("doEditScript err:", err.Error())
	}
}

func (h *HttpHandle) EditScript(ctx *gin.Context) {
	var (
		funcName = "EditScript"
		clientIp = GetClientIp(ctx)
		req      ReqEditScript
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

	if err = h.doEditScript(&req, &apiResp); err != nil {
		log.Error("doEditScript err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEditScript(req *ReqEditScript, apiResp *api_code.ApiResp) error {
	var resp RespEditScript

	apiResp.ApiRespOK(resp)
	return nil
}
