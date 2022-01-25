package handle

import (
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/http_server/api_code"
	"das_register_server/notify"
	"das_register_server/timer"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
)

type ReqOrderInfo struct {
}

type RespOrderInfo struct {
}

func (h *HttpHandle) RpcOrderInfo(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqOrderInfo
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

	if err = h.doOrderInfo(&req[0], apiResp); err != nil {
		log.Error("doOrderInfo err:", err.Error())
	}
}

func (h *HttpHandle) OrderInfo(ctx *gin.Context) {
	var (
		funcName = "OrderInfo"
		clientIp = GetClientIp(ctx)
		req      ReqOrderInfo
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

	if err = h.doOrderInfo(&req, &apiResp); err != nil {
		log.Error("doOrderInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doOrderInfo(req *ReqOrderInfo, apiResp *api_code.ApiResp) error {
	var resp RespOrderInfo
	// register info
	list, err := h.dbDao.GetAccountNumRegisterNum()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetAccountNumRegisterNum err: %s", err.Error())
	}
	msg := GetAccountNumRegisterNumStr(list)

	// order info
	listOrder, err := h.dbDao.GetOrderTotalAmount()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetOrderTotalAmount err: %s", err.Error())
	}

	// refund
	listRefund, err := h.dbDao.GetOrderRefundTotalAmount()
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("GetOrderRefundTotalAmount err: %s", err.Error())
	}
	msg2 := GetOrderAmountStr(listOrder, listRefund)
	res := fmt.Sprintf("%s\n%s", msg, msg2)
	notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Das Info", res)
	apiResp.ApiRespOK(resp)
	return nil
}

func GetOrderAmountStr(list, listRefund []dao.OrderTotalAmount) string {
	msg := `> Order Info
%s
> Refund List
%s`
	orderStr := ""
	for _, v := range list {
		t := timer.GetTokenInfo(v.PayTokenId)
		if t.Symbol == "" {
			continue
		}
		orderStr += fmt.Sprintf("- %s(%d) %s\n", v.PayTokenId, v.Num, v.Amount.DivRound(decimal.New(1, t.Decimals), t.Decimals))
	}
	refundStr := ""
	for _, v := range listRefund {
		t := timer.GetTokenInfo(v.PayTokenId)
		if t.Symbol == "" {
			continue
		}
		refundStr += fmt.Sprintf("- %s(%d) %s\n", v.PayTokenId, v.Num, v.Amount.DivRound(decimal.New(1, t.Decimals), t.Decimals))
	}
	msg = fmt.Sprintf(msg, orderStr, refundStr)
	return msg
}

func GetAccountNumRegisterNumStr(list []dao.AccountNumRegisterNum) string {
	msg := `> Number of registrations per account length
%s%s`
	numStr := ""
	num2Str := "- [10+]: %d\n"
	numMore10 := 0
	for _, v := range list {
		if v.Num < 10 {
			numStr += fmt.Sprintf("- [%d]: %d\n", v.Num, v.Total)
		} else {
			numMore10 += v.Total
		}
	}
	num2Str = fmt.Sprintf(num2Str, numMore10)
	msg = fmt.Sprintf(msg, numStr, num2Str)
	return msg
}
