package example

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"testing"
)

func TestRefundPreTx(t *testing.T) {
	dc, err := getNewDasCoreMainNet()
	if err != nil {
		t.Fatal(err)
	}
	dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		t.Fatal(err)
	}
	asContract, err := core.GetDasContractInfo(common.DasContractNameAlwaysSuccess)
	if err != nil {
		t.Fatal(err)
	}
	preContract, err := core.GetDasContractInfo(common.DasContractNamePreAccountCellType)
	if err != nil {
		t.Fatal(err)
	}

	searchKey := indexer.SearchKey{
		Script:     asContract.ToScript(nil),
		ScriptType: indexer.ScriptTypeLock,
		ArgsLen:    0,
		Filter: &indexer.CellsFilter{
			Script:              preContract.ToScript(nil),
			OutputDataLenRange:  nil,
			OutputCapacityRange: nil,
			BlockRange:          nil,
		},
	}

	liveCells, err := dc.Client().GetCells(context.Background(), &searchKey, indexer.SearchOrderAsc, 1000, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range liveCells.Objects {
		res, err := dc.Client().GetTransaction(context.Background(), v.OutPoint.TxHash)
		if err != nil {
			t.Fatal(err)
		}
		preBuilder, err := witness.PreAccountCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
		if err != nil {
			continue
		} else {
			refundLock := preBuilder.RefundLock
			if refundLock == nil {
				continue
			}
			refundLockScript := molecule.MoleculeScript2CkbScript(refundLock)
			if dasContract.IsSameTypeId(refundLockScript.CodeHash) {
				fmt.Println(v.OutPoint.TxHash.String(), common.Bytes2Hex(refundLockScript.Args))
				ownerHex, _, err := dc.Daf().ScriptToHex(refundLockScript)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Println(ownerHex.DasAlgorithmId, ownerHex.AddressHex)
			}
		}
	}
}
