package timer

import (
	"das_register_server/tables"
	"fmt"
	"sync"
)

var (
	tokenLock sync.RWMutex
	mapToken  map[tables.PayTokenId]tables.TableTokenPriceInfo
)

func (t *TxTimer) doUpdateTokenMap() error {
	tokenLock.Lock()
	defer tokenLock.Unlock()
	list, err := t.dbDao.GetTokenPriceList()
	if err != nil {
		return fmt.Errorf("SearchTokenInfoList err:%s", err.Error())
	}
	mapToken = make(map[tables.PayTokenId]tables.TableTokenPriceInfo)
	for i, v := range list {
		mapToken[v.TokenId] = list[i]
	}
	return nil
}

func GetTokenInfo(tokenId tables.PayTokenId) tables.TableTokenPriceInfo {
	if tokenId == tables.TokenIdDas || tokenId == tables.TokenIdCkbInternal {
		tokenId = tables.TokenIdCkb
	}
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	t, _ := mapToken[tokenId]
	return t
}

func GetTokenList() map[tables.PayTokenId]tables.TableTokenPriceInfo {
	tokenLock.Lock()
	defer tokenLock.Unlock()
	return mapToken
}
