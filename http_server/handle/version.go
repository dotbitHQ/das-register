package handle

import (
	"das_register_server/http_server/api_code"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"time"
)

type ReqVersion struct {
}

type RespVersion struct {
	Version string `json:"version"`
}

func (h *HttpHandle) RpcVersion(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqVersion
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

	if err = h.doVersion(&req[0], apiResp); err != nil {
		log.Error("doVersion err:", err.Error())
	}
}

func (h *HttpHandle) Version(ctx *gin.Context) {
	var (
		funcName = "Version"
		clientIp = GetClientIp(ctx)
		req      ReqVersion
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

	if err = h.doVersion(&req, &apiResp); err != nil {
		log.Error("doVersion err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doVersion(req *ReqVersion, apiResp *api_code.ApiResp) error {
	var resp RespVersion

	resp.Version = time.Now().String()

	apiResp.ApiRespOK(resp)
	return nil
}
