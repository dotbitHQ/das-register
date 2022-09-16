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
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"time"
)

func (t *TxTool) doOrderApplyTx() error {
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
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("UpdatePayStatus", order.OrderId, err.Error()))
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
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, common.DasActionApplyRegister, notify.GetLarkTextNotifyStr("CreateOrderTx", order.OrderId, err.Error()))
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
	applyData = append(applyData, molecule.Go64ToBytes(heightCell.BlockNumber())...)
	applyData = append(applyData, molecule.Go64ToBytes(timeCell.Timestamp())...)
	txParams.OutputsData = append(txParams.OutputsData, applyData)

	applyOutputs := &types.CellOutput{
		Lock: t.ServerScript,
		Type: applyContract.ToScript(nil),
	}
	applyOutputs.Capacity = applyOutputs.OccupiedCapacity(applyData) * common.OneCkb
	txParams.Outputs = append(txParams.Outputs, applyOutputs)

	// search balance
	feeCapacity := uint64(1e4)
	splitCapacity := 3000 * common.OneCkb
	needCapacity := feeCapacity + applyOutputs.Capacity
	liveCell, totalCapacity, err := t.DasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          t.DasCache,
		LockScript:        t.ServerScript,
		CapacityNeed:      needCapacity + splitCapacity,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		return nil, fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	if change := totalCapacity - needCapacity; change > 0 {
		changeList, err := core.SplitOutputCell(change, 1000*common.OneCkb, 3, t.ServerScript, nil)
		if err != nil {
			return nil, fmt.Errorf("SplitOutputCell err: %s", err.Error())
		}
		for i := len(changeList); i > 0; i-- {
			txParams.Outputs = append(txParams.Outputs, changeList[i-1])
			txParams.OutputsData = append(txParams.OutputsData, []byte{})
		}
		//for _, cell := range changeList {
		//	txParams.Outputs = append(txParams.Outputs, cell)
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
