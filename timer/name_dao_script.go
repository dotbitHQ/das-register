package timer

import (
	"das_register_server/dao"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"strings"
)

type NameDaoTimer struct {
	DbDao *dao.DbDao
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

	fmt.Println(memberIds)
	members, err := t.DbDao.GetAccountInfoByAccountIds(memberIds)
	if err != nil {
		return fmt.Errorf("GetAccountInfoByAccountIds err: %s", err.Error())
	}
	for _, v := range members {
		fmt.Println(v.Account, v.AccountId)
	}

	return nil
}
