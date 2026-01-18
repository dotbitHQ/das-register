package internal

import (
	"time"

	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/parnurzeal/gorequest"
)

var log = logger.NewLogger("internal", logger.LevelDebug)

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

	maxRetries := 3
	apiUrl := url + "/latest/block/number"

	for i := 0; i < maxRetries; i++ {
		var resp RespApi
		var isLatest RespIsLatestBlockNumber
		resp.Data = &isLatest

		_, body, errs := gorequest.New().
			Timeout(10 * time.Second).
			Post(apiUrl).
			SendStruct(nil).
			EndStruct(&resp)

		if len(errs) > 0 {
			log.Warnf("IsLatestBlockNumber retry %d/%d err: %v", i+1, maxRetries, errs)
			if i < maxRetries-1 {
				time.Sleep(time.Second * time.Duration(i+1)) // backoff: 1s, 2s
			}
			continue
		}

		log.Info("body:", string(body))
		return isLatest.IsLatestBlockNumber
	}

	log.Errorf("IsLatestBlockNumber failed after %d retries", maxRetries)
	return false
}
