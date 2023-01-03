package example

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
	"testing"
)

func TestSince(t *testing.T) {
	since := uint64(4611686018427474304)
	var max uint64 = math.MaxUint64
	fmt.Println(max, since, max-since)
	s := 24 * 60 * 60
	fmt.Println(utils.SinceFromRelativeTimestamp(uint64(s)))
	fmt.Println(utils.SinceFromRelativeTimestamp(60 * 60))

	dc, err := getNewDasCoreTestnet2()
	if err != nil {
		t.Fatal(err)
	}
	block, err := dc.Client().GetBlockchainInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(block.MedianTime, block.Chain)
	blockNumber, err := dc.Client().GetTipBlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	block2, err := dc.Client().GetBlockByNumber(context.Background(), blockNumber)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(dc.Client().GetBlockMedianTime(context.Background(), block2.Header.Hash))
	//dc.Client().getblo
	//dc.Client().GetBlockMedianTime()
}
