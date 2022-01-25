### 反向解析服务

* resp common:

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {}
}
```

```go
错误码
const (
ApiCodeSuccess        ApiCode = 0
ApiCodeError500       ApiCode = 500
ApiCodeParamsInvalid  ApiCode = 10000 // 请求参数错误
ApiCodeMethodNotExist ApiCode = 10001
ApiCodeDbError        ApiCode = 10002 // 数据库错误
ApiCodeCacheError     ApiCode = 10003 // 缓存错误

ApiCodeTransactionNotExist ApiCode = 11001 // 交易不存在
ApiCodePermissionDenied    ApiCode = 11002 // 权限不足
ApiCodeInsufficientBalance ApiCode = 11007 // 余额不足
ApiCodeTxExpired           ApiCode = 11008 // 交易过期
ApiCodeRejectedOutPoint    ApiCode = 11011 // cell 抢占
ApiCodeNotEnoughChange     ApiCode = 11014 // 不够找零
ApiCodeSyncBlockNumber     ApiCode = 11012 // 同步区块高度
ApiCodeOperationFrequent   ApiCode = 11013 // 操作频繁

ApiCodeReverseAlreadyExist ApiCode = 12001 // 已存在该反向解析
ApiCodeReverseNotExist     ApiCode = 12002 // 不存在该反向解析

ApiCodeAccountNotExist ApiCode = 30003 // 账号不存在
ApiCodeSystemUpgrade   ApiCode = 30019 // 系统升级
)
```

#### 1、查询最新的反向解析

* post: /reverse/latest
* req
    * ChainTypeCkb ChainType = 0
    * ChainTypeEth ChainType = 1
    * ChainTypeBtc ChainType = 2
    * ChainTypeTron ChainType = 3

```json
{
  "chain_type": 0,
  "address": "0x"
}
```

* resp
    * is_valid: 是否有效

```json
{
  "account": "",
  "is_valid": false
}
```

#### 2、设置反向解析

* post: /reverse/declare
* req
    * evm_chain_id：evm链的链ID

```json
{
  "chain_type": 0,
  "address": "0x",
  "account": "",
  "evm_chain_id": 1
}
```

* post

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 0,
      "sign_msg": ""
    }
  ],
  "mm_json": {}
}
```

#### 3、修改反向解析

* post: /reverse/redeclare
* req

```json
{
  "chain_type": 0,
  "address": "0x",
  "account": "",
  "evm_chain_id": 1
}
```

* post

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 0,
      "sign_msg": ""
    }
  ],
  "mm_json": {}
}
```

#### 4、删除反向解析

* post: /reverse/retract
* req

```json
{
  "chain_type": 0,
  "address": "0x",
  "evm_chain_id": 1
}
```

* post

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 0,
      "sign_msg": ""
    }
  ],
  "mm_json": {}
}
```

#### 5、发交易

* post: /transaction/send
* req

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 0,
      "sign_msg": ""
    }
  ]
}

```

* resp

```json
{
  "hash": ""
}
```

#### 6、账号列表（不包含出售账号）

* post: /account/list
* req

```json
{
  "chain_type": 0,
  "address": "0x"
}

```

* resp

```json
{
  "list": [
    {
      "account": "",
      "status": 0,
      "expired_at": 0
    }
  ]
}
```

#### 7、查交易状态

* post: /transaction/status
* req
    * ActionReverseDeclare TxAction = 8 // 设置反向解析
    * ActionReverseRedeclare TxAction = 9 // 编辑反向解析
    * ActionReverseRetract TxAction = 10 // 删除反向解析

```json
{
  "chain_type": 0,
  "address": "0x",
  "actions": [
    8,
    9
  ]
}
```

* resp
    * StatusRejected = -1
    * StatusConfirm = 1
    * StatusPending = 0

```json
{
  "block_number": 0,
  "hash": "0x",
  "action": 8,
  "status": 0
}
```

#### 8、账号详情

* post: /account/detail
* req:

```json
{
  "account": ""
}
```

* resp
    * SearchStatusRegisterNotOpen SearchStatus = -1 //未开放注册
    * SearchStatusRegisterAble SearchStatus = 0 //可注册
    * SearchStatusPaymentConfirm SearchStatus = 1 //支付确认
    * SearchStatusLockedAccount SearchStatus = 2 //账户锁定
    * SearchStatusRegistering SearchStatus = 3 //注册中
    * SearchStatusProposal SearchStatus = 4 //提案中
    * SearchStatusConfirmProposal SearchStatus = 5 //确认提案
    * SearchStatusRegistered SearchStatus = 6 //已注册
    * SearchStatusRetainAccount SearchStatus = 7 // 保留账户
    * SearchStatusOnSale SearchStatus = 8 // 出售
    * SearchStatusOnAuction SearchStatus = 9 // 竞拍
    * SearchStatusUnAvailableAccount SearchStatus = 13 // 禁止账户

```json
{
  "account": "",
  "owner_chain_type": 0,
  "owner": "0x",
  "manager_chain_type": 0,
  "manager": "0x",
  "registered_at": 0,
  "expired_at": 0,
  "status": 0
}
```

#### 9、配置信息

* post: /config/info
* req:
* resp:
    * reverse_record_capacity 反向解析花费的ckb
    * min_change_capacity 最小找0ckb

```json
{
  "reverse_record_capacity": 0,
  "min_change_capacity": 0
}
```

#### 10、余额

* post: /balance/info
* req:

```json
{
  "chain_type": 0,
  "address": "0x",
  "transfer_address": "短地址"
}
```

* resp:
    * das_lock_amount：712余额（05）
    * transfer_address_amount：非712余额(03+短地址)

```json
{
  "das_lock_amount": 0,
  "transfer_address_amount": 0
}
```

#### 11、交易流水

* post: /transaction/list
* req:

```json
{
  "chain_type": 1,
  "address": "0x",
  "page": 1,
  "size": 20
}
```

* post:
    * ActionUndefined TxAction = 99 // 未定义
    * ActionWithdrawFromWallet TxAction = 0 // 提现
    * ActionConsolidateIncome TxAction = 1 // 奖励
    * ActionStartAccountSale TxAction = 2 // 上架一口价
    * ActionEditAccountSale TxAction = 3 // 编辑一口价
    * ActionCancelAccountSale TxAction = 4 // 取消一口价
    * ActionBuyAccount TxAction = 5 // 买账号
    * ActionSaleAccount TxAction = 6 // 卖账号
    * ActionTransferBalance TxAction = 7 // 激活余额
    * ActionDeclareReverseRecord TxAction = 8 // 设置解析记录
    * ActionRedeclareReverseRecord TxAction = 9 // 修改解析记录
    * ActionRetractReverseRecord TxAction = 10 // 删除解析记录
    * ActionDasBalanceTransfer TxAction = 11 // 转账,das 余额注册
    * ActionEditRecords TxAction = 12 // 修改解析记录
    * ActionTransferAccount TxAction = 13 // 修改 owner
    * ActionEditManager TxAction = 14 // 修改manager

```json
{
  "total": 0,
  "list": [
    {
      "block_number": 0,
      "hash": "0x",
      "account": "xxx.bit",
      "action": 1,
      "capacity": 0,
      "timestamp": 0
    }
  ]
}
```

#### 12、das 余额提现

* post: /balance/withdraw
* req:

```json
{
  "chain_type": 0,
  "address": "",
  "receiver_chain_type": "",
  "receiver_address": "",
  "amount": 11600000000,
  "withdraw_all": true,
  "evm_chain_id": 1
}
```

* post:

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 0,
      "sign_msg": ""
    }
  ],
  "mm_json": {}
}
```

#### 12、das 余额激活

* post: /balance/transfer
* req:

```json
{
  "chain_type": 0,
  "address": "",
  "transfer_address": "",
  "evm_chain_id": 1
}
```

* post:

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 0,
      "sign_msg": ""
    }
  ],
  "mm_json": {}
}
```

#### 13、代币列表

* post: /token/list
* req:
* resp:

```json
{
  "token_list": [
    {
      "token_id": "",
      "chain_type": 0,
      "name": "",
      "symbol": "",
      "decimals": 0,
      "price": ""
    }
  ]
}
```  

#### 14、解析记录

* post: /account/records
* req:

```json
{
  "account": "xxxx.bit"
}
```  

* resp:

```json
{
  "records": [
    {
      "key": "",
      "type": "",
      "label": "",
      "value": "",
      "ttl": ""
    }
  ]
}
```