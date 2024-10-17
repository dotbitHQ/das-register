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
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"time"
)

func (t *TxTool) doOrderApplyTx() error {
	if !t.IsRebootTxOK() {
		return nil
	}
	list, err := t.DbDao.GetNeedSendPayOrderList(common.DasActionApplyRegister)
	if err != nil {
		return fmt.Errorf("GetNeedSendPayOrderList err: %s", err.Error())
	}
	for i, _ := range list {
		if err = t.DoOrderApplyTx(&list[i]); err != nil {
			return fmt.Errorf("DoOrderApplyTx err: %s", err.Error())
		}
	}
	return nil
}

func (t *TxTool) DoOrderApplyTx(order *tables.TableDasOrderInfo) error {
	if order == nil || order.Id == 0 {
		return fmt.Errorf("order is nil")
	}
	p := applyTxParams{
		order: order,
	}
	txParams, err := t.buildOrderApplyTx(&p)
	if err != nil {
		return fmt.Errorf("buildOrderApplyTx err: %s", err.Error())
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
	if txFee > 1e4 {
		changeCapacity = changeCapacity - txFee
		txBuilder.Transaction.Outputs[len(txBuilder.Transaction.Outputs)-1].Capacity = changeCapacity
	}
	log.Info("changeCapacity:", txFee, changeCapacity)

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
	if err := t.DbDao.UpdatePayStatus(order.OrderId, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
	}

	log.Info(txBuilder.TxString())
	if hash, err := txBuilder.SendTransaction(); err != nil {
		// update order
		if err := t.DbDao.UpdatePayStatus(order.OrderId, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Error("UpdatePayStatus err:", err.Error(), order.OrderId)
			notify.SendLarkErrNotify(common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("UpdatePayStatus", order.OrderId, err.Error()))
		}
		return fmt.Errorf("SendTransaction err: %s", err.Error())
	} else {
		log.Info("SendTransaction ok:", tables.TxActionApplyRegister, hash)
		t.DasCache.AddCellInputByAction("", txBuilder.Transaction.Inputs)
		// update tx hash
		orderTx := tables.TableDasOrderTxInfo{
			OrderId:   order.OrderId,
			Action:    tables.TxActionApplyRegister,
			Hash:      hash.Hex(),
			Status:    tables.OrderTxStatusDefault,
			Timestamp: time.Now().UnixNano() / 1e6,
		}
		if err := t.DbDao.CreateOrderTx(&orderTx); err != nil {
			log.Error("CreateOrderTx err:", err.Error(), order.OrderId, hash.Hex())
			notify.SendLarkErrNotify(common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("CreateOrderTx", order.OrderId, err.Error()))
		}
	}
	return nil
}

type applyTxParams struct {
	order *tables.TableDasOrderInfo
}

func (t *TxTool) buildOrderApplyTx(p *applyTxParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// height cell, time cell
	heightCell, err := t.DasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}
	timeCell, err := t.DasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}

	// outputs
	applyContract, err := core.GetDasContractInfo(common.DasContractNameApplyRegisterCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	var applyData []byte
	applyData = append(applyData, t.ServerScript.Args...)
	applyData = append(applyData, []byte(p.order.Account)...)
	applyData, _ = blake2b.Blake256(applyData)
	txParams.OutputsData = append(txParams.OutputsData, applyData)

	applyOutputs := &types.CellOutput{
		Lock: t.ServerScript,
		Type: applyContract.ToScript(nil),
	}
	applyOutputs.Capacity = applyOutputs.OccupiedCapacity(applyData) * common.OneCkb
	txParams.Outputs = append(txParams.Outputs, applyOutputs)

	// search balance
	feeCapacity := uint64(1e4)
	needCapacity := feeCapacity + applyOutputs.Capacity

	change, liveCell, err := t.DasCore.GetBalanceCellWithLock(&core.ParamGetBalanceCells{
		LockScript:   t.ServerScript,
		CapacityNeed: needCapacity,
		DasCache:     t.DasCache,
		SearchOrder:  indexer.SearchOrderDesc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCellWithLock err %s", err.Error())
	}

	if change > 0 {
		splitCkb := 2000 * common.OneCkb
		if config.Cfg.Server.SplitCkb > 0 {
			splitCkb = config.Cfg.Server.SplitCkb * common.OneCkb
		}
		changeList, err := core.SplitOutputCell2(change, splitCkb, 200, t.ServerScript, nil, indexer.SearchOrderDesc)
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
	}

	// inputs
	for _, v := range liveCell {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			PreviousOutput: v.OutPoint,
		})
	}

	// witness
	actionWitness, err := witness.GenActionDataWitness(common.DasActionApplyRegister, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// cell deps
	txParams.CellDeps = append(txParams.CellDeps,
		applyContract.ToCellDep(),
		heightCell.ToCellDep(),
		timeCell.ToCellDep(),
	)

	return &txParams, nil
}
