package block_parser

import (
	"das_register_server/config"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/witness"
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
		inviterId := v.InviterId
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
	b.doLarkNotify(inviters, builderPreMap, builderAccMap)

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
		inviterId := v.InviterId
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
	//go func() {
	//	for _, v := range contentList {
	//		tmp := strings.Replace(v, "** ", "", -1)
	//		tmp = strings.Replace(tmp, " **", "", -1)
	//		notify.SendLarkTextNotify(config.Cfg.Notify.LarkRegisterOkKey, "", tmp)
	//	}
	//}()
}

func (b *BlockParser) doLarkNotify(inviters []tables.TableAccountInfo, builderPreMap map[string]*witness.PreAccountCellDataBuilder, builderAccMap map[string]*witness.AccountCellDataBuilder) {
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
		inviterId := v.InviterId
		if acc, ok := inviterMap[inviterId]; ok {
			invitedBy = acc.Account
		}
		registerYears := uint64(1)
		if acc, ok := builderAccMap[k]; ok {
			registerYears = (acc.ExpiredAt - acc.RegisteredAt) / 31536000
		}
		ownerNormal, _, _ := b.DasCore.Daf().ArgsToNormal(common.Hex2Bytes(v.OwnerLockArgs))
		owner := ownerNormal.AddressNormal
		if len(owner) > 4 {
			owner = owner[len(owner)-4:]
		}
		content += fmt.Sprintf(`%s, %d, %4s, %s
`, account, registerYears, owner, invitedBy)
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
		for _, v := range contentList {
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkRegisterOkKey, "", v)
		}
	}()
}
