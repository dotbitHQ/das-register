package dao

import "das_register_server/tables"

type InviterNum struct {
	InviterId string `json:"inviter_id" gorm:"column:inviter_id"`
	Num       int    `json:"num" gorm:"column:num"`
}

func (d *DbDao) GetInviterMoreThan10() (list []InviterNum, err error) {
	sql := `SELECT inviter_id,count(*) num FROM t_rebate_info WHERE inviter_id NOT IN('','0x0','0x0000000000000000000000000000000000000000') GROUP BY inviter_id HAVING num>=10`
	err = d.parserDb.Raw(sql).Find(&list).Error
	return
}

func (d *DbDao) GetNameDaoMember(accountId string) (list []tables.TableAccountInfo, err error) {
	err = d.parserDb.Where("parent_account_id=?", &accountId).Select("account").Find(&list).Error
	return
}

func (d *DbDao) GetInviteeList(inviterId string) (list []tables.TableRebateInfo, err error) {
	err = d.parserDb.
		Where("inviter_id=? AND reward_type=0 AND service_type=1 AND block_timestamp>=1658332800000", inviterId).
		Group("invitee_address").
		Select("invitee_id").
		Find(&list).Error
	return
}

func (d *DbDao) GetInviterCount(accountIds []string, charsetNum int) (count int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).
		Where("account_id IN(?) AND charset_num=?",
			accountIds, charsetNum).Count(&count).Error
	return
}
