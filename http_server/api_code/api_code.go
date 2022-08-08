package api_code

type ApiCode = int

const (
	ApiCodeSuccess        ApiCode = 0
	ApiCodeError500       ApiCode = 500
	ApiCodeParamsInvalid  ApiCode = 10000
	ApiCodeMethodNotExist ApiCode = 10001
	ApiCodeDbError        ApiCode = 10002
	ApiCodeCacheError     ApiCode = 10003

	ApiCodeTransactionNotExist ApiCode = 11001
	//ApiCodePermissionDenied    ApiCode = 11002
	ApiCodeNotSupportAddress   ApiCode = 11005
	ApiCodeInsufficientBalance ApiCode = 11007
	ApiCodeTxExpired           ApiCode = 11008
	ApiCodeAmountInvalid       ApiCode = 11010
	ApiCodeRejectedOutPoint    ApiCode = 11011
	ApiCodeNotEnoughChange     ApiCode = 11014
	ApiCodeSyncBlockNumber     ApiCode = 11012
	ApiCodeOperationFrequent   ApiCode = 11013

	ApiCodeReverseAlreadyExist ApiCode = 12001
	ApiCodeReverseNotExist     ApiCode = 12002

	ApiCodeNotOpenForRegistration       ApiCode = 30001
	ApiCodeAccountNotExist              ApiCode = 30003
	ApiCodeAccountAlreadyRegister       ApiCode = 30004
	ApiCodeAccountLenInvalid            ApiCode = 30014
	ApiCodeOrderNotExist                ApiCode = 30006
	ApiCodeAccountIsExpired             ApiCode = 30010
	ApiCodePermissionDenied             ApiCode = 30011
	ApiCodeAccountContainsInvalidChar   ApiCode = 30015
	ApiCodeReservedAccount              ApiCode = 30017
	ApiCodeInviterAccountNotExist       ApiCode = 30018
	ApiCodeSystemUpgrade                ApiCode = 30019
	ApiCodeRecordInvalid                ApiCode = 30020
	ApiCodeSameLock                     ApiCode = 30023
	ApiCodeChannelAccountNotExist       ApiCode = 30026
	ApiCodeOrderPaid                    ApiCode = 30027
	ApiCodeUnAvailableAccount           ApiCode = 30029
	ApiCodeAccountStatusOnSaleOrAuction ApiCode = 30031
	ApiCodePayTypeInvalid               ApiCode = 30032
	ApiCodeSameOrderInfo                ApiCode = 30033
	ApiCodeSigErr                       ApiCode = 30034 // contracte -31
	ApiCodeOnCross                      ApiCode = 30035
	ApiCodeSubAccountNotEnabled         ApiCode = 30036
	ApiCodeAfterGracePeriod             ApiCode = 30037
)

const (
	MethodTokenList         = "das_tokenList"
	MethodConfigInfo        = "das_dasConfig"
	MethodAccountList       = "das_accountList"
	MethodAccountMine       = "das_myAccounts"
	MethodAccountDetail     = "das_accountDetail"
	MethodAccountRecords    = "das_accountParsingRecords"
	MethodReverseLatest     = "das_reverseLatest"
	MethodReverseList       = "das_reverseList"
	MethodTransactionStatus = "das_transactionStatus"
	MethodBalanceInfo       = "das_balanceInfo"
	MethodTransactionList   = "das_transactionList"
	MethodRewardsMine       = "das_myRewards"
	MethodWithdrawList      = "das_withdrawList"
	MethodAccountSearch     = "das_accountSearch"
	MethodRegisteringList   = "das_registeringAccounts"
	MethodOrderDetail       = "das_orderDetail"
	MethodAddressDeposit    = "das_addressDeposit"
	MethodCharacterSetList  = "das_characterSetList"

	MethodReverseDeclare   = "das_reverseDeclare"
	MethodReverseRedeclare = "das_reverseRedeclare"
	MethodReverseRetract   = "das_reverseRetract"
	MethodTransactionSend  = "das_transactionSend"
	MethodBalanceWithdraw  = "das_balanceWithdraw"
	MethodBalanceTransfer  = "das_balanceTransfer"
	MethodBalanceDeposit   = "das_balanceDeposit"
	MethodEditManager      = "das_editManager"
	MethodEditOwner        = "das_transferAccount"
	MethodEditRecords      = "das_editRecords"
	MethodOrderRenew       = "das_submitRenewOrder"
	MethodBalancePay       = "das_dasBalancePay"
	MethodOrderRegister    = "das_submitRegisterOrder"
	MethodOrderChange      = "das_changeOrder"
	MethodOrderPayHash     = "das_doOrderPayHash"
	MethodEditScript       = "das_editScript"

	MethodCkbRpc = "das_ckbRpc"
)

type ApiResp struct {
	ErrNo  ApiCode     `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

func ApiRespOK(data interface{}) ApiResp {
	return ApiResp{
		ErrNo:  ApiCodeSuccess,
		ErrMsg: "",
		Data:   data,
	}
}

func ApiRespErr(errNo ApiCode, errMsg string) ApiResp {
	return ApiResp{
		ErrNo:  errNo,
		ErrMsg: errMsg,
		Data:   nil,
	}
}

func (a *ApiResp) ApiRespErr(errNo ApiCode, errMsg string) {
	a.ErrNo = errNo
	a.ErrMsg = errMsg
}

func (a *ApiResp) ApiRespOK(data interface{}) {
	a.ErrNo = ApiCodeSuccess
	a.Data = data
}
