package handle

import (
	"context"
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"sync"
	"testing"
)

func TestCheckAccountCharSet(t *testing.T) {
	_ = getDasCore(t)
	var h HttpHandle
	var req ReqAccountSearch
	reqStr := `{"chain_type":1,"address":"0xa0324794ff56ecb258220046034a363d0da98f51","account":"1111ぁぁぁぁ.bit","account_char_str":[{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"1"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":5,"char":"ぁ"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}]}`

	json.Unmarshal([]byte(reqStr), &req)
	var apiResp api_code.ApiResp
	h.checkAccountCharSet(&req, &apiResp)
	fmt.Println(apiResp)
}

func getDasCore(t *testing.T) *core.DasCore {
	ckbUrl := "https://testnet.ckb.dev/"
	indexerUrl := "https://testnet.ckb.dev/indexer"
	// ckb node
	ckbClient, err := rpc.DialWithIndexer(ckbUrl, indexerUrl)
	if err != nil {
		t.Fatal(err)
	}

	// das init
	env := core.InitEnvOpt(common.DasNetTypeTestnet2, common.DasContractNameConfigCellType)
	ops := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(config.Cfg.Server.Net),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	var wg sync.WaitGroup
	dasCore := core.NewDasCore(context.Background(), &wg, ops...)
	dasCore.InitDasContract(env.MapContract)
	if err := dasCore.InitDasConfigCell(); err != nil {
		t.Fatal(err)
	}
	return dasCore
}
