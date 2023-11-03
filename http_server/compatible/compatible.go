package compatible

import (
	"das_register_server/config"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"reflect"
)

func ChainTypeAndCoinType(req interface{}, dc *core.DasCore) (dasAddressHex core.DasAddressHex, err error) {
	var address string
	var chainType common.ChainType
	var coinType string
	var coin_key_info core.KeyInfo
	if v, res := GetFieldFromInterface(req, "Address"); res {
		if vv, ok := v.(string); ok {
			address = vv
		}
	}
	if v, res := GetFieldFromInterface(req, "ChainType"); res {
		if vv, ok := v.(common.ChainType); ok {
			chainType = vv
		}
	}
	if v, res := GetFieldFromInterface(req, "Type"); res {
		if vv, ok := v.(string); ok {
			coinType = vv
		}
	}
	if v, res := GetFieldFromInterface(req, "KeyInfo"); res {
		if vv, ok := v.(core.KeyInfo); ok {
			coin_key_info = vv
		}
	}
	if address != "" {
		dasAddressHex, err = dc.Daf().NormalToHex(core.DasAddressNormal{
			ChainType:     chainType,
			AddressNormal: address,
			Is712:         true,
		})
		if err != nil {
			err = fmt.Errorf("NormalToHex err: %s", err.Error())
		}
		return dasAddressHex, err
	} else {
		chainTypeAddress := core.ChainTypeAddress{
			Type:    coinType,
			KeyInfo: coin_key_info,
		}
		res, err := chainTypeAddress.FormatChainTypeAddress(config.Cfg.Server.Net, true)
		if err != nil {
			err = fmt.Errorf("FormatChainTypeAddress err: %s", err.Error())
			return dasAddressHex, err
		}
		return *res, err
	}
}

func GetFieldFromInterface(input interface{}, fieldName string) (interface{}, bool) {
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Struct {
		return nil, false
	}

	fieldValue := val.FieldByName(fieldName)
	if fieldValue.IsValid() {
		return fieldValue.Interface(), true
	}
	return nil, false
}
