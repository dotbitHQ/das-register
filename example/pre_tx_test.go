package example

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/molecule"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"testing"
)

func TestPreTx(t *testing.T) {
	dc, err := getNewDasCoreMainNet()
	if err != nil {
		t.Fatal(err)
	}
	res, err := dc.Client().GetTransaction(context.Background(), types.HexToHash("0x831371eeb2b7de5d3ea9a3ba91ac5fd3d1a09dac8b26308210a537aaad658592"))
	if err != nil {
		t.Fatal(err)
	}
	preBuilder, err := witness.PreAccountCellDataBuilderFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		t.Fatal()
	} else {
		refundLock := preBuilder.RefundLock

		refundLockScript := molecule.MoleculeScript2CkbScript(refundLock)
		fmt.Println(refundLockScript.CodeHash.String(), common.Bytes2Hex(refundLockScript.Args))
	}
}

func TestBalance(t *testing.T) {
	dc, err := getNewDasCoreMainNet()
	if err != nil {
		t.Fatal(err)
	}
	res, _, err := dc.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache: nil,
		LockScript: &types.Script{
			CodeHash: types.HexToHash("0x9376c3b5811942960a846691e16e477cf43d7c7fa654067c9948dfcd09a32137"),
			HashType: "type",
			Args:     common.Hex2Bytes("0x053a6cab3323833f53754db4202f5741756c436ede053a6cab3323833f53754db4202f5741756c436ede"),
		},
		CapacityNeed:      0,
		CapacityForChange: 0,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	for _, v := range res {
		fmt.Println(v.OutPoint.TxHash.String())
	}
}

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
