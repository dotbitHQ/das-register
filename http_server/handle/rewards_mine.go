package handle

import (
	"context"
	"das_register_server/http_server/compatible"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"net/http"
)

type ReqRewardsMine struct {
	core.ChainTypeAddress
	ChainType common.ChainType `json:"chain_type"`
	Address   string           `json:"address"`
	Pagination
}

type RespRewardsMine struct {
	Count int64           `json:"count"`
	Total decimal.Decimal `json:"total"`
	List  []RewardsData   `json:"list"`
}

type RewardsData struct {
	Invitee        string          `json:"invitee"`
	InvitationTime uint64          `json:"invitation_time"`
	Reward         decimal.Decimal `json:"reward"`
}

func (h *HttpHandle) RpcRewardsMine(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqRewardsMine
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

	if err = h.doRewardsMine(h.ctx, &req[0], apiResp); err != nil {
		log.Error("doVersion err:", err.Error())
	}
}

func (h *HttpHandle) RewardsMine(ctx *gin.Context) {
	var (
		funcName = "RewardsMine"
		clientIp = GetClientIp(ctx)
		req      ReqRewardsMine
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doRewardsMine(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doRewardsMine err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doRewardsMine(ctx context.Context, req *ReqRewardsMine, apiResp *api_code.ApiResp) error {
	var resp RespRewardsMine
	resp.List = make([]RewardsData, 0)

	addressHex, err := compatible.ChainTypeAndCoinType(*req, h.dasCore)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return err
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	list, err := h.dbDao.GetMyRewards(req.ChainType, req.Address, tables.ServiceTypeRegister, tables.RewardTypeInviter, req.GetLimit(), req.GetOffset())
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search rewards err")
		return fmt.Errorf("GetMyRewards err: %s", err.Error())
	}

	for _, v := range list {
		reward, _ := decimal.NewFromString(fmt.Sprintf("%d", v.Reward))
		resp.List = append(resp.List, RewardsData{
			Invitee:        v.InviteeAccount,
			InvitationTime: v.BlockTimestamp,
			Reward:         reward,
		})
	}

	rc, err := h.dbDao.GetMyRewardsCount(req.ChainType, req.Address, tables.ServiceTypeRegister, tables.RewardTypeInviter)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search rewards count err")
		return fmt.Errorf("GetMyRewardsCount err: %s", err.Error())
	}
	resp.Total = rc.TotalReward
	resp.Count = rc.CountNumber

	apiResp.ApiRespOK(resp)
	return nil
}
