package txtool

import (
	"das_register_server/config"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"strings"
	"time"
)

func (t *TxTool) doOrderPreRegisterTx() error {
	list, err := t.DbDao.GetNeedSendPreRegisterTxOrderList()
	if err != nil {
		return fmt.Errorf("GetNeedSendPreRegisterTxOrderList err: %s", err.Error())
	}
	for i, _ := range list {
		if err = t.DoOrderPreRegisterTx(&list[i]); err != nil {
			return fmt.Errorf("DoOrderPreRegisterTx err: %s", err.Error())
		}
	}
	return nil
}

func (t *TxTool) DoOrderPreRegisterTx(order *tables.TableDasOrderInfo) error {
	if order == nil || order.Id == 0 {
		return fmt.Errorf("order is nil")
	}
	orderContent, err := order.GetContent()
	if err != nil {
		return fmt.Errorf("GetContent err: %s", err.Error())
	}
	orderTxApply, err := t.DbDao.GetOrderTxByAction(order.OrderId, tables.TxActionApplyRegister)
	if err != nil {
		return fmt.Errorf("GetOrderTxByAction err: %s", err.Error())
	} else if orderTxApply.Id == 0 {
		return fmt.Errorf("order apply register tx is nil: %s", order.OrderId)
	}
	// inviter channel
	inviterScript, channelScript, inviterId, err := t.getOrderInviterChannelScript(&orderContent)
	if err != nil {
		return fmt.Errorf("getOrderInviterChannelScript err: %s", err.Error())
	}
	// owner lock args
	ownerLockScript, _, err := t.DasCore.Daf().HexToScript(core.DasAddressHex{
		DasAlgorithmId: order.ChainType.ToDasAlgorithmId(true),
		AddressHex:     order.Address,
		IsMulti:        true,
		ChainType:      order.ChainType,
	})
	if err != nil {
		return fmt.Errorf("NormalToScript err: %s", err.Error())
	}
	p := preRegisterTxParams{
		order:         order,
		applyCellHash: orderTxApply.Hash,
		inviterId:     inviterId,
		inviterScript: inviterScript,
		channelScript: channelScript,
		ownerLockArgs: ownerLockScript.Args,
		refundLock:    t.ServerScript,
		accountChars:  orderContent.AccountCharStr,
		registerYears: orderContent.RegisterYears,
	}
	txParams, err := t.buildOrderPreRegisterTx(&p)
	if err != nil {
		return fmt.Errorf("buildOrderPreRegisterTx err: %s", err.Error())
	}

	txBuilder := txbuilder.NewDasTxBuilderFromBase(t.TxBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return fmt.Errorf("BuildTransaction err: %s", err.Error())
	}

	// check has pre tx
	preOrder, err := t.DbDao.GetPreRegisteredOrderByAccountId(order.AccountId)
	if err != nil {
		return fmt.Errorf("GetPreRegisteredOrderByAccountId err: %s", err.Error())
	} else if preOrder.Id > 0 && time.Now().Unix() < (preOrder.Timestamp/1e3)+2592000 { // refund
		log.Info("UpdateOrderToRefund:", order.OrderId)
		if err := t.DbDao.UpdateOrderToRefund(order.OrderId); err != nil {
			return fmt.Errorf("UpdateOrderToRefund err: %s [%s]", err.Error(), order.OrderId)
		}
		return nil
	}

	// update order
	if err := t.DbDao.UpdatePreRegisterStatus(order.OrderId, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return fmt.Errorf("UpdatePreRegisterStatus err: %s", err.Error())
	}
	//
	if hash, err := txBuilder.SendTransaction(); err != nil {
		// update order
		if err := t.DbDao.UpdatePreRegisterStatus(order.OrderId, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Error("UpdatePayStatus err:", err.Error(), order.OrderId)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionPreRegister, notify.GetLarkTextNotifyStr("UpdatePayStatus", order.OrderId, err.Error()))
		}
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		log.Info("SendTransaction ok:", tables.TxActionPreRegister, hash)
		t.DasCache.AddCellInputByAction("", txBuilder.Transaction.Inputs)
		// update tx hash
		orderTx := tables.TableDasOrderTxInfo{
			OrderId:   order.OrderId,
			Action:    tables.TxActionPreRegister,
			Hash:      hash.Hex(),
			Status:    tables.OrderTxStatusDefault,
			Timestamp: time.Now().UnixNano() / 1e6,
		}
		if err := t.DbDao.CreateOrderTx(&orderTx); err != nil {
			log.Error("CreateOrderTx err:", err.Error(), order.OrderId, hash.Hex())
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionPreRegister, notify.GetLarkTextNotifyStr("CreateOrderTx", order.OrderId, err.Error()))
		}
	}

	return nil
}

func (t *TxTool) getAccountScript(accountId []byte) (*types.Script, error) {
	acc, err := t.DbDao.GetAccountInfoByAccountId(common.Bytes2Hex(accountId))
	if err != nil {
		return nil, fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
	} else if acc.Id > 0 {
		ownerLockScript, _, err := t.DasCore.Daf().HexToScript(core.DasAddressHex{
			DasAlgorithmId: acc.OwnerChainType.ToDasAlgorithmId(true),
			AddressHex:     acc.Owner,
			IsMulti:        true,
			ChainType:      acc.OwnerChainType,
		})
		if err != nil {
			return nil, fmt.Errorf("FormatAddressToDasLockScript err: %s", err.Error())
		}
		return ownerLockScript, nil
	}
	return nil, nil
}

func (t *TxTool) getOrderInviterChannelScript(orderContent *tables.TableOrderContent) (*types.Script, *types.Script, []byte, error) {
	var inviterScript, channelScript *types.Script
	var inviterId []byte
	if orderContent == nil {
		return nil, nil, nil, fmt.Errorf("order content is nil")
	}
	if orderContent.InviterAccount != "" {
		inviterId = common.GetAccountIdByAccount(orderContent.InviterAccount)
		script, err := t.getAccountScript(inviterId)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("getAccountScript[%s] err: %s", orderContent.InviterAccount, err.Error())
		} else {
			inviterScript = script
		}
	}
	if orderContent.ChannelAccount != "" {
		channelId := common.GetAccountIdByAccount(orderContent.ChannelAccount)
		script, err := t.getAccountScript(channelId)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("getAccountScript[%s] err: %s", orderContent.InviterAccount, err.Error())
		} else if script != nil {
			channelScript = script
		} else {
			channelScript = t.ServerScript
		}
	} else if config.Cfg.Server.PayServerAddress != "" {
		if parseAddr, err := address.Parse(config.Cfg.Server.PayServerAddress); err != nil {
			log.Error("address.Parse err: ", err.Error(), config.Cfg.Server.PayServerAddress)
		} else {
			channelScript = parseAddr.Script
		}
	}
	return inviterScript, channelScript, inviterId, nil
}

type preRegisterTxParams struct {
	order         *tables.TableDasOrderInfo
	applyCellHash string
	inviterId     []byte
	inviterScript *types.Script
	channelScript *types.Script
	ownerLockArgs []byte
	refundLock    *types.Script
	accountChars  []common.AccountCharSet
	registerYears int
}

func (t *TxTool) buildOrderPreRegisterTx(p *preRegisterTxParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs
	applyHash := types.HexToHash(p.applyCellHash)
	applyTx, err := t.DasCore.Client().GetTransaction(t.Ctx, applyHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", applyHash.String())
	}
	applyCapacity := applyTx.Transaction.Outputs[0].Capacity
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: &types.OutPoint{
			TxHash: applyHash,
			Index:  0,
		},
	})

	// time cell
	timeCell, err := t.DasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	quoteCell, err := t.DasCore.GetQuoteCell()
	if err != nil {
		return nil, fmt.Errorf("GetQuoteCell err: %s", err.Error())
	}
	quote := quoteCell.Quote()
	// config cell
	priceBuilder, err := t.DasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsPrice, common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
	}
	invitedDiscount := uint32(0)
	if p.inviterScript != nil {
		if invitedDiscount, err = priceBuilder.PriceInvitedDiscount(); err != nil {
			return nil, fmt.Errorf("PriceInvitedDiscount err: %s", err.Error())
		}
	}
	accountLength := common.GetAccountLength(p.order.Account)
	price := priceBuilder.PriceConfig(accountLength)
	if price == nil {
		return nil, fmt.Errorf("PriceConfig is nil")
	}
	newPrice, _, _ := priceBuilder.AccountPrice(accountLength)
	priceCapacity := (newPrice / quote) * common.OneCkb
	if invitedDiscount > 0 {
		priceCapacity = (priceCapacity / common.PercentRateBase) * (common.PercentRateBase - uint64(invitedDiscount))
	}
	priceCapacity = priceCapacity * uint64(p.registerYears)
	log.Info("buildOrderPreRegisterTx:", priceCapacity, newPrice, p.registerYears, quote, invitedDiscount)
	// basicCapacity
	basicCapacity, _ := priceBuilder.BasicCapacityFromOwnerDasAlgorithmId(common.Bytes2Hex(p.ownerLockArgs))
	preparedFeeCapacity, _ := priceBuilder.PreparedFeeCapacity()
	basicCapacity = basicCapacity + preparedFeeCapacity + uint64(len([]byte(p.order.Account)))*common.OneCkb
	log.Info("pre capacity:", basicCapacity, priceCapacity)
	accountId := common.GetAccountIdByAccount(p.order.Account)
	accountChars := tables.AccountCharSetListToMoleculeAccountChars(p.accountChars)

	// char type
	var accountCharTypeMap = make(map[common.AccountCharType]struct{})
	common.GetAccountCharType(accountCharTypeMap, p.accountChars)

	// witness
	var preBuilder witness.PreAccountCellDataBuilder
	preWitness, preData, err := preBuilder.GenWitness(&witness.PreAccountCellParam{
		NewIndex:        0,
		Action:          common.DasActionPreRegister,
		CreatedAt:       timeCell.Timestamp(),
		InvitedDiscount: invitedDiscount,
		Quote:           quoteCell.Quote(),
		InviterScript:   p.inviterScript,
		ChannelScript:   p.channelScript,
		InviterId:       p.inviterId,
		OwnerLockArgs:   p.ownerLockArgs,
		RefundLock:      p.refundLock,
		Price:           *price,
		AccountChars:    accountChars,
	})
	if err != nil {
		return nil, fmt.Errorf("GenWitness err: %s", err.Error())
	}
	actionWitness, err := witness.GenActionDataWitness(common.DasActionPreRegister, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)
	txParams.Witnesses = append(txParams.Witnesses, preWitness)

	// outputs
	preContract, err := core.GetDasContractInfo(common.DasContractNamePreAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	alwaysContract, err := core.GetDasContractInfo(common.DasContractNameAlwaysSuccess)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	preOutputs := &types.CellOutput{
		Lock: alwaysContract.ToScript(nil),
		Type: preContract.ToScript(nil),
	}
	preData = append(preData, accountId...)
	txParams.OutputsData = append(txParams.OutputsData, preData)

	preOutputs.Capacity = basicCapacity + priceCapacity
	txParams.Outputs = append(txParams.Outputs, preOutputs)

	// search balance
	feeCapacity := uint64(1112663)
	needCapacity := feeCapacity + preOutputs.Capacity - applyCapacity

	liveCell, totalCapacity, err := t.DasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          t.DasCache,
		LockScript:        t.ServerScript,
		CapacityNeed:      needCapacity,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	if change := totalCapacity - needCapacity; change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     t.ServerScript,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	// inputs
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// cell deps
	applyContract, err := core.GetDasContractInfo(common.DasContractNameApplyRegisterCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	heightCell, err := t.DasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}
	applyConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsApply)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	accountConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	emojiConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsCharSetEmoji)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	digitConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsCharSetDigit)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	enConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsCharSetEn)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	releaseConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsRelease)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	unavailableConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsUnavailable)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	priceConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsPrice)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	//
	bys, err := blake2b.Blake160([]byte(strings.TrimSuffix(p.order.Account, common.DasAccountSuffix)))
	if err != nil {
		return nil, fmt.Errorf("blake2b.Blake160 err: %s", err.Error())
	}
	accountHashIndex := uint32(bys[0] % 20)
	PreservedAccountConfig, err := core.GetDasConfigCellInfo(common.GetConfigCellTypeArgsPreservedAccountByIndex(accountHashIndex))
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		alwaysContract.ToCellDep(),
		applyContract.ToCellDep(),
		preContract.ToCellDep(),
		timeCell.ToCellDep(),
		heightCell.ToCellDep(),
		quoteCell.ToCellDep(),
		priceConfig.ToCellDep(),
		applyConfig.ToCellDep(),
		accountConfig.ToCellDep(),
		//emojiConfig.ToCellDep(),
		//digitConfig.ToCellDep(),
		//enConfig.ToCellDep(),
		releaseConfig.ToCellDep(),
		unavailableConfig.ToCellDep(),
		PreservedAccountConfig.ToCellDep(),
	)
	for k, _ := range accountCharTypeMap {
		switch k {
		case common.AccountCharTypeEmoji:
			txParams.CellDeps = append(txParams.CellDeps, emojiConfig.ToCellDep())
		case common.AccountCharTypeNumber:
			txParams.CellDeps = append(txParams.CellDeps, digitConfig.ToCellDep())
		case common.AccountCharTypeEn:
			txParams.CellDeps = append(txParams.CellDeps, enConfig.ToCellDep())
		}
	}

	return &txParams, nil
}
