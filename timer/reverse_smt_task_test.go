package timer

import (
	"testing"
)

func Test_ReverseSmtHasOldCreate(t *testing.T) {
	if err := txTimer.doReverseSmtTask(); err != nil {
		t.Fatal(err)
	}
}
