package reverse_smt

import (
	"das_register_server/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/smt"
	"strconv"
)

const (
	SmtNamePrefix = "reverse_record"
	SmtNum        = 256
)

func GetReverseSmt(addresses ...string) *smt.SmtServer {
	if len(addresses) == 0 || addresses[0] == "" {
		return smt.NewSmtSrv(config.Cfg.Server.SmtServer, SmtNamePrefix)
	}
	address := addresses[0]
	splitStr := string(common.Hex2Bytes(address)[:2])
	splitNum, _ := strconv.ParseInt(splitStr, 16, 64)
	smtName := fmt.Sprintf("%s_%d", SmtNamePrefix, splitNum%SmtNum)
	tree := smt.NewSmtSrv(config.Cfg.Server.SmtServer, smtName)
	return tree
}
