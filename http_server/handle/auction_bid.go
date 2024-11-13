package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/json"
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
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
)

type ReqAuctionBid struct {
	Account               string       `json:"account"  binding:"required"`
	CoinType              string       `json:"coin_type"`    //default record
	core.ChainTypeAddress                                    // ccc address
	PayKeyInfo            core.KeyInfo `json:"pay_key_info"` //pay address
	addressHex            *core.DasAddressHex
}

type RespAuctionBid struct {
	SignInfo
}

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

	if err = h.doAccountAuctionBid(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAccountAuctionBid err:", err.Error(), funcName, clientIp)
	}
	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doAccountAuctionBid(ctx context.Context, req *ReqAuctionBid, apiResp *http_api.ApiResp) (err error) {
	var resp RespAuctionBid
	//new owner address (ccc address)
	req.addressHex, err = req.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params is invalid: "+err.Error())
		return nil
	}
	//fromLock, _, err := h.dasCore.Daf().HexToScript(*req.addressHex)
	//if err != nil {
	//	apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "key info is invalid: "+err.Error())
	//	return nil
	//}

	if req.addressHex.DasAlgorithmId != common.DasAlgorithmIdAnyLock {
		apiResp.ApiRespErr(http_api.ApiCodeAnyLockAddressInvalid, "address invalid")
		return nil
	}
	if req.addressHex.DasAlgorithmId == common.DasAlgorithmIdAnyLock &&
		req.addressHex.ParsedAddress.Script.CodeHash.Hex() == transaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH {
		apiResp.ApiRespErr(http_api.ApiCodeAnyLockAddressInvalid, "address invalid")
		return nil
	}
	newOwnerLock := req.addressHex.ParsedAddress.Script

	//pay address daslock
	toCTA := core.ChainTypeAddress{
		Type:    "blockchain",
		KeyInfo: req.PayKeyInfo,
	}
	payAddrHex, err := toCTA.FormatChainTypeAddress(config.Cfg.Server.Net, true)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeInvalidTargetAddress, "receiver address is invalid")
		return fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
	}
	fromLock, _, err := h.dasCore.Daf().HexToScript(*payAddrHex)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "key info is invalid: "+err.Error())
		return nil
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil && err != gorm.ErrRecordNotFound {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	if acc.Status != tables.AccountStatusNormal && acc.Status != tables.AccountStatusOnCross {
		apiResp.ApiRespErr(http_api.ApiCodeAccountStatusNotNormal, "account status is not normal or on cross")
		return
	}
	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "GetTimeCell err")
		return fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	nowTime := timeCell.Timestamp()

	if status, _, err := h.checkDutchAuction(ctx, acc.ExpiredAt, uint64(nowTime)); err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "checkDutchAuction err")
		return fmt.Errorf("checkDutchAuction err: %s", err.Error())
	} else if status != tables.SearchStatusOnDutchAuction {
		apiResp.ApiRespErr(http_api.ApiCodeAuctionAccountNotFound, "This account has not been in dutch auction")
		return nil
	}

	_, accLen, err := common.GetDotBitAccountLength(req.Account)
	if err != nil {
		return
	}
	if accLen == 0 {
		err = fmt.Errorf("accLen is 0")
		return
	}
	_, baseAmount, accountPrice, err := h.getAccountPrice(ctx, req.addressHex, uint8(accLen), req.Account, false)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "get account price err")
		return fmt.Errorf("getAccountPrice err: %s", err.Error())
	}
	basicPrice := baseAmount.Add(accountPrice)

	log.Info(ctx, "expiredat: ", int64(acc.ExpiredAt), "nowTime: ", nowTime)

	auctionConfig, err := h.GetAuctionConfig(h.dasCore)
	if err != nil {
		err = fmt.Errorf("GetAuctionConfig err: %s", err.Error())
		return
	}
	premiumPrice := decimal.NewFromFloat(common.Premium(int64(acc.ExpiredAt+uint64(auctionConfig.GracePeriodTime)), nowTime))
	amountDP := basicPrice.Add(premiumPrice).Mul(decimal.NewFromInt(common.UsdRateBase)).BigInt().Uint64()
	log.Info(ctx, "baseAmount: ", baseAmount, " accountPrice: ", accountPrice, " basicPrice: ", basicPrice, " premiumPrice: ", premiumPrice, " amountDP: ", amountDP)

	log.Info(ctx, "GetDpCells:", common.Bytes2Hex(fromLock.Args), amountDP)
	_, _, _, err = h.dasCore.GetDpCells(&core.ParamGetDpCells{
		DasCache:           h.dasCache,
		LockScript:         fromLock,
		AmountNeed:         amountDP,
		CurrentBlockNumber: 0,
		SearchOrder:        indexer.SearchOrderAsc,
	})

	if err != nil {
		if err == core.ErrInsufficientFunds {
			apiResp.ApiRespErr(http_api.ApiCodeInsufficientBalance, err.Error())
			return
		} else {
			apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
			return fmt.Errorf("dasCore.GetDpCells err: ", err.Error())
		}
	}
	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionBidExpiredAccountAuction
	reqBuild.Account = req.Account
	//reqBuild.ChainType = req.addressHex.ChainType
	//reqBuild.Address = req.addressHex.AddressHex
	reqBuild.Capacity = 0
	reqBuild.AuctionInfo = AuctionInfo{
		BasicPrice:   basicPrice,
		PremiumPrice: premiumPrice,
		BidTime:      nowTime,
	}
	reqBuild.EvmChainId = req.GetChainId(config.Cfg.Server.Net)
	log.Info("doAccountAuctionBid EvmChainId:", reqBuild.EvmChainId)

	// to lock & normal cell lock
	if config.Cfg.Server.TransferWhitelist == "" || config.Cfg.Server.CapacityWhitelist == "" {
		return fmt.Errorf("TransferWhitelist or CapacityWhitelist is empty")
	}
	toLock, err := address.Parse(config.Cfg.Server.TransferWhitelist)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}

	normalCellLock, err := address.Parse(config.Cfg.Server.CapacityWhitelist)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}

	//default record
	var initialRecords []witness.Record
	coinType := req.KeyInfo.CoinType

	if addr, err := common.FormatAddressByCoinType(string(coinType), req.addressHex.AddressHex); err == nil {
		initialRecords = append(initialRecords, witness.Record{
			Key:   string(coinType),
			Type:  "address",
			Label: "",
			Value: addr,
			TTL:   300,
		})
	} else {
		log.Error(ctx, "buildOrderPreRegisterTx FormatAddressByCoinType err: ", err.Error())
	}

	var p auctionBidParams
	p.Account = &acc
	p.AmountDP = amountDP
	p.FromLock = fromLock
	p.NewOwnerLock = newOwnerLock
	p.ToLock = toLock.Script
	p.NormalCellLock = normalCellLock.Script
	p.TimeCell = timeCell
	p.DefaultRecord = initialRecords
	txParams, err := h.buildAuctionBidTx(&reqBuild, &p)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildEditManagerTx err: %s", err.Error())
	}
	if _, si, err := h.buildTx(ctx, &reqBuild, txParams); err != nil {
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
	DefaultRecord  []witness.Record
	AmountDP       uint64
	NewOwnerLock   *types.Script
	FromLock       *types.Script
	ToLock         *types.Script
	NormalCellLock *types.Script
	TimeCell       *core.TimeCell
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
	quoteCell, err := h.dasCore.GetQuoteCell()
	if err != nil {
		return nil, fmt.Errorf("GetQuoteCell err: %s", err.Error())
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

	actionWitness, err := witness.GenActionDataWitness(common.DasActionBidExpiredAccountAuction, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	//-----acc witness
	moleculeRefundLock := molecule.CkbScript2MoleculeScript(p.NewOwnerLock)
	accWitness, accData, err := builder.GenWitness(&witness.AccountCellParam{
		OldIndex:   0,
		NewIndex:   0,
		Action:     common.DasActionBidExpiredAccountAuction,
		RegisterAt: uint64(p.TimeCell.Timestamp()),
		Records:    p.DefaultRecord,
		RefundLock: &moleculeRefundLock,
	})
	txParams.Witnesses = append(txParams.Witnesses, accWitness)

	//input account cell
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	//output account cell
	newOwnerAddrHex := core.DasAddressHex{
		DasAlgorithmId: common.DasAlgorithmIdEth712,
		AddressHex:     common.BlackHoleAddress,
		IsMulti:        false,
		ChainType:      common.ChainTypeEth,
	}
	lockArgs, err := h.dasCore.Daf().HexToArgs(newOwnerAddrHex, newOwnerAddrHex)

	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: accTx.Transaction.Outputs[builder.Index].Capacity,
		Lock:     contractDas.ToScript(lockArgs),
		Type:     accTx.Transaction.Outputs[builder.Index].Type,
	})

	newExpiredAt := p.TimeCell.Timestamp() + common.OneYearSec
	byteExpiredAt := molecule.Go64ToBytes(newExpiredAt)
	accData = append(accData, accTx.Transaction.OutputsData[builder.Index][32:]...)
	accData1 := accData[:common.ExpireTimeEndIndex-common.ExpireTimeLen]
	accData2 := accData[common.ExpireTimeEndIndex:]
	newAccData := append(accData1, byteExpiredAt...)
	newAccData = append(newAccData, accData2...)
	txParams.OutputsData = append(txParams.OutputsData, newAccData)

	//----- output did cell start -----
	input0 := txParams.Inputs[0]
	indexDidCellFrom := uint64(0)
	didCellParamList := []core.GenDidCellParam{
		{
			DidCellLock: p.NewOwnerLock,
			Account:     builder.Account,
			ExpireAt:    builder.ExpiredAt,
		},
	}
	didCellList, outputsDataList, witnessList, err := h.dasCore.GenDidCellList(input0, indexDidCellFrom, didCellParamList)
	if err != nil {
		return nil, fmt.Errorf("GenDidCellList err: %s", err.Error())
	}

	txParams.LatestWitness = append(txParams.LatestWitness, witnessList[0])
	txParams.Outputs = append(txParams.Outputs, didCellList[0])
	txParams.OutputsData = append(txParams.OutputsData, outputsDataList[0])
	contractDidCell, err := core.GetDasContractInfo(common.DasContractNameDidCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	indexDidCell := uint64(1)
	didEntity := witness.DidEntity{
		Target: witness.CellMeta{
			Index:  indexDidCell,
			Source: witness.SourceTypeOutputs,
		},
		ItemId:               witness.ItemIdWitnessDataDidCellV0,
		DidCellWitnessDataV0: &witness.DidCellWitnessDataV0{Records: nil},
	}
	didCellWitness, err := didEntity.ObjToBys()
	if err != nil {
		return nil, fmt.Errorf("didEntity.ObjToBys err: %s", err.Error())
	}
	txParams.LatestWitness = append(txParams.LatestWitness, didCellWitness)
	didCellArgs, err := common.GetDidCellTypeArgs(txParams.Inputs[0], indexDidCell)
	if err != nil {
		return nil, fmt.Errorf("common.GetDidCellTypeArgs err: %s", err.Error())
	}
	didCell := types.CellOutput{
		Capacity: 0,
		Lock:     p.NewOwnerLock, //ccc address lock
		Type:     contractDidCell.ToScript(didCellArgs),
	}
	didCellDataLV := witness.DidCellDataLV{
		Flag:        witness.DidCellDataLVFlag,
		Version:     witness.DidCellDataLVVersion,
		WitnessHash: didEntity.HashBys(),
		ExpireAt:    builder.ExpiredAt,
		Account:     builder.Account,
	}
	contentBys, err := didCellDataLV.ObjToBys()
	if err != nil {
		return nil, fmt.Errorf("contentBys.ObjToBys err: %s", err.Error())
	}
	sporeData := witness.SporeData{
		ContentType: []byte{},
		Content:     contentBys,
		ClusterId:   witness.GetClusterId(h.dasCore.NetType()),
	}
	didCellDataBys, err := sporeData.ObjToBys()
	if err != nil {
		return nil, fmt.Errorf("sporeData.ObjToBys err: %s", err.Error())
	}
	didCellCapacity, err := h.dasCore.GetDidCellOccupiedCapacity(didCell.Lock, didCellDataLV.Account)
	if err != nil {
		return nil, fmt.Errorf("GetDidCellOccupiedCapacity err: %s", err.Error())
	}
	didCell.Capacity = didCellCapacity
	txParams.Outputs = append(txParams.Outputs, &didCell)
	txParams.OutputsData = append(txParams.OutputsData, didCellDataBys)
	//----- output did cell end -----

	//didcell storage fee
	didCellCapacity, err = h.dasCore.GetDidCellOccupiedCapacity(p.NewOwnerLock, req.Account)
	if err != nil {
		err = fmt.Errorf("GetDidCellOccupiedCapacity err: %s", err.Error())
		return nil, err
	}
	quote := quoteCell.Quote()
	decQuote, _ := decimal.NewFromString(fmt.Sprintf("%d", quote))
	decUsdRateBase := decimal.NewFromInt(common.UsdRateBase)
	didCellAmount, _ := decimal.NewFromString(fmt.Sprintf("%d", didCellCapacity/common.OneCkb))
	didCellAmount = didCellAmount.Mul(decQuote).DivRound(decUsdRateBase, 6)

	//dp
	liveCell, totalDP, totalCapacity, err := h.dasCore.GetDpCells(&core.ParamGetDpCells{
		DasCache:           h.dasCache,
		LockScript:         p.FromLock,
		AmountNeed:         p.AmountDP + didCellAmount.BigInt().Uint64(),
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
		FromLock:           p.FromLock,
		ToLock:             p.ToLock,
		DPLiveCell:         liveCell,
		DPLiveCellCapacity: totalCapacity,
		DPTotalAmount:      totalDP,
		DPTransferAmount:   p.AmountDP,
		DPSplitCount:       config.Cfg.Server.SplitCount,
		DPSplitAmount:      config.Cfg.Server.SplitAmount,
		NormalCellLock:     p.NormalCellLock,
	})
	if err != nil {
		return nil, fmt.Errorf("SplitDPCell err: %s", err.Error())
	}
	for i, _ := range outputs {
		txParams.Outputs = append(txParams.Outputs, outputs[i])
		txParams.OutputsData = append(txParams.OutputsData, outputsData[i])
	}

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

	//old owner capacity
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: accCellCapacity,
		Lock:     contractDas.ToScript(oldAccOwnerArgs),
		Type:     balanceContract.ToScript(nil),
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte{})

	log.Info("normalCellCapacity:", normalCellCapacity, common.Bytes2Hex(p.NormalCellLock.Args))
	//change
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
	heightCell, err := h.dasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}
	priceConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsPrice)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	txParams.CellDeps = append(txParams.CellDeps,
		configCellAcc.ToCellDep(),
		priceConfig.ToCellDep(),
		configCellMain.ToCellDep(),
		configCellDP.ToCellDep(),
		contractDP.ToCellDep(),
		p.TimeCell.ToCellDep(),
		accContract.ToCellDep(),
		contractDas.ToCellDep(),
		heightCell.ToCellDep(),
		quoteCell.ToCellDep(),
	)
	return &txParams, nil
}

type AuctionConfig struct {
	GracePeriodTime, AuctionPeriodTime, DeliverPeriodTime uint32
}

func (h *HttpHandle) GetAuctionConfig(dasCore *core.DasCore) (res *AuctionConfig, err error) {
	builderConfigCell, err := dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsAccount)
	var gracePeriodTime, auctionPeriodTime, deliverPeriodTime uint32
	if err != nil {
		var builderCache core.CacheConfigCellBase
		strCache, errCache := h.dasCore.GetConfigCellByCache(core.CacheConfigCellKeyBase)
		if errCache != nil {
			log.Error("GetConfigCellByCache err: ", err.Error())
			err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
			return
		} else if strCache == "" {
			err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
			return
		} else if errCache = json.Unmarshal([]byte(strCache), &builderCache); errCache != nil {
			log.Error("json.Unmarshal err: ", err.Error())
			err = fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
			return
		}
		gracePeriodTime = builderCache.ExpirationGracePeriod
		auctionPeriodTime = builderCache.ExpirationAuctionPeriod
		deliverPeriodTime = builderCache.ExpirationDeliverPeriod
	} else {
		gracePeriodTime, err = builderConfigCell.ExpirationGracePeriod()
		if err != nil {
			err = fmt.Errorf("ExpirationGracePeriod err: %s", err.Error())
			return
		}
		auctionPeriodTime, err = builderConfigCell.ExpirationAuctionPeriod()
		if err != nil {
			err = fmt.Errorf("ExpirationAuctionPeriod err: %s", err.Error())
			return
		}
		deliverPeriodTime, err = builderConfigCell.ExpirationDeliverPeriod()
		if err != nil {
			err = fmt.Errorf("ExpirationDeliverPeriod err: %s", err.Error())
			return
		}
	}

	res = &AuctionConfig{
		GracePeriodTime:   gracePeriodTime,
		AuctionPeriodTime: auctionPeriodTime,
		DeliverPeriodTime: deliverPeriodTime,
	}
	return
}
