package handle

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
)

type ReqDasOrderDetail struct {
	OrderIdList []string `json:"order_id_list"`
}

type RespDasOrderDetail struct {
	OrderDetailList []DasOrderDetail `json:"order_detail_list"`
}

type DasOrderDetail struct {
	OrderId        string                `json:"order_id"`
	OrderStatus    tables.OrderStatus    `json:"order_status"`
	Action         common.DasAction      `json:"action"`
	Account        string                `json:"account"`
	AccountId      string                `json:"account_id"`
	RegisterStatus tables.RegisterStatus `json:"register_status"`
}

func (h *HttpHandle) DasOrderDetail(ctx *gin.Context) {
	var (
		funcName = "DasOrderDetail"
		clientIp = GetClientIp(ctx)
		req      ReqDasOrderDetail
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

	if err = h.doDasOrderDetail(&req, &apiResp); err != nil {
		log.Error("doDasOrderDetail err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doDasOrderDetail(req *ReqDasOrderDetail, apiResp *api_code.ApiResp) error {
	var resp RespDasOrderDetail
	resp.OrderDetailList = make([]DasOrderDetail, 0)

	list, err := h.dbDao.GetOrderListByOrderIds(req.OrderIdList)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search order list fail")
		return fmt.Errorf("GetOrderListByOrderIds err: %s", err.Error())
	}

	for _, v := range list {
		tmp := DasOrderDetail{
			OrderId:        v.OrderId,
			OrderStatus:    v.OrderStatus,
			Action:         v.Action,
			Account:        v.Account,
			AccountId:      v.AccountId,
			RegisterStatus: v.RegisterStatus,
		}
		resp.OrderDetailList = append(resp.OrderDetailList, tmp)
	}

	apiResp.ApiRespOK(resp)
	return nil
}
