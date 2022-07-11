package example

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"math/big"
	"strings"
	"testing"
)

func TestRenew(t *testing.T) {
	listStr := ``
	list := strings.Split(strings.ToLower(listStr), "\n")
	//fmt.Println(list,len(list))

	for _, v := range list {
		a := fmt.Sprintf(`curl -X POST http://127.0.0.1:8119/v1/account/renew -d'{"chain_type":1,"address":"0x15a33588908cf8edb27d1abe3852bf287abd3891","account":"%s","renew_years":1}'`, strings.TrimSpace(v))
		fmt.Println(a)
	}

}

func TestAccountToTokenId(t *testing.T) {
	account := "abc.bit"
	fmt.Println("account:", account)
	accountIdBys := common.GetAccountIdByAccount(account)
	fmt.Println("account id hex:", common.Bytes2Hex(accountIdBys))
	fmt.Println("account id bys:", accountIdBys)
	tokenId := new(big.Int).SetBytes(accountIdBys).String()
	fmt.Println("token id:", tokenId)
}
