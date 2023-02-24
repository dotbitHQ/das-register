package handle

import "testing"

func Test_GenSignMsg(t *testing.T) {
	c := ReverseSmtSignCache{
		Nonce:   1,
		Account: "test.bit",
	}
	c.GenSignMsg()
	t.Log(c.SignMsg)
}
