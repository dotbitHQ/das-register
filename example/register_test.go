package example

import (
	"das_register_server/http_server/handle"
	"encoding/json"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/sign"
	"github.com/ethereum/go-ethereum/signer/core"
	"github.com/scorpiotzh/toolib"
)

func doSignList(si *handle.SignInfo) error {
	privateTron := ""
	privateEth := ""
	for i, v := range si.SignList {
		switch v.SignType {
		case common.DasAlgorithmIdTron:
			data, err := sign.TronSignature(true, common.Hex2Bytes(v.SignMsg), privateTron)
			if err != nil {
				return err
			}
			si.SignList[i].SignMsg = common.Bytes2Hex(data)
		case common.DasAlgorithmIdEth712:
			var obj3 core.TypedData
			mmJson := toolib.JsonString(si.MMJson)
			_ = json.Unmarshal([]byte(mmJson), &obj3)
			mmHash, signature, err := sign.EIP712Signature(obj3, privateEth)
			if err != nil {
				return err
			}
			fmt.Println(common.Bytes2Hex(signature))
			fmt.Println(common.Bytes2Hex(mmHash))
			signMsg := append(signature, mmHash...)
			si.SignList[i].SignMsg = common.Bytes2Hex(signMsg) + "0000000000000005"
		}
	}
	return nil
}

func sendTx(si *handle.SignInfo) error {
	var req handle.ReqTransactionSend
	req.SignKey = si.SignKey
	req.SignList = si.SignList
	url := TestUrl + "/transaction/send"
	var data handle.RespTransactionSend
	if err := doReq(url, req, &data); err != nil {
		return err
	}
	fmt.Println(data.Hash)
	return nil
}
