package dao

import "das_register_server/tables"

func (d *DbDao) GetTokenPriceList() (list []tables.TableTokenPriceInfo, err error) {
	err = d.parserDb.Where("token_id NOT IN('bsc_bep20_usdt','eth_erc20_usdt','tron_trc20_usdt','stripe_usd')").Order("id DESC").Find(&list).Error
	return
}
