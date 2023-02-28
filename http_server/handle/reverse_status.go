package handle

import (
	"das_register_server/http_server/api_code"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqReverseStatus struct {
	core.ChainTypeAddress
}

type RespReverseStatus struct {
	Status int `json:"status"`
}

func (h *HttpHandle) ReverseStatus(ctx *gin.Context) {
	var (
		funcName = "ReverseStatus"
		clientIp = GetClientIp(ctx)
		req      ReqReverseStatus
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

	if err = h.doReverseStatus(&req, &apiResp); err != nil {
		log.Errorf("doReverseStatus err: %s funcName: %s clientIp: %s", err, funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

// doReverseStatus
func (h *HttpHandle) doReverseStatus(req *ReqReverseStatus, apiResp *api_code.ApiResp) error {
	res := checkReqKeyInfo(h.dasCore.Daf(), &req.ChainTypeAddress, apiResp)
	address := strings.ToLower(res.AddressHex)
	reverseRecord, err := h.dbDao.GetReverseSmtRecordByAddress(address, uint8(res.DasAlgorithmId))
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "db error")
		return fmt.Errorf("GetReverseSmtRecordByAddress err: %s", err)
	}
	resp := &RespReverseStatus{}
	apiResp.Data = resp

	if reverseRecord.ID == 0 {
		return nil
	}
	if reverseRecord.TaskID == "" {
		resp.Status = 1
		return nil
	}

	reverseTaskInfo, err := h.dbDao.GetLatestReverseSmtTaskByTaskID(reverseRecord.TaskID)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "db error")
		return fmt.Errorf("GetLatestReverseSmtTaskByTaskID err: %s", err)
	}
	if reverseTaskInfo.SmtStatus == tables.ReverseSmtStatusConfirm &&
		reverseTaskInfo.TxStatus == tables.ReverseSmtTxStatusConfirm ||
		reverseTaskInfo.SmtStatus == tables.ReverseSmtStatusRollbackConfirm {
		return nil
	}

	resp.Status = 1
	return nil
}
