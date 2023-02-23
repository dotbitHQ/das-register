package timer

import (
	"github.com/dotbitHQ/das-lib/sign"
	"testing"
)

func Test_ReverseSmtHasOldCreate(t *testing.T) {
	if err := txTimer.doReverseSmtTask(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalSign(t *testing.T) {
	signHandler := sign.LocalSign("a46f1213966ec1c4624557f4f84dee9f07f4faca684b184cc650ceddd21cecf7")
	bs, err := signHandler("0x3636373236663664323036343639363433613230392f4e4d4c6e5957486f78325a48714b6933627044714264467554384d703148317954717537626e4542733d")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bs))
}
