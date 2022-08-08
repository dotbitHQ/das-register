package example

import (
	"das_register_server/http_server/handle"
	"fmt"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestCharacterSetList(t *testing.T) {
	var req handle.ReqCharacterSetList
	url := TestUrl + "/character/set/list"

	var data handle.RespCharacterSetList
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(toolib.JsonString(&data))
}
