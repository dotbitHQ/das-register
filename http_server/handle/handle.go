package handle

import (
	"context"
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/http_server/api_code"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/dascache"
	"github.com/DeAccountSystems/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/mylog"
)

var (
	log = mylog.NewLogger("http_handle", mylog.LevelDebug)
)

type HttpHandle struct {
	ctx                    context.Context
	dbDao                  *dao.DbDao
	rc                     *cache.RedisCache
	dasCore                *core.DasCore
	dasCache               *dascache.DasCache
	txBuilderBase          *txbuilder.DasTxBuilderBase
	mapReservedAccounts    map[string]struct{}
	mapUnAvailableAccounts map[string]struct{}
}

type HttpHandleParams struct {
	DbDao                  *dao.DbDao
	Rc                     *cache.RedisCache
	Ctx                    context.Context
	DasCore                *core.DasCore
	DasCache               *dascache.DasCache
	TxBuilderBase          *txbuilder.DasTxBuilderBase
	MapReservedAccounts    map[string]struct{}
	MapUnAvailableAccounts map[string]struct{}
}

func Initialize(p HttpHandleParams) *HttpHandle {
	hh := HttpHandle{
		dbDao:                  p.DbDao,
		rc:                     p.Rc,
		ctx:                    p.Ctx,
		dasCore:                p.DasCore,
		dasCache:               p.DasCache,
		txBuilderBase:          p.TxBuilderBase,
		mapReservedAccounts:    p.MapReservedAccounts,
		mapUnAvailableAccounts: p.MapUnAvailableAccounts,
	}
	return &hh
}

// 获取IP
func GetClientIp(ctx *gin.Context) string {
	clientIP := fmt.Sprintf("%v", ctx.Request.Header.Get("X-Real-IP"))
	return fmt.Sprintf("(%s)(%s)", clientIP, ctx.Request.RemoteAddr)
}

// post 请求绑定参数
func shouldBindJSON(ctx *gin.Context, req interface{}) (*api_code.ApiResp, error) {
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp := api_code.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return &resp, err
	} else {
		return &api_code.ApiResp{}, nil
	}
}

func (h *HttpHandle) checkSystemUpgrade(apiResp *api_code.ApiResp) error {
	if config.Cfg.Server.IsUpdate {
		apiResp.ApiRespErr(api_code.ApiCodeSystemUpgrade, "The service is under maintenance, please try again later.")
		return fmt.Errorf("backend system upgrade")
	}
	ConfigCellDataBuilder, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsMain)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	status, _ := ConfigCellDataBuilder.Status()
	if status != 1 {
		apiResp.ApiRespErr(api_code.ApiCodeSystemUpgrade, "The service is under maintenance, please try again later.")
		return fmt.Errorf("contract system upgrade")
	}
	return nil
}
