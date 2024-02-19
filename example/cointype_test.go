package example

import (
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
)

type ReqEditManager struct {
	ChainType common.ChainType `json:"chain_type"`
	core.ChainTypeAddress
	Address    string `json:"address"`
	Account    string `json:"account"`
	EvmChainId int64  `json:"evm_chain_id"`
	RawParam   struct {
		ManagerChainType common.ChainType `json:"manager_chain_type"`
		ManagerAddress   string           `json:"manager_address"`
	} `json:"raw_param"`
}

//
//func TestCointype(t *testing.T) {
//	req := ReqEditManager{
//		EvmChainId: 5,
//		ChainType:  1,
//		Address:    "0xd437b8e9cA16Fce24bF3258760c3567214213C5A",
//		Account:    "michsjwq.bit",
//	}
//	req.Type = "blockchain"
//	req.KeyInfo = core.KeyInfo{
//		CoinType: "60",
//		Key:      "0xd437b8e9cA16Fce24bF3258760c3567214213C5A",
//	}
//	dc, _ := getNewDasCoreTestnet2()
//	err, addrHex := compatible.ChaintyeAndCoinType(req, dc)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(addrHex)
//}
