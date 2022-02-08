package block_parser

import (
	"das_register_server/config"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/witness"
	"strings"
)

func (b *BlockParser) ActionConfirmProposal(req FuncTransactionHandleReq) (resp FuncTransactionHandleResp) {
	if isCV, err := isCurrentVersionTx(req.Tx, common.DasContractNameAccountCellType); err != nil {
		resp.Err = fmt.Errorf("isCurrentVersion err: %s", err.Error())
		return
	} else if !isCV {
		return
	}
	// propose
	builderPropose, err := witness.ProposalCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("ProposalCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	proposeHash := req.Tx.Inputs[builderPropose.Index].PreviousOutput.TxHash.Hex()

	list, err := b.DbDao.GetOrderTxByActionAndHashList(tables.TxActionPropose, []string{proposeHash})
	if err != nil {
		resp.Err = fmt.Errorf("GetOrderTxByActionAndHashList err: %s", err.Error())
		return
	}
	//
	var orderIds []string
	var orderTxList []tables.TableDasOrderTxInfo
	for _, v := range list {
		orderIds = append(orderIds, v.OrderId)
		orderTxList = append(orderTxList, tables.TableDasOrderTxInfo{
			OrderId:   v.OrderId,
			Action:    tables.TxActionConfirmProposal,
			Hash:      req.TxHash,
			Status:    tables.OrderTxStatusConfirm,
			Timestamp: int64(req.BlockTimestamp),
		})
	}

	// pre
	builderPreMap, err := witness.PreAccountIdCellDataBuilderFromTx(req.Tx, common.DataTypeOld)
	if err != nil {
		resp.Err = fmt.Errorf("PreAccountIdCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	// account
	builderAccMap, err := witness.AccountIdCellDataBuilderFromTx(req.Tx, common.DataTypeNew)
	if err != nil {
		resp.Err = fmt.Errorf("AccountIdCellDataBuilderFromTx err: %s", err.Error())
		return
	}
	var accountIds, accounts, inviterIds []string
	for k, v := range builderPreMap {
		accountIds = append(accountIds, k)
		accounts = append(accounts, v.Account)
		inviterId, _ := v.InviterId()
		inviterIds = append(inviterIds, inviterId)
	}
	// inviters
	inviters, err := b.DbDao.GetAccountInfoByAccountIds(inviterIds)
	if err != nil {
		resp.Err = fmt.Errorf("GetAccountInfoByAccountIds err: %s", err.Error())
		return
	}

	// update
	if err := b.DbDao.DoActionConfirmProposal(orderIds, accountIds, orderTxList); err != nil {
		resp.Err = fmt.Errorf("UpdateOrdersRegisterStatus err: %s", err.Error())
		return
	}
	// notify
	notify.SendLarkRegisterNotify(&notify.SendLarkRegisterNotifyParam{
		Action:  common.DasActionConfirmProposal,
		Account: strings.Join(accounts, ","),
		OrderId: "",
		Time:    req.BlockTimestamp,
		Hash:    req.TxHash,
	})
	// discord
	doDiscordNotify(inviters, builderPreMap, builderAccMap)

	return
}

func doDiscordNotify(inviters []tables.TableAccountInfo, builderPreMap map[string]*witness.PreAccountCellDataBuilder, builderAccMap map[string]*witness.AccountCellDataBuilder) {
	var inviterMap = make(map[string]tables.TableAccountInfo)
	for i, v := range inviters {
		inviterMap[v.AccountId] = inviters[i]
	}
	content := ""
	count := 0
	var contentList []string
	for k, v := range builderPreMap {
		account := v.Account
		invitedBy := ""
		inviterId, _ := v.InviterId()
		if acc, ok := inviterMap[inviterId]; ok {
			invitedBy = acc.Account
		}
		registerYears := uint64(1)
		if acc, ok := builderAccMap[k]; ok {
			registerYears = (acc.ExpiredAt - acc.RegisteredAt) / 31536000
		}
		content += fmt.Sprintf(`** %s ** registered for %d year(s), invited by %s
`, account, registerYears, invitedBy)
		count++
		if count == 15 {
			contentList = append(contentList, content)
			content = ""
			count = 0
		}
	}
	if content != "" {
		contentList = append(contentList, content)
	}
	go func() {
		log.Info("doDiscordNotify:", len(contentList))
		for _, v := range contentList {
			if err := notify.SendNotifyDiscord(config.Cfg.Notify.DiscordWebhook, v); err != nil {
				log.Error("SendNotifyDiscord err:", err.Error())
			}
		}
	}()
}