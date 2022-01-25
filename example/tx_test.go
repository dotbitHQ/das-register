package example

import (
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"testing"
)

func TestOccupiedCapacity(t *testing.T) {
	applyOutputs := &types.CellOutput{
		Lock: &types.Script{
			CodeHash: types.HexToHash("0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8"),
			HashType: types.HashTypeType,
			Args:     common.Hex2Bytes("0xc866479211cadf63ad115b9da50a6c16bd3d226d"),
		},
		Type: &types.Script{
			CodeHash: types.HexToHash("0x0fbff871dd05aee1fda2be38786ad21d52a2765c6025d1ef6927d761d51a3cd1"),
			HashType: types.HashTypeType,
			Args:     nil,
		},
	}
	applyData := common.Hex2Bytes("0xc8490e95b537250bc9e3685c0ba8b9c9708b8c45592f31a68f50e88c06957cd984db3d00000000003ae5df6100000000")
	applyOutputs.Capacity = applyOutputs.OccupiedCapacity(applyData)
	fmt.Println(applyOutputs.Capacity)
}

func TestPrice(t *testing.T) {
	newPrice := uint64(5000000)
	quote := uint64(19537)
	invitedDiscount := uint64(500)
	priceCapacity := (newPrice * 1 / quote) * common.OneCkb
	priceCapacity = (priceCapacity / common.PercentRateBase) * (common.PercentRateBase - invitedDiscount)
	fmt.Println(priceCapacity)
}
