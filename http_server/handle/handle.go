package handle

import (
	"context"
	"das_register_server/cache"
	"das_register_server/config"
	"das_register_server/dao"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/txbuilder"
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

func GetClientIp(ctx *gin.Context) string {
	clientIP := fmt.Sprintf("%v", ctx.Request.Header.Get("X-Real-IP"))
	return fmt.Sprintf("(%s)(%s)", clientIP, ctx.Request.RemoteAddr)
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
	ok, err := h.dasCore.CheckContractStatusOK(common.DasContractNameAccountCellType)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
		return fmt.Errorf("CheckContractStatusOK err: %s", err.Error())
	} else if !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSystemUpgrade, "The service is under maintenance, please try again later.")
		return fmt.Errorf("contract system upgrade")
	}
	return nil
}

func checkChainType(chainType common.ChainType) bool {
	switch chainType {
	case common.ChainTypeTron, common.ChainTypeMixin, common.ChainTypeEth, common.ChainTypeDogeCoin, common.ChainTypeWebauthn:
		return true
	}
	return false
}

func checkBalanceErr(err error, apiResp *api_code.ApiResp) {
	if err == core.ErrRejectedOutPoint {
		apiResp.ApiRespErr(api_code.ApiCodeRejectedOutPoint, err.Error())
	} else if err == core.ErrNotEnoughChange {
		apiResp.ApiRespErr(api_code.ApiCodeNotEnoughChange, err.Error())
	} else if err == core.ErrInsufficientFunds {
		apiResp.ApiRespErr(api_code.ApiCodeInsufficientBalance, err.Error())
	} else {
		apiResp.ApiRespErr(api_code.ApiCodeError500, err.Error())
	}
}
