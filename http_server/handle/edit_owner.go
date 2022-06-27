package handle

import (
	"das_register_server/config"
	"das_register_server/http_server/api_code"
	"das_register_server/internal"
	"das_register_server/tables"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/gin-gonic/gin"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqEditOwner struct {
	ChainType  common.ChainType `json:"chain_type"`
	Address    string           `json:"address"`
	Account    string           `json:"account"`
	EvmChainId int64            `json:"evm_chain_id"`
	RawParam   struct {
		ReceiverChainType common.ChainType `json:"receiver_chain_type"`
		ReceiverAddress   string           `json:"receiver_address"`
	} `json:"raw_param"`
}

type RespEditOwner struct {
	SignInfo
}

func (h *HttpHandle) RpcEditOwner(p json.RawMessage, apiResp *api_code.ApiResp) {
	var req []ReqEditOwner
	err := json.Unmarshal(p, &req)
	if err != nil {
		log.Error("json.Unmarshal err:", err.Error())
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	} else if len(req) == 0 {
		log.Error("len(req) is 0")
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return
	}

	if err = h.doEditOwner(&req[0], apiResp); err != nil {
		log.Error("doVersion err:", err.Error())
	}
}

func (h *HttpHandle) EditOwner(ctx *gin.Context) {
	var (
		funcName = "EditOwner"
		clientIp = GetClientIp(ctx)
		req      ReqEditOwner
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doEditOwner(&req, &apiResp); err != nil {
		log.Error("doEditOwner err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doEditOwner(req *ReqEditOwner, apiResp *api_code.ApiResp) error {
	var resp RespEditOwner

	addressHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.ChainType,
		AddressNormal: req.Address,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address NormalToHex err")
		return fmt.Errorf("NormalToHex err: %s", err.Error())
	}
	req.ChainType, req.Address = addressHex.ChainType, addressHex.AddressHex

	ownerHex, err := h.dasCore.Daf().NormalToHex(core.DasAddressNormal{
		ChainType:     req.RawParam.ReceiverChainType,
		AddressNormal: req.RawParam.ReceiverAddress,
		Is712:         true,
	})
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "owner address NormalToHex err")
		return fmt.Errorf("owner NormalToHex err: %s", err.Error())
	}
	req.RawParam.ReceiverChainType, req.RawParam.ReceiverAddress = ownerHex.ChainType, ownerHex.AddressHex
	if !checkChainType(req.RawParam.ReceiverChainType) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, fmt.Sprintf("chain type [%d] inavlid", req.RawParam.ReceiverChainType))
		return nil
	}
	//
	if req.Account == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "account is invalid")
		return nil
	}
	if req.Address == "" || req.RawParam.ReceiverAddress == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "address is invalid")
		return nil
	}

	if err := h.checkSystemUpgrade(apiResp); err != nil {
		return fmt.Errorf("checkSystemUpgrade err: %s", err.Error())
	}

	if ok := internal.IsLatestBlockNumber(config.Cfg.Server.ParserUrl); !ok {
		apiResp.ApiRespErr(api_code.ApiCodeSyncBlockNumber, "sync block number")
		return fmt.Errorf("sync block number")
	}

	if exi := h.rc.AccountLimitExist(req.Account); exi {
		apiResp.ApiRespErr(api_code.ApiCodeOperationFrequent, "the operation is too frequent")
		return fmt.Errorf("AccountActionLimitExist: %d %s %s", req.ChainType, req.Address, req.Account)
	}

	accountId := common.Bytes2Hex(common.GetAccountIdByAccount(req.Account))
	acc, err := h.dbDao.GetAccountInfoByAccountId(accountId)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeDbError, "search account err")
		return fmt.Errorf("SearchAccount err: %s", err.Error())
	}
	if acc.Id == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeAccountNotExist, "account not exist")
		return nil
	} else if acc.Status == tables.AccountStatusOnSale || acc.Status == tables.AccountStatusOnAuction {
		apiResp.ApiRespErr(api_code.ApiCodeAccountStatusOnSaleOrAuction, "account on sale or auction")
		return nil
	} else if acc.IsExpired() {
		apiResp.ApiRespErr(api_code.ApiCodeAccountIsExpired, "account is expired")
		return nil
	} else if req.ChainType != acc.OwnerChainType || !strings.EqualFold(req.Address, acc.Owner) {
		apiResp.ApiRespErr(api_code.ApiCodePermissionDenied, "transfer owner permission denied")
		return nil
	} else if req.RawParam.ReceiverChainType == acc.OwnerChainType && strings.EqualFold(req.RawParam.ReceiverAddress, acc.Owner) {
		apiResp.ApiRespErr(api_code.ApiCodeSameLock, "same address")
		return nil
	} else if acc.ParentAccountId != "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "not support sub account")
		return nil
	}

	if (req.ChainType == common.ChainTypeMixin && req.RawParam.ReceiverChainType != common.ChainTypeMixin) ||
		(req.ChainType != common.ChainTypeMixin && req.RawParam.ReceiverChainType == common.ChainTypeMixin) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "ChainType is invalid")
		return nil
	}

	var reqBuild reqBuildTx
	reqBuild.Action = common.DasActionTransferAccount
	reqBuild.Account = req.Account
	reqBuild.ChainType = req.ChainType
	reqBuild.Address = req.Address
	reqBuild.Capacity = 0
	reqBuild.EvmChainId = req.EvmChainId

	var p editOwnerParams
	p.account = &acc
	p.ownerChainType = req.RawParam.ReceiverChainType
	p.ownerAddress = req.RawParam.ReceiverAddress
	txParams, err := h.buildEditOwnerTx(&reqBuild, &p)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildEditManagerTx err: %s", err.Error())
	}
	if si, err := h.buildTx(&reqBuild, txParams); err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "build tx err: "+err.Error())
		return fmt.Errorf("buildTx: %s", err.Error())
	} else {
		resp.SignInfo = *si
	}

	apiResp.ApiRespOK(resp)
	return nil
}

type editOwnerParams struct {
	account        *tables.TableAccountInfo
	ownerChainType common.ChainType
	ownerAddress   string
}

func (h *HttpHandle) buildEditOwnerTx(req *reqBuildTx, p *editOwnerParams) (*txbuilder.BuildTransactionParams, error) {
	var txParams txbuilder.BuildTransactionParams

	// inputs account cell
	accOutPoint := common.String2OutPointStruct(p.account.Outpoint)
	txParams.Inputs = append(txParams.Inputs, &types.CellInput{
		PreviousOutput: accOutPoint,
	})

	// witness action
	actionWitness, err := witness.GenActionDataWitness(common.DasActionTransferAccount, nil)
	if err != nil {
		return nil, fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)

	// witness account cell
	res, err := h.dasCore.Client().GetTransaction(h.ctx, accOutPoint.TxHash)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction err: %s", err.Error())
	}
	builderMap, err := witness.AccountCellDataBuilderMapFromTx(res.Transaction, common.DataTypeNew)
	if err != nil {
		return nil, fmt.Errorf("AccountCellDataBuilderMapFromTx err: %s", err.Error())
	}
	builder, ok := builderMap[req.Account]
	if !ok {
		return nil, fmt.Errorf("builderMap not exist account: %s", req.Account)
	}

	timeCell, err := h.dasCore.GetTimeCell()
	if err != nil {
		return nil, fmt.Errorf("GetTimeCell err: %s", err.Error())
	}

	accWitness, accData, err := builder.GenWitness(&witness.AccountCellParam{
		OldIndex:              0,
		NewIndex:              0,
		Action:                common.DasActionTransferAccount,
		LastTransferAccountAt: timeCell.Timestamp(),
	})
	txParams.Witnesses = append(txParams.Witnesses, accWitness)
	accData = append(accData, res.Transaction.OutputsData[builder.Index][32:]...)

	// outputs account cell
	builderConfigCell, err := h.dasCore.ConfigCellDataBuilderByTypeArgs(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("ConfigCellDataBuilderByTypeArgs err: %s", err.Error())
	}
	commonFee, err := builderConfigCell.AccountCommonFee()
	if err != nil {
		return nil, fmt.Errorf("AccountCommonFee err: %s", err.Error())
	}
	capacity := res.Transaction.Outputs[builder.Index].Capacity - commonFee

	contractAcc, err := core.GetDasContractInfo(common.DasContractNameAccountCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	contractDas, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return nil, fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}

	ownerHex := core.DasAddressHex{
		DasAlgorithmId: p.ownerChainType.ToDasAlgorithmId(true),
		AddressHex:     p.ownerAddress,
		IsMulti:        false,
		ChainType:      p.ownerChainType,
	}
	lockArgs, err := h.dasCore.Daf().HexToArgs(ownerHex, ownerHex)
	if err != nil {
		return nil, fmt.Errorf("HexToArgs err: %s", err.Error())
	}
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: capacity,
		Lock:     contractDas.ToScript(lockArgs),
		Type:     contractAcc.ToScript(nil),
	})
	txParams.OutputsData = append(txParams.OutputsData, accData)

	// cell deps
	heightCell, err := h.dasCore.GetHeightCell()
	if err != nil {
		return nil, fmt.Errorf("GetHeightCell err: %s", err.Error())
	}

	configCellAcc, err := core.GetDasConfigCellInfo(common.ConfigCellTypeArgsAccount)
	if err != nil {
		return nil, fmt.Errorf("GetDasConfigCellInfo err: %s", err.Error())
	}

	txParams.CellDeps = append(txParams.CellDeps,
		heightCell.ToCellDep(),
		timeCell.ToCellDep(),
		configCellAcc.ToCellDep(),
	)

	return &txParams, nil
}
