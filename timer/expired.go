package timer

import "fmt"

func (t *TxTimer) checkExpired() error {
	list, err := t.dbDao.GetNeedExpiredOrders()
	if err != nil {
		return fmt.Errorf("GetNeedExpiredOrders err: %s", err.Error())
	}
	for _, v := range list {
		log.Info("DoExpiredOrder:", v.OrderId)
		if err := t.dbDao.DoExpiredOrder(v.OrderId); err != nil {
			return fmt.Errorf("DoExpiredOrder err: %s", err.Error())
		}
	}
	return nil
}
