package api_code

import (
	"bytes"
	"das_register_server/config"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/mylog"
	"net/http"
	"time"
)

var log = mylog.NewLogger("api_code", mylog.LevelDebug)

type ReqPushLog struct {
	Index   string        `json:"index"`
	Method  string        `json:"method"`
	Ip      string        `json:"ip"`
	Latency time.Duration `json:"latency"`
	ErrMsg  string        `json:"err_msg"`
	ErrNo   int           `json:"err_no"`
}

func PushLog(url string, req ReqPushLog) {
	if url == "" {
		return
	}
	go func() {
		resp, _, errs := gorequest.New().Post(url).SendStruct(&req).End()
		if len(errs) > 0 {
			log.Error("PushLog err:", errs)
		} else if resp.StatusCode != http.StatusOK {
			log.Error("PushLog StatusCode:", resp.StatusCode)
		}
	}()
}

func DoMonitorLog(method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ip := getClientIp(ctx)

		blw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		ctx.Next()
		statusCode := ctx.Writer.Status()

		if statusCode == http.StatusOK && blw.body.String() != "" {
			var resp ApiResp
			if err := json.Unmarshal(blw.body.Bytes(), &resp); err == nil {
				if resp.ErrNo != ApiCodeSuccess {
					log.Warn("DoMonitorLog:", method, resp.ErrNo, resp.ErrMsg)
				}
				if method == MethodTransactionStatus && resp.ErrNo == ApiCodeTransactionNotExist {
					resp.ErrNo = ApiCodeSuccess
				} else if method == MethodOrderDetail && resp.ErrNo == ApiCodeOrderNotExist {
					resp.ErrNo = ApiCodeSuccess
				}
				pushLog := ReqPushLog{
					Index:   config.Cfg.Server.PushLogIndex,
					Method:  method,
					Ip:      ip,
					Latency: time.Since(startTime),
					ErrMsg:  resp.ErrMsg,
					ErrNo:   resp.ErrNo,
				}
				PushLog(config.Cfg.Server.PushLogUrl, pushLog)
			}
		}
	}
}

func DoMonitorLogRpc(apiResp *ApiResp, method, clientIp string, startTime time.Time) {
	pushLog := ReqPushLog{
		Index:   config.Cfg.Server.PushLogIndex,
		Method:  method,
		Ip:      clientIp,
		Latency: time.Since(startTime),
		ErrMsg:  apiResp.ErrMsg,
		ErrNo:   apiResp.ErrNo,
	}
	if apiResp.ErrNo != ApiCodeSuccess {
		log.Warn("DoMonitorLog:", method, apiResp.ErrNo, apiResp.ErrMsg)
	}
	PushLog(config.Cfg.Server.PushLogUrl, pushLog)
}

func getClientIp(ctx *gin.Context) string {
	return fmt.Sprintf("%v", ctx.Request.Header.Get("X-Real-IP"))
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (b bodyWriter) Write(bys []byte) (int, error) {
	b.body.Write(bys)
	return b.ResponseWriter.Write(bys)
}
