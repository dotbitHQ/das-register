package internal

import (
	"github.com/parnurzeal/gorequest"
	"github.com/scorpiotzh/mylog"
)

var log = mylog.NewLogger("internal", mylog.LevelDebug)

type RespApi struct {
	ErrNo  int         `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

type RespIsLatestBlockNumber struct {
	IsLatestBlockNumber bool `json:"isLatestBlockNumber"`
}

func IsLatestBlockNumber(url string) bool {
	if url == "" {
		return true
	}
	url += "/latest/block/number"
	var resp RespApi
	var isLatest RespIsLatestBlockNumber
	resp.Data = &isLatest

	_, body, err := gorequest.New().Post(url).SendStruct(nil).EndStruct(&resp)
	log.Info("body:", string(body))
	if err != nil {
		log.Errorf("IsLatestBlockNumber err: %v", err)
		return false
	}
	return isLatest.IsLatestBlockNumber
}
