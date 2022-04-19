package example

import (
	"das_register_server/http_server/handle"
	"fmt"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestBalanceDeposit(t *testing.T) {
	var req handle.ReqBalanceDeposit
	req.FromCkbAddress = ""
	req.ToCkbAddress = ""
	req.Amount = 1

	url := TestUrl + "/balance/deposit"

	var data handle.RespBalanceDeposit
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(data))
}
