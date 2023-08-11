package handle

import (
	"das_register_server/config"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqCkbRpc struct {
	ID      int             `json:"id"`
	JsonRpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type RespCkbRpc struct {
	Id      int             `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
}

type ApiRespError struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func ApiRespErr(id, errNo int, errMsg string) ApiRespError {
	return ApiRespError{
		Id:      id,
		Jsonrpc: "2.0",
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{errNo, errMsg},
	}
}

func (h *HttpHandle) CkbRpc(ctx *gin.Context) {
	var (
		funcName = "CkbRpc"
		clientIp = GetClientIp(ctx)
		req      ReqCkbRpc
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		ctx.JSON(http.StatusOK, ApiRespErr(req.ID, api_code.ApiCodeParamsInvalid, "params invalid"))
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	h.doCkbRpc(ctx, &req)
}

func (h *HttpHandle) doCkbRpc(ctx *gin.Context, req *ReqCkbRpc) {
	var resp RespCkbRpc

	url := ""
	switch req.Method {
	case "get_cells", "get_cells_capacity":
		url = config.Cfg.Chain.IndexUrl
	case "get_blockchain_info", "get_block_by_number", "send_transaction":
		url = config.Cfg.Chain.CkbUrl
	default:
		ctx.JSON(http.StatusOK, ApiRespErr(req.ID, api_code.ApiCodeMethodNotExist, fmt.Sprintf("method [%s] not exist", req.Method)))
		return
	}

	res, body, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&resp)
	if len(errs) > 0 {
		log.Errorf("call rpc err: %v", errs)
		ctx.JSON(http.StatusOK, ApiRespErr(req.ID, api_code.ApiCodeError500, fmt.Sprintf("call rpc err: %v", errs)))
		return
	} else if res.StatusCode != http.StatusOK {
		ctx.JSON(http.StatusOK, ApiRespErr(req.ID, api_code.ApiCodeError500, fmt.Sprintf("http status: %d", res.StatusCode)))
		return
	}
	//log.Info("CkbNodeRpc:", req.Method, string(body))
	if resp.Result == nil {
		var apiErr ApiRespError
		_ = json.Unmarshal(body, &apiErr)
		ctx.JSON(http.StatusOK, apiErr)
	} else {
		ctx.JSON(http.StatusOK, resp)
	}
}
