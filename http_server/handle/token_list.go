package handle

import (
	"das_register_server/tables"
	"das_register_server/timer"
	"encoding/json"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
)

type RespTokenList struct {
	TokenList []TokenData `json:"token_list"`
}

type TokenData struct {
	TokenId   tables.PayTokenId `json:"token_id"`
	ChainType int               `json:"chain_type"`
	CoinType  string            `json:"coin_type"`
	Contract  string            `json:"contract"`
	Name      string            `json:"name"`
	Symbol    string            `json:"symbol"`
	Decimals  int32             `json:"decimals"`
	Logo      string            `json:"logo"`
	Price     decimal.Decimal   `json:"price"`
}

func (h *HttpHandle) RpcTokenList(p json.RawMessage, apiResp *api_code.ApiResp) {
	if err := h.doTokenList(apiResp); err != nil {
		log.Error("doTokenList err:", err.Error())
	}
}

func (h *HttpHandle) TokenList(ctx *gin.Context) {
	var (
		funcName = "TokenList"
		clientIp = GetClientIp(ctx)
		apiResp  api_code.ApiResp
		err      error
	)
	log.Info("ApiReq:", funcName, clientIp, ctx)

	if err = h.doTokenList(&apiResp); err != nil {
		log.Error("doTokenList err:", err.Error(), funcName, clientIp, ctx)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doTokenList(apiResp *api_code.ApiResp) error {
	var resp RespTokenList

	resp.TokenList = make([]TokenData, 0)
	list := timer.GetTokenList()
	for _, v := range list {
		resp.TokenList = append(resp.TokenList, TokenData{
			TokenId:   v.TokenId,
			ChainType: v.ChainType,
			CoinType:  v.CoinType,
			Name:      v.Name,
			Symbol:    v.Symbol,
			Decimals:  v.Decimals,
			Logo:      v.Logo,
			Price:     v.Price,
		})
	}
	apiResp.ApiRespOK(resp)
	return nil
}
