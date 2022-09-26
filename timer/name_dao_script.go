package timer

import (
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/notify"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/robfig/cron/v3"
	"strings"
)

type NameDaoTimer struct {
	DbDao *dao.DbDao
	cron  *cron.Cron
}

func (t *NameDaoTimer) CheckNameDaoMember(account string, charset int) error {
	list, err := t.DbDao.GetInviterMoreThan10()
	if err != nil {
		return fmt.Errorf("GetInviterMoreThan10 err: %s", err.Error())
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(account))
	memberList, err := t.DbDao.GetNameDaoMember(accountId)
	if err != nil {
		return fmt.Errorf("GetNameDaoMember err: %s", err.Error())
	}
	var memberMap = make(map[string]struct{})
	for _, v := range memberList {
		acc := strings.Replace(v.Account, account, "bit", -1)
		accId := common.Bytes2Hex(common.GetAccountIdByAccount(acc))
		memberMap[accId] = struct{}{}
	}

	var memberIds []string
	for _, v := range list {
		if _, ok := memberMap[v.InviterId]; ok {
			continue
		}
		inviteeList, err := t.DbDao.GetInviteeList(v.InviterId)
		if err != nil {
			return fmt.Errorf("GetInviteeList err: %s", err.Error())
		}
		if len(inviteeList) < 10 {
			continue
		}
		var invitees []string
		for _, ee := range inviteeList {
			invitees = append(invitees, ee.InviteeId)
		}
		count, err := t.DbDao.GetInviterCount(invitees, charset)
		if err != nil {
			return fmt.Errorf("GetInviterCount err: %s", err.Error())
		}
		if count >= 10 {
			memberIds = append(memberIds, v.InviterId)
		}
	}

	members, err := t.DbDao.GetAccountInfoByAccountIds(memberIds)
	if err != nil {
		return fmt.Errorf("GetAccountInfoByAccountIds err: %s", err.Error())
	}
	log.Info("CheckNameDaoMember:", len(members))
	if len(members) > 0 {
		msg := ``
		for _, v := range members {
			log.Info("CheckNameDaoMember:", v.Account, v.AccountId)
			msg += fmt.Sprintf("%s %s\n", v.Account, v.AccountId)
		}
		notify.SendLarkTextNotifyAtAll(config.Cfg.Notify.LarkErrorKey, "NameDao UnMint SubAccount List", msg)
	}

	return nil
}

func (t *NameDaoTimer) RunCheckNameDaoMember() {
	if config.Cfg.Notify.LarkErrorKey == "" {
		return
	}
	t.cron = cron.New(cron.WithSeconds())
	spec := "0 0 2 * * ? "
	_, err := t.cron.AddFunc(spec, func() {
		if err := t.CheckNameDaoMember("0x.bit", 2); err != nil {
			log.Error("CheckNameDaoMember err: %s", err.Error())
		}
	})
	if err != nil {
		log.Error("RunCheckNameDaoMember err: %s", err.Error())
	} else {
		t.cron.Start()
	}
}

func (t *NameDaoTimer) CloseCron() {
	if t.cron != nil {
		t.cron.Stop()
	}
	log.Warn("cron done")
}
