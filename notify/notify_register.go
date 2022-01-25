package notify

import (
	"das_register_server/config"
	"fmt"
	"time"
)

type SendLarkRegisterNotifyParam struct {
	Action  string
	Account string
	OrderId string
	Time    uint64
	Hash    string
}

func SendLarkRegisterNotify(p *SendLarkRegisterNotifyParam) {
	if p == nil {
		return
	}
	msg := `> account: %s
> order_id: %s
> time: %s
> hash: %s`
	t := time.Unix(int64(p.Time/1000), 0)

	msg = fmt.Sprintf(msg, p.Account, p.OrderId, t.Format("2006-01-02 15:04:05"), p.Hash)
	SendLarkTextNotify(config.Cfg.Notify.LarkRegisterKey, p.Action, msg)
}
