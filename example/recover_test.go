package example

import (
	"context"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"testing"
)

func TestRecoverCkb(t *testing.T) {
	ckbUrl := ""
	indexerUrl := ""
	ckbClient, err := rpc.DialWithIndexer(ckbUrl, indexerUrl)
	if err != nil {
		t.Fatal(err)
	}
	addrParse, err := address.Parse("ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q")
	if err != nil {
		t.Fatal(err)
	}
	searchKey := indexer.SearchKey{
		Script:     addrParse.Script,
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:              nil,
			OutputDataLenRange:  &[2]uint64{32, 33},
			OutputCapacityRange: nil,
			BlockRange:          nil,
		},
	}
	//total, err := ckbClient.GetCellsCapacity(context.Background(), &searchKey)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println("total:", total)
	liveCells, err := ckbClient.GetCells(context.Background(), &searchKey, indexer.SearchOrderDesc, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range liveCells.Objects {
		fmt.Println(v.BlockNumber, v.OutPoint.TxHash, v.OutPoint.Index)
	}
}
