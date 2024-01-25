package elastic

import (
	"das_register_server/config"
	"fmt"
	"github.com/olivere/elastic/v7"
)

type Es struct {
	EsCli *elastic.Client
}

func InitEs() (es *Es, err error) {
	addr := config.Cfg.ES.Addr
	user := config.Cfg.ES.User
	pwd := config.Cfg.ES.Password
	if addr == "" || user == "" || pwd == "" {

	}
	client, err := elastic.NewClient(elastic.SetURL(addr), elastic.SetBasicAuth(user, pwd), elastic.SetSniff(false))
	if err != nil {
		err = fmt.Errorf("elastic.NewClient err: %s", err.Error())
		return
	}
	return &Es{client}, nil
}
