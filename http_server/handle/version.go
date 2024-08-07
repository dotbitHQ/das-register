package handle

import (
	"context"
	"encoding/json"
	api_code "github.com/dotbitHQ/das-lib/http_api"
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

	if err = h.doVersion(h.ctx, &req[0], apiResp); err != nil {
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doVersion(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doVersion err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doVersion(ctx context.Context, req *ReqVersion, apiResp *api_code.ApiResp) error {
	var resp RespVersion
	resp.Version = time.Now().String()
	apiResp.ApiRespOK(resp)
	return nil
}
