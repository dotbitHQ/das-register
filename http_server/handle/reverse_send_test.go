package handle

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/sign"
	"testing"
)

func TestPersonalSign(t *testing.T) {
	signature := fixSignature("0x6e6c9b3f72a47eb15ee6336e776e91aee8a2d4390f86ffa226b099482b95537a3bfd96bb663cfe0e16daa4300c1d8eb3b13a5d8fec9c97112dbafa07b52e29f11b")
	signOk, err := sign.VerifyPersonalSignature(common.Hex2Bytes(signature), []byte("0x3636373236663664323036343639363433613230392f4e4d4c6e5957486f78325a48714b6933627044714264467554384d703148317954717537626e4542733d"), "0xc0d0087dA03480f9d7e7E1D76d2DCa4bb0A98B17")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(signOk)
}
