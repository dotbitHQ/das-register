package dao

import (
	"das_register_server/tables"
	"github.com/dotbitHQ/das-lib/common"
)

func (d *DbDao) CreatePending(pending *tables.TableRegisterPendingInfo) error {
	return d.db.Create(&pending).Error
}

func (d *DbDao) GetPendingList(limit int) (list []tables.TableRegisterPendingInfo, err error) {
	err = d.db.Where(" block_number=0 AND status=0 ").Order(" id ").Limit(limit).Find(&list).Error
	return
}

func (d *DbDao) UpdatePendingToConfirm(id, blockNumber, blockTimestamp uint64) error {
	return d.db.Model(tables.TableRegisterPendingInfo{Id: id}).Updates(map[string]interface{}{
		"block_number":    blockNumber,
		"status":          tables.StatusConfirm,
		"block_timestamp": blockTimestamp,
	}).Error
}

func (d *DbDao) UpdatePendingToRejected(timestamp int64) error {
	return d.db.Model(tables.TableRegisterPendingInfo{}).
		Where(" block_number=0 AND status=0 AND block_timestamp<? ", timestamp).
		Updates(map[string]interface{}{
			"status": tables.StatusRejected,
		}).Error
}

func (d *DbDao) GetPendingStatus(chainType common.ChainType, address string, actions []common.DasAction) (tx tables.TableRegisterPendingInfo, err error) {
	err = d.db.Where(" chain_type=? AND address=? AND block_number=0 AND action IN(?) AND status=0 ", chainType, address, actions).Order(" id DESC ").First(&tx).Error
	return
}

func (d *DbDao) SearchMaybeRejectedPending(timestamp int64) (list []tables.TableRegisterPendingInfo, err error) {
	err = d.db.Where(" block_number=0 AND `status`=0 AND block_timestamp<? ", timestamp).Limit(100).Find(&list).Error
	return
}

func (d *DbDao) UpdatePendingStatusToConfirm(action, outpoint string, blockNumber, blockTimestamp uint64) error {
	return d.db.Model(tables.TableRegisterPendingInfo{}).
		Where("action=? AND outpoint=?", action, outpoint).
		Updates(map[string]interface{}{
			"block_number":    blockNumber,
			"block_timestamp": blockTimestamp,
			"status":          tables.StatusConfirm,
		}).Error
}
