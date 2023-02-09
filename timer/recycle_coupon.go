package timer

import "das_register_server/tables"

func (t *TxTimer) DoResetCoupon() error {
	res, err := t.dbDao.GetUsedCouponNotChecked()
	if err != nil {
		return err
	}
	log.Info("doResetCoupon start:", res)
	var failedCouponOrderId []string
	var successCouponOrderId []string
	for i, _ := range res {
		if res[i].OrderStatus == tables.OrderStatusClosed && res[i].RegisterStatus != tables.RegisterStatusRegistered {
			failedCouponOrderId = append(failedCouponOrderId, res[i].OrderId)
		} else if res[i].OrderStatus == tables.OrderStatusClosed && res[i].RegisterStatus == tables.RegisterStatusRegistered {
			successCouponOrderId = append(successCouponOrderId, res[i].OrderId)
		} else {

		}
	}

	if len(failedCouponOrderId) != 0 {
		if err := t.dbDao.ResetCoupon(failedCouponOrderId); err != nil {
			return err
		}
		log.Info("doResetCoupon failedCoupon: ", failedCouponOrderId)
	}

	if len(successCouponOrderId) != 0 {
		if err := t.dbDao.MarkChecked(successCouponOrderId); err != nil {
			return err
		}
		log.Info("doResetCoupon successCoupon: ", successCouponOrderId)
	}
	return nil
}
