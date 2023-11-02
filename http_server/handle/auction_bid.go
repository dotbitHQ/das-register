package handle

import (
	"das_register_server/config"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type ReqAuctionBid struct {
	Account   string `json:"account"`
	address   string
	chainType common.ChainType
	core.ChainTypeAddress
}

type RespAuctionBid struct {
	SignInfo
}

//查询价格
func (h *HttpHandle) GetAccountAuctionBid(ctx *gin.Context) {
	var (
		funcName = "GetAccountAuctionPrice"
		clientIp = GetClientIp(ctx)
		req      ReqAuctionBid
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doGetAccountAuctionBid(&req, &apiResp); err != nil {
		log.Error("GetAccountAuctionInfo err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doGetAccountAuctionBid(req *ReqAuctionBid, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuctionBid
	if req.Account == "" {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid: account is empty")
		return
	}
	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	nowTime := uint64(time.Now().Unix())
	//exp + 90 + 27 +3
	//now > exp+117 exp< now - 117
	//now< exp+90 exp>now -90
	if acc.ExpiredAt > nowTime-90*24*3600 || acc.ExpiredAt < nowTime-117*24*3600 {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return
	}
	//计算长度
	_, accLen, err := common.GetDotBitAccountLength(req.Account)
	if err != nil {
		return
	}
	if accLen == 0 {
		err = fmt.Errorf("accLen is 0")
		return
	}
	baseAmount, accountPrice, err := h.getAccountPrice(uint8(accLen), "", req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	basicPrice := baseAmount.Add(accountPrice)
	premiumPrice := common.Premium(int64(acc.ExpiredAt))
	//check user`s DP

	//
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	fmt.Println(addrHex.DasAlgorithmId, addrHex.DasSubAlgorithmId, addrHex.AddressHex)
	req.address, req.chainType = addrHex.AddressHex, addrHex.ChainType

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionTransferAccount
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.chainType
	reqBuild.Address = req.address
	reqBuild.Capacity = 0
	var p auctionBidParams
	p.account = &acc
	txParams, err := h.buildAuctionBidTx(&reqBuild, &p)

	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildEditManagerTx err: %s", err.Error())
	}
	if si, err := h.buildTx(&reqBuild, txParams); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

type auctionBidParams struct {
	account *tables.TableAccountInfo
}

func (h *HttpHandle) buildAuctionBidTx(req *reqBuildTx, p *auctionBidParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs account cell
	accOutPoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})
	return &txParams, nil
}
