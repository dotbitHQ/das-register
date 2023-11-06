package handle

import (
	"das_register_server/config"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
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
func (h *HttpHandle) AccountAuctionBid(ctx *gin.Context) {
	var (
		funcName = "AccountAuctionBid"
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

	if err = h.doAccountAuctionBid(&req, &apiResp); err != nil {
		log.Error("GetAccountAuctionInfo err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doAccountAuctionBid(req *ReqAuctionBid, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuctionBid
	addrHex, err := req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	fmt.Println(addrHex.DasAlgorithmId, addrHex.DasSubAlgorithmId, addrHex.AddressHex)
	req.address, req.chainType = addrHex.AddressHex, addrHex.ChainType

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
	premiumPrice := decimal.NewFromInt(common.Premium(int64(acc.ExpiredAt), int64(nowTime)))
	//check user`s DP

	//

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasBidExpiredAccountAuction
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.chainType
	reqBuild.Address = req.address
	reqBuild.Capacity = 0
	reqBuild.AuctionInfo = AuctionInfo{
		BasicPrice:   basicPrice,
		PremiumPrice: premiumPrice,
		BidTime:      int64(nowTime),
	}
	var p auctionBidParams
	p.account = &acc
	p.basicPrice = basicPrice
	p.premiumPrice = premiumPrice
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
	account      *tables.TableAccountInfo
	basicPrice   decimal.Decimal
	premiumPrice decimal.Decimal
}

func (h *HttpHandle) buildAuctionBidTx(req *reqBuildTx, p *auctionBidParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	contractAcc, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	// inputs account cell
	accOutPoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	// witness account cell
	res, err := h.dasCore.Client().GetTransaction(h.ctx, accOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	builderMap, err := witness.AccountCellDataBuilderMapFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	}
	builder, ok := builderMap[req.Account]
	if !ok {
		return nil, fmt.Errorf("builderMap not exist account: %s", req.Account)
	}

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}

	//witness
	//-----witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasBidExpiredAccountAuction, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)
	//-----acc witness
	accWitness, accData, err := builder.GenWitness(&witness.AccountCellParam{
		OldIndex:              0,
		NewIndex:              0,
		Action:                common.DasBidExpiredAccountAuction,
		LastTransferAccountAt: timeCell.Timestamp(),
	})
	txParams.Witnesses = append(txParams.Witnesses, accWitness)

	// inputs
	//-----AccountCell
	accOutpoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutpoint,
	})

	//------DPCell

	//------NormalCell
	needCapacity := res.Transaction.Outputs[builder.Index].Capacity
	liveCell, totalCapacity, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        h.serverScript,
		CapacityNeed:      needCapacity,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	if change := totalCapacity - needCapacity; change > 0 {
		splitCkb := 2000 * common.OneCkb
		if config.Cfg.Server.SplitCkb > 0 {
			splitCkb = config.Cfg.Server.SplitCkb * common.OneCkb
		}
		changeList, err := core.SplitOutputCell2(change, splitCkb, 200, h.serverScript, nil, indexer.SearchOrderAsc)
		if err != nil {
			return nil, fmt.Errorf("SplitOutputCell2 err: %s", err.Error())
		}
		for i := 0; i < len(changeList); i++ {
			txParams.Outputs = append(txParams.Outputs, changeList[i])
			txParams.OutputsData = append(txParams.OutputsData, []byte{})
		}
	}
	// inputs
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	//output

	//-----AccountCell
	lockArgs, err := h.dasCore.Daf().HexToArgs(core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	}, core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	})
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: res.Transaction.Outputs[builder.Index].Capacity,
		Lock:     contractDas.ToScript(lockArgs),
		Type:     contractAcc.ToScript(nil),
	})
	newExpiredAt := int64(builder.ExpiredAt) + common.OneYearSec
	byteExpiredAt := molecule.Go64ToBytes(newExpiredAt)
	accData = append(accData, res.Transaction.OutputsData[builder.Index][32:]...)
	accData1 := accData[:common.ExpireTimeEndIndex-common.ExpireTimeLen]
	accData2 := accData[common.ExpireTimeEndIndex:]
	newAccData := append(accData1, byteExpiredAt...)
	newAccData = append(newAccData, accData2...)
	txParams.OutputsData = append(txParams.OutputsData, newAccData) // change expired_at

	//DPCell

	//oldowner balanceCell
	oldOwnerAddrHex := core.DasAddressHex{
		DasAlgorithmId: p.account.OwnerChainType.ToDasAlgorithmId(true),
		AddressHex:     p.account.Owner,
		IsMulti:        false,
		ChainType:      p.account.OwnerChainType,
	}
	oldOwnerLockArgs, err := h.dasCore.Daf().HexToArgs(oldOwnerAddrHex, oldOwnerAddrHex)
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: res.Transaction.Outputs[builder.Index].Capacity,
		Lock:     contractDas.ToScript(oldOwnerLockArgs),
		Type:     contractAcc.ToScript(nil),
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	return &txParams, nil
}
