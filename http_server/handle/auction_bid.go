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
	"github.com/nervosnetwork/ckb-sdk-go/address"
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
	fromLock, _, err := h.dasCore.Daf().HexToScript(*addrHex)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "key info is invalid: "+err.Error())
		return nil
	}
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
	if status, _, err := h.checkDutchAuction(acc.ExpiredAt); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "checkDutchAuction err")
		return fmt.Errorf("checkDutchAuction err: %s", err.Error())
	} else if status != tables.SearchStatusOnDutchAuction {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return nil
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
	//totalPrice := basicPrice.Add(premiumPrice)
	//check user`s DP
	amountDP := basicPrice.Add(premiumPrice).BigInt().Uint64() * common.UsdRateBase
	log.Info("GetDpCells:", common.Bytes2Hex(fromLock.Args), amountDP)
	_, _, _, err = h.dasCore.GetDpCells(&core.ParamGetDpCells{
		DasCache:           h.dasCache,
		LockScript:         fromLock,
		AmountNeed:         amountDP,
		CurrentBlockNumber: 0,
		SearchOrder:        indexer.SearchOrderAsc,
	})
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("dasCore.GetDpCells err: ", err.Error())
	}
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

	// to lock & normal cell lock
	//转账地址 用于接收荷兰拍竞拍dp的地址
	if config.Cfg.Server.TransferWhitelist == "" || config.Cfg.Server.CapacityWhitelist == "" {
		return fmt.Errorf("TransferWhitelist or CapacityWhitelist is empty")
	}
	toLock, err := address.Parse(config.Cfg.Server.TransferWhitelist)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	//回收地址（接收dpcell 释放的capacity）
	normalCellLock, err := address.Parse(config.Cfg.Server.CapacityWhitelist)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	var p auctionBidParams
	p.Account = &acc
	p.AmountDP = amountDP
	p.FromLock = fromLock
	p.ToLock = toLock.Script
	p.NormalCellLock = normalCellLock.Script
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
	Account        *tables.TableAccountInfo
	AmountDP       uint64
	FromLock       *types.Script
	ToLock         *types.Script
	NormalCellLock *types.Script
}

func (h *HttpHandle) buildAuctionBidTx(req *reqBuildTx, p *auctionBidParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	balanceContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}

	accOutPoint := common.String2OutPointStruct(p.Account.Outpoint)
	// witness account cell
	accTx, err := h.dasCore.Client().GetTransaction(h.ctx, accOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}

	builderMap, err := witness.AccountCellDataBuilderMapFromTx(accTx.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	}
	builder, ok := builderMap[req.Account]
	if !ok {
		return nil, fmt.Errorf("builderMap not exist account: %s", req.Account)
	}
	accCellCapacity := accTx.Transaction.Outputs[builder.Index].Capacity
	oldAccOwnerArgs := accTx.Transaction.Outputs[builder.Index].Lock.Args

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
		RegisterAt:            uint64(timeCell.Timestamp()),
	})
	txParams.Witnesses = append(txParams.Witnesses, accWitness)

	//input account cell
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	//output account cell
	newOwnerAddrHex := core.DasAddressHex{
		DasAlgorithmId: req.ChainType.ToDasAlgorithmId(true),
		AddressHex:     req.Address,
		IsMulti:        false,
		ChainType:      req.ChainType,
	}
	lockArgs, err := h.dasCore.Daf().HexToArgs(newOwnerAddrHex, newOwnerAddrHex)

	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: accTx.Transaction.Outputs[builder.Index].Capacity,
		Lock:     contractDas.ToScript(lockArgs),
		Type:     accTx.Transaction.Outputs[builder.Index].Type,
	})
	newExpiredAt := timeCell.Timestamp() + common.OneYearSec
	byteExpiredAt := molecule.Go64ToBytes(newExpiredAt)
	accData = append(accData, accTx.Transaction.OutputsData[builder.Index][32:]...)
	accData1 := accData[:common.ExpireTimeEndIndex-common.ExpireTimeLen]
	accData2 := accData[common.ExpireTimeEndIndex:]
	newAccData := append(accData1, byteExpiredAt...)
	newAccData = append(newAccData, accData2...)
	fmt.Println("newAccData: ", newAccData)
	txParams.OutputsData = append(txParams.OutputsData, newAccData) // change expired_at

	//dp
	liveCell, totalDP, totalCapacity, err := h.dasCore.GetDpCells(&core.ParamGetDpCells{
		DasCache:           h.dasCache,
		LockScript:         p.FromLock,
		AmountNeed:         p.AmountDP,
		CurrentBlockNumber: 0,
		SearchOrder:        indexer.SearchOrderAsc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetDpCells err: %s", err.Error())
	}

	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// outputs
	outputs, outputsData, normalCellCapacity, err := h.dasCore.SplitDPCell(&core.ParamSplitDPCell{
		FromLock:           p.FromLock,    //发送dp方的lock
		ToLock:             p.ToLock,      //接收dp方的lock
		DPLiveCell:         liveCell,      //发送方的dp cell
		DPLiveCellCapacity: totalCapacity, //发送方dp的capacity
		DPTotalAmount:      totalDP,       //总的dp金额
		DPTransferAmount:   p.AmountDP,    //要转账的dp金额
		DPSplitCount:       config.Cfg.Server.SplitCount,
		DPSplitAmount:      config.Cfg.Server.SplitAmount,
		NormalCellLock:     p.NormalCellLock, //回收dp cell的ckb的接收地址
	})
	if err != nil {
		return nil, fmt.Errorf("SplitDPCell err: %s", err.Error())
	}
	for i, _ := range outputs {
		txParams.Outputs = append(txParams.Outputs, outputs[i])
		txParams.OutputsData = append(txParams.OutputsData, outputsData[i])
	}

	//input dp cell capacity < output dp cell capacity 需要用注册商的nomal cell 来垫付
	//input dp cell capacity > output dp cell capacity : has been return in "h.dasCore.SplitDPCell"
	normalCells, totalNormal, err := h.dasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          h.dasCache,
		LockScript:        h.serverScript,
		CapacityNeed:      normalCellCapacity + accCellCapacity,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	for _, v := range normalCells {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}
	//返还给旧账户的 capacity
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: accCellCapacity,
		Lock:     contractDas.ToScript(oldAccOwnerArgs),
		Type:     balanceContract.ToScript(nil),
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	log.Info("normalCellCapacity:", normalCellCapacity, common.Bytes2Hex(p.NormalCellLock.Args))
	//找零
	if change := totalNormal - normalCellCapacity - accCellCapacity; change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     h.serverScript,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// cell deps
	configCellAcc, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	accContract, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	configCellMain, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsMain)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	configCellDP, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsDPoint)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	contractDP, err := core.GetDasContractInfo(common.DasContractNameDpCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	txParams.CellDeps = append(txParams.CellDeps,
		configCellAcc.ToCellDep(),
		configCellMain.ToCellDep(),
		configCellDP.ToCellDep(),
		contractDP.ToCellDep(),
		timeCell.ToCellDep(),
		accContract.ToCellDep(),
		contractDas.ToCellDep(),
	)
	return &txParams, nil
}
