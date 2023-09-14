package txtool

import "time"

func (t *TxTool) IsRebootTxOK() bool {
	rebootSeconds := time.Since(t.RebootTime).Seconds()
	if rebootSeconds > 60 {
		return true
	}
	log.Debug("IsRebootTxOK:", rebootSeconds)
	return false
}
