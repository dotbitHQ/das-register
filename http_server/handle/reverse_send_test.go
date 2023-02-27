package handle

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/sign"
	"testing"
)

func TestPersonalSign(t *testing.T) {
	digest := common.Hex2Bytes("0x66726f6d206469643a2057f2e68d6c982d02c0495307b820107448ae7ab3de841ec072b7250c90efd55e")
	signature := common.Hex2Bytes(fixSignature("0x06b3abdf1a885d2a4741d39250a1080d66e3ba47add98c091574b1feb886a68e20e587add97c600064f0cace958671ebabe83c351f5d6265f1808d16e2ec65361c"))
	address := "0xdeeFC10a42cD84c072f2b0e2fA99061a74A0698c"
	signOk, err := sign.VerifyPersonalSignature(signature, digest, address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(signOk)
}
