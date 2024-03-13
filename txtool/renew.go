package txtool

import (
	"das_register_server/config"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/sjatsh/uint128"
	"time"
)

func (t *TxTool) doOrderRenewTx() error {
	if !t.IsRebootTxOK() {
		return nil
	}
	list, err := t.DbDao.GetNeedSendPayOrderList(common.DasActionRenewAccount)
	if err != nil {
		return fmt.Errorf("GetNeedSendPayOrderList err: %s", err.Error())
	}
	for i, _ := range list {
		if err = t.DoOrderRenewTx(&list[i]); err != nil {
			return fmt.Errorf("DoOrderRenewTx err: %s", err.Error())
		}
	}
	return nil
}

func (t *TxTool) DoOrderRenewTx(order *tables.TableDasOrderInfo) error {
	if order == nil || order.Id == 0 {
		return fmt.Errorf("order is nil")
	}
	orderContent, err := order.GetContent()
	if err != nil {
		return fmt.Errorf("GetContent err: %s", err.Error())
	}

	accountId := common.GetAccountIdByAccount(order.Account)
	acc, err := t.DbDao.GetAccountInfoByAccountId(common.Bytes2Hex(accountId))
	if err != nil {
		return fmt.Errorf("GetAccountInfoByAccountId err: %s", err.Error())
	} else if acc.Status == tables.AccountStatusOnCross {
		log.Error("DoOrderRenewTx:", order.OrderId, acc.Status)
		msg := fmt.Sprintf(`order id: %s, account on cross`, order.OrderId)
		notify.SendLarkErrNotify(common.DasActionRenewAccount, msg)
		if err := t.DbDao.UpdateOrderStatusClosed(order.OrderId); err != nil {
			return fmt.Errorf("UpdateOrderStatusClosed err: %s", err.Error())
		}
		return nil
	}

	// build tx
	p := renewTxParams{
		order:      order,
		account:    &acc,
		renewYears: orderContent.RenewYears,
	}
	txParams, err := t.buildOrderRenewTx(&p)
	if err != nil {
		return fmt.Errorf("buildOrderPreRegisterTx err: %s", err.Error())
	}

	txBuilder := txbuilder.NewDasTxBuilderFromBase(t.TxBuilderBase, nil)
	if err := txBuilder.BuildTransaction(txParams); err != nil {
		return fmt.Errorf("BuildTransaction err: %s", err.Error())
	}

	sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()

	txFeeRate := config.Cfg.Server.TxTeeRate
	if txFeeRate == 0 {
		txFeeRate = 1
	}
	txFee := txFeeRate * sizeInBlock
	changeCapacity := txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity
	if sizeInBlock > 1e4 {
		changeCapacity = changeCapacity - txFee
		txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity = changeCapacity
	}
	log.Info("changeCapacity:", txFee, changeCapacity)

	// update order
	if err := t.DbDao.UpdatePayStatus(order.OrderId, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
	}
	if hash, err := txBuilder.SendTransaction(); err != nil {
		// update order
		if err := t.DbDao.UpdatePayStatus(order.OrderId, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Error("UpdatePayStatus err:", err.Error(), order.OrderId)
			notify.SendLarkErrNotify(common.DasActionRenewAccount, notify.GetLarkTextNotifyStr("UpdatePayStatus", order.OrderId, err.Error()))
		}
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		log.Info("SendTransaction ok:", tables.TxActionRenewAccount, hash)
		t.DasCache.AddCellInputByAction("", txBuilder.Transaction.Inputs)
		// update tx hash
		orderTx := tables.TableDasOrderTxInfo{
			OrderId:   order.OrderId,
			Action:    tables.TxActionRenewAccount,
			Hash:      hash.Hex(),
			Status:    tables.OrderTxStatusDefault,
			Timestamp: time.Now().UnixNano() / 1e6,
		}
		if err := t.DbDao.CreateOrderTx(&orderTx); err != nil {
			log.Error("CreateOrderTx err:", err.Error(), order.OrderId, hash.Hex())
			notify.SendLarkErrNotify(common.DasActionRenewAccount, notify.GetLarkTextNotifyStr("CreateOrderTx", order.OrderId, err.Error()))
		}
	}

	return nil
}

type renewTxParams struct {
	order      *tables.TableDasOrderInfo
	account    *tables.TableAccountInfo
	renewYears int
}

func (t *TxTool) buildOrderRenewTx(p *renewTxParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// config cell
	quoteCell, err := t.DasCore.GetQuoteCell()
	if err != nil {
		return nil, fmt.Errorf("GetQuoteCell err: %s", err.Error())
	}
	quote := quoteCell.Quote()
	priceBuilder, err := t.DasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsPrice)
	if err != nil {
		return nil, fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
	}

	// inputs
	accOutpoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutpoint,
	})

	// outputs
	accTx, err := t.DasCore.Client().GetTransaction(t.Ctx, accOutpoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	mapAcc, err := witness.AccountIdCellDataBuilderFromTx(accTx.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	}
	accBuilder, ok := mapAcc[p.account.AccountId]
	if !ok {
		return nil, fmt.Errorf("account map builder is nil [%s]", p.account.Outpoint)
	}

	//accountLength := common.GetAccountLength(p.order.Account)
	accountLength := uint8(accBuilder.AccountChars.Len())

	_, renewPrice, _ := priceBuilder.AccountPrice(accountLength)
	priceCapacity := uint128.From64(renewPrice).Mul(uint128.From64(common.OneCkb)).Div(uint128.From64(quote)).Big().Uint64()
	priceCapacity = priceCapacity * uint64(p.renewYears)
	log.Info("buildOrderRenewTx:", priceCapacity, renewPrice, p.renewYears, quote)

	// renew years
	newExpiredAt := int64(accBuilder.ExpiredAt) + int64(p.renewYears)*common.OneYearSec
	byteExpiredAt := molecule.Go64ToBytes(newExpiredAt)

	accWitness, accData, err := accBuilder.GenWitness(&witness.AccountCellParam{
		OldIndex: 0,
		NewIndex: 0,
		Action:   common.DasActionRenewAccount,
	})
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: accTx.Transaction.Outputs[accBuilder.Index].Capacity,
		Lock:     accTx.Transaction.Outputs[accBuilder.Index].Lock,
		Type:     accTx.Transaction.Outputs[accBuilder.Index].Type,
	})
	accData = append(accData, accTx.Transaction.OutputsData[accBuilder.Index][32:]...)
	accData1 := accData[:common.ExpireTimeEndIndex-common.ExpireTimeLen]
	accData2 := accData[common.ExpireTimeEndIndex:]
	newAccData := append(accData1, byteExpiredAt...)
	newAccData = append(newAccData, accData2...)
	txParams.OutputsData = append(txParams.OutputsData, newAccData) // change expired_at
	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionRenewAccount, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)
	txParams.Witnesses = append(txParams.Witnesses, accWitness)

	// income cell
	incomeCell, err := t.genIncomeCell(&genIncomeCellParam{priceCapacity: priceCapacity})
	if err != nil {
		return nil, fmt.Errorf("genIncomeCell err: %s", err.Error())
	}
	txParams.Outputs = append(txParams.Outputs, incomeCell.incomeCell)
	txParams.OutputsData = append(txParams.OutputsData, incomeCell.incomeCellData)
	txParams.Witnesses = append(txParams.Witnesses, incomeCell.incomeWitness)

	// search balance
	feeCapacity := uint64(1e4)
	needCapacity := feeCapacity + incomeCell.incomeCell.Capacity

	change, liveCell, err := t.DasCore.GetBalanceCellWithLock(&core.ParamGetBalanceCells{
		LockScript:        t.ServerScript,
		CapacityNeed:      needCapacity,
		DasCache:          t.DasCache,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCellWithLock err %s", err.Error())
	}

	// inputs
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// change
	if change > 0 {
		splitCkb := 2000 * common.OneCkb
		if config.Cfg.Server.SplitCkb > 0 {
			splitCkb = config.Cfg.Server.SplitCkb * common.OneCkb
		}
		changeList, err := core.SplitOutputCell2(change, splitCkb, 200, t.ServerScript, nil, indexer.SearchOrderAsc)
		if err != nil {
			return nil, fmt.Errorf("SplitOutputCell2 err: %s", err.Error())
		}
		for i := 0; i < len(changeList); i++ {
			txParams.Outputs = append(txParams.Outputs, changeList[i])
			txParams.OutputsData = append(txParams.OutputsData, []byte{})
		}

		//changeList, err := core.SplitOutputCell(change, 1000*common.OneCkb, 3, t.ServerScript, nil)
		//if err != nil {
		//	return nil, fmt.Errorf("SplitOutputCell err: %s", err.Error())
		//}
		//for i := len(changeList); i > 0; i-- {
		//	txParams.Outputs = append(txParams.Outputs, changeList[i-1])
		//	txParams.OutputsData = append(txParams.OutputsData, []byte{})
		//}
		//for _, cell := range changeList {
		//	txParams.Outputs = append(txParams.Outputs, cell)
		//	txParams.OutputsData = append(txParams.OutputsData, []byte{})
		//}
	}

	// cell deps
	dasLockContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	accContract, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	timeCell, err := t.DasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}
	heightCell, err := t.DasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}
	accountConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	priceConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsPrice)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	incomeConfig, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsIncome)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}
	txParams.CellDeps = append(txParams.CellDeps,
		dasLockContract.ToCellDep(),
		accContract.ToCellDep(),
		incomeContract.ToCellDep(),
		timeCell.ToCellDep(),
		heightCell.ToCellDep(),
		quoteCell.ToCellDep(),
		accountConfig.ToCellDep(),
		priceConfig.ToCellDep(),
		incomeConfig.ToCellDep(),
	)

	return &txParams, nil
}

type genIncomeCellParam struct {
	priceCapacity uint64
}
type genIncomeCellRes struct {
	incomeCell     *types.CellOutput
	incomeCellData []byte
	incomeWitness  []byte
}

func (t *TxTool) genIncomeCell(p *genIncomeCellParam) (*genIncomeCellRes, error) {
	var res genIncomeCellRes
	builder, err := t.DasCore.ConfigCellDataBuilderByTypeArgsList(common.ConfigCellTypeArgsIncome)
	if err != nil {
		return nil, fmt.Errorf("ConfigCellDataBuilderByTypeArgsList err: %s", err.Error())
	}
	incomeCellBaseCapacity, err := builder.IncomeBasicCapacity()
	if err != nil {
		return nil, fmt.Errorf("IncomeBasicCapacity err: %s", err.Error())
	}
	log.Info("IncomeBasicCapacity:", incomeCellBaseCapacity, p.priceCapacity)

	incomeCellCapacity := p.priceCapacity
	creator := molecule.ScriptDefault()
	var lockList []*types.Script
	var incomeCapacities []uint64
	if p.priceCapacity < incomeCellBaseCapacity {
		incomeCellCapacity = incomeCellBaseCapacity
		creator = molecule.CkbScript2MoleculeScript(t.ServerScript)
		lockList = append(lockList, t.ServerScript)
		diff := incomeCellBaseCapacity - p.priceCapacity
		incomeCapacities = append(incomeCapacities, diff)
	}
	asContract, err := core.GetDasContractInfo(common.DasContractNameAlwaysSuccess)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	incomeContract, err := core.GetDasContractInfo(common.DasContractNameIncomeCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	res.incomeCell = &types.CellOutput{
		Capacity: incomeCellCapacity,
		Lock:     asContract.ToScript(nil),
		Type:     incomeContract.ToScript(nil),
	}
	dasLock := t.DasCore.GetDasLock()
	lockList = append(lockList, dasLock)
	incomeCapacities = append(incomeCapacities, p.priceCapacity)

	res.incomeWitness, res.incomeCellData, _ = witness.CreateIncomeCellWitness(&witness.NewIncomeCellParam{
		Creator:     &creator,
		BelongTos:   lockList,
		Capacities:  incomeCapacities,
		OutputIndex: 1,
	})
	return &res, nil
}
