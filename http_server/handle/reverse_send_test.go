package handle

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/sign"
	"testing"
)

func TestPersonalSign(t *testing.T) {
	signMsg := common.Bytes2Hex([]byte("66726f6d206469643a209/NMLnYWHox2ZHqKi3bpDqBdFuT8Mp1H1yTqu7bnEBs="))
	signature := fixSignature("0xb01a6bde383c8ae129faa38786a3a0849a62f43e82d6984b0fb1ca043b4796dd7fea1ccdc809e85f126791cc6addb655ba69f8a3dc592851d68b863adbdb28ef1b")
	signOk, err := sign.VerifyPersonalSignature(common.Hex2Bytes(signature), common.Hex2Bytes(signMsg), "0xc0d0087dA03480f9d7e7E1D76d2DCa4bb0A98B17")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(signOk)
}
