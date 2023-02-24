package timer

import (
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/witness"
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

func Test_GenWitness(t *testing.T) {

	builder := witness.NewReverseSmtBuilder()

	record := &witness.ReverseSmtRecord{
		Version:     1,
		Action:      "update",
		Signature:   "0x307836653663396233663732613437656231356565363333366537373665393161656538613264343339306638366666613232366230393934383262393535333761336266643936626236363363666530653136646161343330306331643865623362313361356438666563396339373131326462616661303762353265323966313162",
		SignType:    3,
		Address:     "0xc0d0087dA03480f9d7e7E1D76d2DCa4bb0A98B17",
		Proof:       "0x3078346334663030",
		PrevNonce:   0,
		PrevAccount: "",
		NextRoot:    []byte("df41487f90abe236cfc3b57cd269e50c75cd21a262ce2f8bd141fba8d28ef65d"),
		NextAccount: "test.bit",
	}
	bs, err := record.GenBytes()
	if err != nil {
		t.Fatal(err)
	}

	parseRecord, err := builder.FromBytes(bs)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*parseRecord)
}
