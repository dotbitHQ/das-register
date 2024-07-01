* [Query API LIST](#query-api-list)
    * [Version](#version)
    * [Token List](#token-list)
    * [Config Info](#config-info)
    * [Account List](#account-list)
    * [Account Mine](#account-mine)
    * [Account Detail](#account-detail)
    * [Account Records](#account-records)
    * [Reverse Latest (Deprecated)](#reverse-latest)
    * [Reverse List (Deprecated)](#reverse-list)
    * [Transaction Status](#transaction-status)
    * [Balance Info](#balance-info)
    * [Transaction List](#transaction-list)
    * [Rewards Mine](#rewards-mine)
    * [Withdraw List](#withdraw-list)
    * [Account Search](#account-search)
    * [Account Registering List](#account-registering-list)
    * [Account Order Detail](#account-order-detail)
    * [Address Deposit](#address-deposit)
    * [Character Set List](#character-set-list)
    * [Account Auction Info](#account-auction-info)
    * [Account Auction Price](#account-auction-price)
    * [Account Auction OrderStatus](#account-auction-order_status)
    * [Account Auction PendingOrders](#account-auction-pending_orders)
    * [Account Recommend](#account-recommend)
    * [Account Check Coupon](#account-check-coupon)
* [OPERATE API LIST](#operate-api-list)
    * [Reverse Declare (Deprecated)](#reverse-declare)
    * [Reverse Redeclare (Deprecated)](#reverse-redeclare)
    * [Reverse Retract (Deprecated)](#reverse-retract)
    * [Transaction Send](#transaction-send)
    * [Balance Pay](#balance-pay)
    * [Balance Withdraw](#balance-withdraw)
    * [Balance Transfer](#balance-transfer)
    * [Balance Deposit](#balance-deposit)
    * [Account Edit Manager](#account-edit-manager)
    * [Account Edit Owner](#account-edit-owner)
    * [Account Edit Records](#account-edit-records)
    * [Account Order Renew](#account-order-renew)
    
    * [Account Order Register](#account-order-register)
    * [Account Order Change](#account-order-change)
    * [Account Order Pay Hash](#account-order-pay-hash)
    * [Account Register](#account-register)
    * [Account Renew](#account-renew)
    * [Account Auction Bid](#account-auction-bid)
* [NODE RPC](#node-rpc)
    * [Node Ckb Rpc](#node-ckb-rpc)

### Query API LIST

Please familiarize yourself with the meaning of some common parameters before reading the API list:

| param                                                                             | description                                        |
|:----------------------------------------------------------------------------------|:---------------------------------------------------|
| type                                                                              | Filled with "blockchain" for now                   |
| [coin_type](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)        | 60: eth, 195: trx, 9006: bsc, 966: matic, 3: doge  |
| account                                                                           | Contains the suffix `.bit` in it                   |
| key                                                                               | Generally refers to the blockchain address for now |


#### Version

**Request**

* path: /v1/version
* param: none

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "version": 1.0
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/config/info
```


#### Token List

**Request**

* path: /v1/token/list
* param: none

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "token_list": [
      {
        "token_id": "ckb_ckb",
        "chain_type": 0,
        "coin_type": "309",
        "contract": "",
        "name": "Nervos Network",
        "symbol": "CKB",
        "decimals": 8,
        "logo": "https://app.da.systems/images/components/portal-wallet.svg",
        "price": "0.01850608"
      },
      {
        "token_id": "polygon_matic",
        "chain_type": 1,
        "contract": "",
        "name": "Polygon",
        "symbol": "MATIC",
        "decimals": 18,
        "logo": "https://app.da.systems/images/components/polygon.svg",
        "price": "2.15"
      },
      {
        "token_id": "bsc_bnb",
        "chain_type": 5,
        "contract": "",
        "name": "Binance",
        "symbol": "BNB",
        "decimals": 18,
        "logo": "https://app.da.systems/images/components/binance-smart-chain.svg",
        "price": "435.85"
      },
      {
        "token_id": "wx_cny",
        "chain_type": 4,
        "contract": "",
        "name": "WeChat Pay",
        "symbol": "¥",
        "decimals": 2,
        "logo": "https://app.da.systems/images/components/wechat_pay.png",
        "price": "0.1569"
      },
      {
        "token_id": "tron_trx",
        "chain_type": 3,
        "contract": "",
        "name": "TRON",
        "symbol": "TRX",
        "decimals": 6,
        "logo": "https://app.da.systems/images/components/tron.svg",
        "price": "0.064233"
      },
      {
        "token_id": "btc_btc",
        "chain_type": 2,
        "contract": "",
        "name": "Bitcoin",
        "symbol": "BTC",
        "decimals": 8,
        "logo": "https://app.da.systems/images/components/bitcoin.svg",
        "price": "42161"
      },
      {
        "token_id": "eth_eth",
        "chain_type": 1,
        "contract": "",
        "name": "Ethereum",
        "symbol": "ETH",
        "decimals": 18,
        "logo": "https://app.da.systems/images/components/ethereum.svg",
        "price": "3115.47"
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/token/list
```

#### Config Info

**Request**

* path: /v1/config/info
* param: none

**Response**

```json
{
  "data": {
    "account_expiration_grace_period": 7776000,
    "edit_manager_throttle": 300,
    "edit_records_throttle": 300,
    "income_cell_min_transfer_value": 12000000000,
    "inviter_discount": "0.05",
    "max_account_len": 42,
    "min_account_len": 4,
    "min_change_capacity": 12000000000,
    "min_sell_price": 20000000000,
    "min_ttl": 300,
    "premium": "0",
    "premium_base": "0.6",
    "premium_percentage": "0.037",
    "profit_rate_of_inviter": "0.1",
    "reverse_record_capacity": 20100000000,
    "sale_cell_capacity": 20100000000,
    "timestamp_on_chain": 1719417385,
    "transfer_throttle": 300
  },
  "err_msg": "",
  "err_no": 0
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/config/info
```

#### Account List

**Request**

* path: /v1/account/list
    * get user's not on sale accounts
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "list": [
      {
        "account": "9aaaaaaa.bit",
        "status": 8,
        "expired_at": 1718955772000,
        "registered_at": 1624347772000,
        "enable_sub_account": 0,
        "renew_sub_account_price": 15,
        "nonce": 2
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/list -d'{"chain_type":1,"address":"0xc9f53b1d85356B60453F867610888D89a0B667Ad"}'
```

#### Account Mine

**Request**

* path: /v1/account/mine
    * get user's accounts by pagination
* param:
    * CategoryDefault Category = 0
    * CategoryMainAccount Category = 1 
    * CategorySubAccount Category = 2
    * CategoryOnSale Category = 3
    * CategoryExpireSoon Category = 4
    * CategoryToBeRecycled Category = 5
    * CategoryMainAccountDisableSecondLevelDID = 6
    * CategoryMainAccountEnableSecondLevelDID = 7

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "309",
    "key": "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqumxlsr77xk84ppzmm2rlmfvam37gc2t75yqwdn0cpl0rtr6ss3da4pla5kwaclyv99l27qlxcl"
  },
  "keyword": "",
  "category": 0,
  "page": 1,
  "size": 20
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "total": 2,
    "list": [
      {
        "account": "0001.bit",
        "status": 6,
        "expired_at": 1822199174000,
        "registered_at": 1632983174000,
        "enable_sub_account": 0,
        "renew_sub_account_price": 12,
        "nonce": 2
      },
      {
        "account": "10086.bit",
        "status": 8,
        "expired_at": 1662546730000,
        "registered_at": 1629196330000,
        "enable_sub_account": 0,
        "renew_sub_account_price": 12,
        "nonce": 2
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/mine -d'{"type":"blockchain","key_info":{"coin_type":"309","key":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqumxlsr77xk84ppzmm2rlmfvam37gc2t75yqwdn0cpl0rtr6ss3da4pla5kwaclyv99l27qlxcl"},"keyword":"","category":0,"page":1,"size":20}'
```
#### Account Detail

**Request**

* path: /v1/account/detail
* param:

```json
{
  "account": "king.bit"
}
```

**Response**
* status
  * -1: not open for registration
  * 0: can be registered
  * 1: payment confirmation
  * 2: application for registration
  * 3: pre-registration
  * 4: proposal
  * 5: confirming the proposal
  * 6: has been registered
  * 7: reserve account
  * 8: on sale
  * 9: auction
  * 13: unregisterable
  * 14: sub-account
  * 15: cross-chain
  * 17: on dutch auction period
  * 18: on dutch auction deliver period
  * 19: account in approval transfer
* re_registered_time: Time for re registration
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "account": "king.bit",
    "owner": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
    "owner_chain_type": 1,
    "owner_coin_type": "60",
    "manager": "0x3919a8eb619ccae32fba88d333829929db2f4324",
    "manager_chain_type": 1,
    "manager_coin_type": "60",
    "registered_at": 1632983024000,
    "expired_at": 1664519024000,
    "status": 6,
    "account_price": "10",
    "base_amount": "3.89",
    "confirm_proposal_hash": "0xec7bec47a4d3ad467253925a7e097f311e0738d625d55f8b3420cabaaa9b5201",
    "enable_sub_account": 1,
    "renew_sub_account_price": 12,
    "nonce": 2,
    "custom_script": "",
    "premium_percentage": "0.036",
    "premium_base": "0.52",
    "re_register_time": 0
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/detail -d'{"account":"king.bit"}'
```

#### Account Records

**Request**

* path: /v1/account/records
* param:

```json
{
  "account": "king.bit"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "records": [
      {
        "key": "twitter",
        "type": "profile",
        "label": "",
        "value": "111",
        "ttl": "300"
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/records -d'{"account":"king.bit"}'
```

#### Reverse Latest (Deprecated)

**Request**

* path: /v1/reverse/latest
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "account": "9aaaaaaa.bit",
    "is_valid": true
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/reverse/latest -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad"}'
```

#### Reverse List (Deprecated)

**Request**

* path: /v1/reverse/list
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "list": [
      {
        "account": "9aaaaaaa.bit",
        "block_number": 3752755,
        "hash": "0x9b6d4eee5c32f9b4aa52a1188e035d5afe695fbea2d90504d9d62bc869bd5ca8",
        "index": 0
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/reverse/list -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad"}'
```


#### Transaction Status

**Request**

* path: /v1/transaction/status
* param:
  * actions: business transaction type
    * ActionWithdrawFromWallet TxAction = 0  // withdraw
    * ActionConsolidateIncome  TxAction = 1  // rewards
    * ActionStartAccountSale   TxAction = 2  // on sale
    * ActionEditAccountSale    TxAction = 3  // edit sale
    * ActionCancelAccountSale  TxAction = 4  // cancel sale
    * ActionBuyAccount         TxAction = 5  // buy account
    * ActionSaleAccount        TxAction = 6  // sale account
    * ActionTransferBalance    TxAction = 7  // transfer balance
    * ActionDeclareReverseRecord   TxAction = 8  // declare reverse
    * ActionRedeclareReverseRecord TxAction = 9  // edit reverse
    * ActionRetractReverseRecord   TxAction = 10 // delete reverse
    * ActionDasBalanceTransfer TxAction = 11 // das balance pay
    * ActionEditRecords        TxAction = 12 // edit records
    * ActionTransferAccount    TxAction = 13 // edit owner
    * ActionEditManager        TxAction = 14 // edit manager
    * ActionRenewAccount       TxAction = 15 // renew

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "actions": [
    8,
    9
  ]
}
```

**Response**
  * status: tx pending

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "block_number": 0,
    "hash": "0x0343c250842fc57daef9fc30e5b9e1270c753a43215db46b19563c588417fcae",
    "action": 9,
    "status": 0
  }
}
```

```json
{
  "err_no": 11001,
  "err_msg": "not exits tx",
  "data": null
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/transaction/status -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","actions":[8,9]}'
```

#### Balance Info

**Request**

* path: /v1/balance/info
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "transfer_address": "ckt1qyq9k9w0anm56x8denhq3u6cvag637tzs68sn6f99z"
}
```

**Response**

```json
{
  "data": {
    "das_lock_amount": 1105535870712,
    "transfer_address_amount": 20000000000
  },
  "err_msg": "",
  "err_no": 0
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/balance/info -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","transfer_address":"ckt1qyq9k9w0anm56x8denhq3u6cvag637tzs68sn6f99z"}'
```

#### Transaction List

**Request**

* path: /v1/transaction/list
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "page": 1,
  "size": 2
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "total": 239,
    "list": [
      {
        "hash": "0x0b762bcd7679365be06696f7a8ff59472bc28b1294ee55374e840ee500f72436",
        "block_number": 4034830,
        "action": 2,
        "account": "tangzhihong008.bit",
        "capacity": 20100000000,
        "timestamp": 1641905797301
      },
      {
        "hash": "0xeb88df17e43a83a17ca2d98413060e54553da3422b736afa0ea88259048e0db1",
        "block_number": 4034012,
        "action": 0,
        "account": "",
        "capacity": 12000000000,
        "timestamp": 1641899346929
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/transaction/list -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","page":1,"size":2}'
```

#### Rewards Mine

**Request**

* path: /v1/rewards/mine
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "page": 1,
  "size": 2
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "count": 14,
    "total": "81725000000",
    "list": [
      {
        "invitee": "9baaaaaa.bit",
        "invitation_time":1719418521,
        "reward": "5149000000"
      },
      {
        "invitee": "9caaaaaa.bit",
        "invitation_time":1719418521,
        "reward": "12872500000"
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/rewards/mine -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","page":1,"size":2}'
```

#### Withdraw List

**Request**

* path: /v1/withdraw/list
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "page": 1,
  "size": 2
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "count": 52,
    "total": "34346803135995",
    "list": [
      {
        "hash": "0xeb88df17e43a83a17ca2d98413060e54553da3422b736afa0ea88259048e0db1",
        "block_number": 4034012,
        "amount": "12000000000"
      },
      {
        "hash": "0xd4b619fbbddd7bcd08170f922d79dbc86f6b1ef6131425d64b538fa85b11ac52",
        "block_number": 4034007,
        "amount": "12000000000"
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/withdraw/list -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","page":1,"size":2}'
```

#### Account Search

**Request**

* path: /v1/account/search
* param:
  * account_char_str <sup>[1](https://github.com/dotbitHQ/cell-data-generator/tree/master/data)</sup>: 
    * char_set_name: 0-emoji,1-digit,2-en,3-Chinese Simplified,4-Chinese Traditional,5-Japanese,6-Korean,7-Russian,8-Turkish,9-Thai,10-Vietnamese
    * char: account name single character
```json
{
  "account": "aaaa.bit",
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1, // 1-evm, 3-tron, 7-doge
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account_char_str": [
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "."
    },
    {
      "char_set_name": 2,
      "char": "b"
    },
    {
      "char_set_name": 2,
      "char": "i"
    },
    {
      "char_set_name": 2,
      "char": "t"
    }
  ]
}
```

**Response**
  * status: 
    * -1: Not open for registration
    * 0: Can be registered
    * 1: confirm payment
    * 2: send apply tx
    * 3: send pre tx
    * 4: send propose tx
    * 5: send confirm propose tx
    * 6: registered
    * 7: reserved account
    * 8: on sale at did.top
    * 13: unregisterable
    * 14: unminted subaccount
    * 15: waiting for cross-chain eth nft
    * 19: account in approval transfer

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "status": -1,
    "account": "aaaa.bit",
    "account_price": "0", // account price
    "base_amount": "0", // basic storage fee of ckb cell
    "is_self": false, // whether the register tx info is the current address
    "register_tx_map": {// tx info for each step
      "1": {
        "chain_id":1, // 0-ckb, 1-evm, 3-tron, 7-doge
        "hash": "", // pay hash
        "token_id": "eth_eth" //
      }, 
      "2": {}, // apply tx hash
      "3": {}, // pre tx hash
      "4": {},// propose tx hash
      "5": {}// confirm propose tx hash
    },
    "open_timestamp": 0, // 
    "premium_percentage": "",// for stripe usd premium
    "premium_base": "" // for stripe usd premium
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/search -d'{"account":"aaaa.bit","chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account_char_str":[{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}]}'
```

#### Account Registering List

**Request**

* path: /v1/account/registering/list
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "registering_accounts": [
      {
        "account": "",
        "status": 1,
        "cross_coin_type": "309"
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/registering/list -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad"}'
```

#### Account Order Detail

**Request**

* path: /v1/account/order/detail
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1, // 1-evm, 3-tron, 7-doge
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad", 
  "account": "xasdaaxaaa.bit",
  "action": "pre_register
}
```

**Response**
  * status: 0-unpaid, 1-confirm payment, 2-account is being registered
  * coin_type<sup>[1](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)</sup>: 60-eth, 195-trx, 9006-bsc, 966-matic, doge-3
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "order_id": "780bb68a7dd3b0554d95d6e0b3ca3ef3",
    "account": "asxasadasx.bit",
    "action": "pre_register",
    "status": 0,
    "timestamp": 1642059562457, // order time
    "pay_token_id": "ckb_das", // payment token type
    "pay_amount": "50502165739", // payment amount, token minimum precision
    "pay_type": "",
    "receipt_address": "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q", // user payment address
    "inviter_account": "", // inviter account of order
    "channel_account": "", // channel account of order
    "register_years": 1, // number of years of account registration
    "coin_type": "", // used to init account records
    "cross_coin_type": "", // used to cross chain
    "contract_address": "",// for usdt
    "client_secret": ""// for stripe usd
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/detail -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"xasdaaxaaa.bit","action":"apply_register"}'
```

#### Address Deposit

**Request**

* path: /v1/address/deposit
* param:
    * algorithm_id: 3-evm, 5-712, 4-tron, 6-Ed25519

```json
{
  "algorithm_id": 6,
  "address": "0x111..."
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "ckb_address": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/address/deposit -d'{"algorithm_id":6,"address":"0xe1090ce82474cbe0b196d1e62ec349ec05a61076c68d14129265370ca7e051c4"}'
```

#### Character Set List

**Request**

* path: /v1/character/set/list
* param:

```json
{
  "char_type": 0
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "emoji_list": [
      "",
      ""
    ],
    "digit_list": [
      "",
      ""
    ],
    "en_list": [
      "",
      ""
    ],
    "ko_list": [
      "",
      ""
    ],
    "vi_list": [
      "",
      ""
    ],
    "th_list": [
      "",
      ""
    ],
    "tr_list": [
      "",
      ""
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/character/set/list -d'{}'
```
#### Account Auction Info

**Request**

* path: /v1/account/auction/info
  * get dutch auction info of a account
* param:

```json
{
  "account":"michaeltest1.bit",
  "type":"blockchain",
  "key_info":{
    "coin_type":"60",
    "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "account_id": "",
    "account": "",
    "bid_status": 2,
    "hash": "",
    "start_auction_time": 0,
    "end_auction_time": 0,
    "expired_at": 0,
    "account_price": "5",
    "base_amount": "0.82"
  }
}
```

**Usage**

```curl
curl --location 'http://127.0.0.1:8120/v1/account/auction/info' \
--header 'Content-Type: application/json' \
--data '{
    "account":"michaeltest1.bit",
    "type":"blockchain",
    "key_info":{
        "coin_type":"60",
        "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
    }
}'
```


#### Account Auction Price

**Request**

* path: /v1/account/auction/price
  * get dutch auction price of a account
* param:

```json
{
  "account":"michaeltest1.bit"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "account_price": "5",
    "base_amount": "0.82",
    "premium_price": 20
  }
}
```

**Usage**

```curl
curl --location 'http://127.0.0.1:8120/v1/account/auction/price' \
--header 'Content-Type: application/json' \
--data '{
    "account":"michaeltest1.bit"
}'
```

#### Account Auction OrderStatus

**Request**

* path: /v1/account/auction/order-status
  * get  status of a dutch auction order
* param:

```json
{
  "account":"michaeltest1.bit",
  "hash": "0xb9e094dd6fcaa6c68d44233cb5331e63bd966fa86659fc45d30089336021f26e",
  "type":"blockchain",
  "key_info":{
    "coin_type":"60",
    "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "account": "michaeltest1.bit",
    "hash": "0xeb4871b7af2ca7129a43c5991c408148abd195eb5699223fad11a712b1e1d584",
    "status": 0,
    "basic_price": "6",
    "premium_price": "100"
  }
}
```

**Usage**

```curl
curl --location 'http://127.0.0.1:8120/v1/account/auction/order-status' \
--header 'Content-Type: application/json' \
--data '{
    "account":"michaeltest1.bit",
    "type":"blockchain",
    "key_info":{
        "coin_type":"60",
        "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
    }
}'
```

#### Account Auction PendingOrders

**Request**

* path: /v1/account/auction/pending-order
  * get dutch auction order with pending status
* param:

```json
{
  "type":"blockchain",
  "key_info":{
    "coin_type":"60",
    "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": [
    {
      "account": "michaeltest1.bit",
      "outpoint": "0xeb4871b7af2ca7129a43c5991c408148abd195eb5699223fad11a712b1e1d584-0",
      "status": 0,
      "basic_price": "6",
      "premium_price": "100"
    }
  ]
}
```

**Usage**

```curl
curl --location 'http://127.0.0.1:8120/v1/account/auction/pending-order' \
--header 'Content-Type: application/json' \
--data '{
    "type":"blockchain",
    "key_info":{
        "coin_type":"60",
        "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
    }
}'
```

#### Account Recommend

**Request**

* path: /v1/account/recommend
  * get recommend account list
* param:

```json
{
  "account":"goadgame.bit",
  "page":1,
  "size":6
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "total_page": 2,
    "page": 1,
    "acc_list": [
      "goadgame.bit",
      "loadgame.bit",
      "roadgame.bit",
      "toadgame.bit",
      "goalgame.bit",
      "goasgame.bit"
    ]
  }
}
```

**Usage**

```curl
curl --location 'http://localhost:8120/v1/account/recommend' \
--header 'Content-Type: application/json' \
--data '{
    "account":"goad.bit"
}'
```
#### Account Check Coupon

**Request**

* path: /v1/account/recommend
  * get recommend account list
* param:

```json
{
  "coupon":""
}
```

**Response**
* type:
  * 1: 4 
  * 2: 5
* status:
  * 1: Not found
  * 2: Used
  * 3: Expired
  * 4: Available
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "type": 2,
    "status": 1
  }
}
```

**Usage**

```curl
curl --location 'http://localhost:8120/v1/check/coupon' \
--header 'Content-Type: application/json' \
--data '{
    "coupon":""
}'
```


### OPERATE API LIST


#### Reverse Declare (Deprecated)

**Request**

* path: /v1/reverse/declare
* param:
  * evm_chain_id: eth-1/5 bsc-56/97 polygon-137/8001

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "aaaaa.bit",
  "evm_chain_id": 5
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/reverse/declare -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"aaaa.bit","evm_chain_id":5}'
```

#### Reverse Redeclare (Deprecated)

**Request**

* path: /v1/reverse/redeclare
* param:
```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "aaaaa.bit",
  "evm_chain_id": 5
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/reverse/redeclare -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"9aaaaaaa.bit","evm_chain_id":5}'
```

#### Reverse Retract (Deprecated)

**Request**

* path: /v1/reverse/retract
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "evm_chain_id": 5
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/reverse/retract -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","evm_chain_id":5}'
```

#### Transaction Send

**Request**

* path: /v1/transaction/send
* param:
  * sign_address: The actual address of the signature（Unnecessary，Used for webauthn algorithm）
```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 5,
      "sign_msg": ""
    }
  ],
  "sign_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggquyxv8jked54atrex9zwks38g48fy73vdsyqwzrxretvk62743unz38tggn52n5j0gkxcmk8jru"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "hash": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/transaction/send -d'{"sign_key": "","sign_list": [{"sign_type": 5,"sign_msg": ""}]}'
```
#### Balance Pay

**Request**

* path: /v1/balance/pay
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "evm_chain_id": 97,
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "order_id": "d2ccdcee9a7c163efb50d4a808a3d3f1"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/balance/pay -d'{"evm_chain_id":97,"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","order_id":"d2ccdcee9a7c163efb50d4a808a3d3f1"}'
```

#### Balance Withdraw

**Request**

* path: /v1/balance/withdraw
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "receiver_address": "ckt1qsexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6q7f75a3mpf4ddsy20uxwcgg3rvf5zmx0tgre86nk8v9x44kq3flsemppzyd3xstveadktaqfa",
  "withdraw_all": false,
  "amount": "20000000000",
  "evm_chain_id": 5
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/balance/withdraw -d'{"evm_chain_id":5,"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","receiver_chain_type":0,"receiver_address":"ckt1qsexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6q7f75a3mpf4ddsy20uxwcgg3rvf5zmx0tgre86nk8v9x44kq3flsemppzyd3xstveadktaqfa","withdraw_all":false,"amount":"20000000000"}'
```

#### Balance Transfer

**Request**

* path: /v1/balance/transfer
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "transfer_address": ""
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/balance/transfer -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad"}'
```

#### Balance Deposit

**Request**

* path: /v1/balance/deposit
* param:

```json
{
  "from_ckb_address": "",
  "to_ckb_address": "",
  "amount": ""
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 6,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/balance/deposit -d'{"from_ckb_address":"","to_ckb_address":"","amount":20000000000}'
```

#### Account Edit Manager

**Request**

* path: /v1/account/edit/manager
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "evm_chain_id": 5,
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "0001.bit",
  "raw_param": {
    "manager_chain_type": 1,
    "manager_coin_type": "60",
    "manager_address": "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/edit/manager -d'{"evm_chain_id":5,"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"0001.bit","raw_param":{"manager_chain_type":1,"manager_address":"0xc9f53b1d85356B60453F867610888D89a0B667Ad"}}'
```

#### Account Edit Owner

**Request**

* path: /v1/account/edit/owner
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "0001.bit",
  "evm_chain_id": 5,
  "raw_param": {
    "receiver_chain_type": 1,
    "receiver_coin_type": "60",
    "receiver_address": "0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/edit/owner -d'{"evm_chain_id":5,"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"0001.bit","raw_param":{"receiver_chain_type":1,"receiver_address":"0x15a33588908cF8Edb27D1AbE3852Bf287Abd3891"}}'
```

#### Account Edit Records

**Request**

* path: /v1/account/edit/records
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "evm_chain_id": 5,
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "11111111.bit",
  "raw_param": {
    "records": [
      {
        "type": "profile",
        "key": "twitter",
        "label": "",
        "value": "111",
        "ttl": "300",
      }
    ]
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/edit/records -d'{"evm_chain_id":5,"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"11111111.bit","raw_param":{"records":[{"type":"profile","key":"twitter","label":"","value":"111","ttl":"300","action":"add"}]}}'
```

#### Account Order Renew

**Request**

* path: /v1/account/order/renew
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "11111111.bit",
  "pay_chain_type": 0,
  "pay_token_id": "ckb_das",
  "pay_type": "",
  "pay_address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "renew_years": 1
}

```

**Response**

```json
{
  "errno": 0,
  "errmsg": "",
  "data": {
    "order_id": "278d079c57ec940f84a518085ba99894",
    "token_id": "ckb_das",
    "receipt_address": "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q",
    "amount": "28498885953",
    "contract_address": "",
    "client_secret": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/renew -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"0001.bit","pay_chain_type":0,"pay_token_id":"ckb_das","pay_type":"","pay_address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","renew_years":1}'
```



#### Account Order Register

**Request**

* path: /v1/account/order/register
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "asxasadasx.bit",
  "pay_chain_type": 0,
  "pay_token_id": "ckb_das",
  "pay_address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "pay_type": "",
  "register_years": 1,
  "inviter_account": "",
  "channel_account": "",
  "account_char_str": [
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "s"
    },
    {
      "char_set_name": 2,
      "char": "x"
    },
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "s"
    },
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "d"
    },
    {
      "char_set_name": 2,
      "char": "a"
    },
    {
      "char_set_name": 2,
      "char": "s"
    },
    {
      "char_set_name": 2,
      "char": "x"
    },
    {
      "char_set_name": 2,
      "char": "."
    },
    {
      "char_set_name": 2,
      "char": "b"
    },
    {
      "char_set_name": 2,
      "char": "i"
    },
    {
      "char_set_name": 2,
      "char": "t"
    }
  ],
  "coin_type": "",
  "cross_coin_type": "",
  "gift_card": ""
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "order_id": "780bb68a7dd3b0554d95d6e0b3ca3ef3",
    "token_id": "ckb_das",
    "receipt_address": "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q",
    "amount": "50502165739",
    "contract_address": "",// for usdt
    "client_secret": ""// for stripe usd
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/register -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"asxasadasx.bit","pay_chain_type":0,"pay_token_id":"ckb_das","pay_address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","pay_type":"","register_years":1,"inviter_account":"","channel_account":"","account_char_str":[{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"x"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"d"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"x"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}]}'
```

#### Account Order Change

**Request**

* path: /v1/account/order/change
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "asxasadasx.bit",
  "pay_chain_type": 0,
  "pay_token_id": "ckb_das",
  "pay_address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "pay_type": "",
  "register_years": 2,
  "inviter_account": "",
  "channel_account": "",
  "coin_type": "",
  "cross_coin_type": ""
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "order_id": "cf81838049ccb007eac0ce0872193dc3",
    "token_id": "ckb_das",
    "receipt_address": "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q",
    "amount": "76945148487",
    "code_url": "",
    "pay_type": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/change -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"asxasadasx.bit","pay_chain_type":0,"pay_token_id":"ckb_das","pay_address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","pay_type":"","register_years":2,"inviter_account":"","channel_account":""}'
```

#### Account Order Pay Hash

**Request**

* path: /v1/account/order/pay/hash
* param:

```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "60",
    "key": "0x111..."
  },
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "asxasadasx.bit",
  "order_id": "17b05429826dc39016a6f9e4de9c55ba",
  "pay_hash": "0xfcc6eca382311be8702862f36c0863e27c4f63beedd9ff786b8413558be14559"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": null
}
```

```json
{
  "err_no": 30006,
  "err_msg": "order not exist",
  "data": null
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/pay/hash -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"asxasadasx.bit","order_id":"17b05429826dc39016a6f9e4de9c55ba","pay_hash":"0xfcc6eca382311be8702862f36c0863e27c4f63beedd9ff786b8413558be14559"}'
```

#### Account Register

**Request**

* (Internal Service Api)
* path: /v1/account/register
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "1234.bit",
  "account_char_str": [
    {
      "char_set_name": 1,
      "char": "1"
    },
    {
      "char_set_name": 1,
      "char": "2"
    },
    {
      "char_set_name": 1,
      "char": "3"
    },
    {
      "char_set_name": 1,
      "char": "4"
    },
    {
      "char_set_name": 2,
      "char": "."
    },
    {
      "char_set_name": 2,
      "char": "b"
    },
    {
      "char_set_name": 2,
      "char": "i"
    },
    {
      "char_set_name": 2,
      "char": "t"
    }
  ],
  "register_years": 1,
  "inviter_account": "",
  "channel_account": ""
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "order_id": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8119/v1/account/register -d'{"chain_type":1,"address":"0xc9f53b1d85356B60453F867610888D89a0B667Ad","account":"1234567887.bit","account_char_str":[{"char_set_name":1,"char":"1"},{"char_set_name":1,"char":"2"},{"char_set_name":1,"char":"3"},{"char_set_name":1,"char":"4"},{"char_set_name":1,"char":"5"},{"char_set_name":1,"char":"6"},{"char_set_name":1,"char":"7"},{"char_set_name":1,"char":"8"},{"char_set_name":1,"char":"8"},{"char_set_name":1,"char":"7"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}],"register_years":1,"inviter_account":"","channel_account":""}'
```

#### Account Renew

**Request**

* (Internal Service Api)
* path: /v1/account/renew
* param:

```json
{
  "chain_type": 1,
  "address": "0x111...",
  "account": "121134.bit",
  "renew_years": 1
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "order_id": "97ef7068d2c6534092dc496808e8b760"
  }
}
```

```curl
curl -X POST http://127.0.0.1:8119/v1/account/renew -d'{"chain_type":1,"address":"0xc9f53b1d85356B60453F867610888D89a0B667Ad","account":"1234567887.bit","renew_years":1}'
```

### NODE RPC

#### Node Ckb Rpc

**Request**

* path: /v1/node/ckb/rpc
* param:

```json
{
  "jsonrpc": "2.0",
  "id": 2976777,
  "method": "get_blockchain_info",
  "params": []
}
```

```json
{
  "id": 0,
  "jsonrpc": "2.0",
  "method": "get_cells",
  "params": [
    {
      "script": {
        "code_hash": "0x58c5f491aba6d61678b7cf7edf4910b1f5e00ec0cde2f42e0abb4fd9aff25a63",
        "args": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
        "hash_type": "type"
      },
      "script_type": "lock",
      "filter": {
        "output_data_len_range": [
          "0x0",
          "0x1"
        ]
      }
    },
    "asc",
    "0x100",
    null
  ]
}
```

```json

{
  "id": 0,
  "jsonrpc": "2.0",
  "method": "send_transaction",
  "params": []
}
```

**Response**

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/node/ckb/rpc -d'{"jsonrpc":"2.0","id":2976777,"method":"get_blockchain_info","params":[]}'
```






#### Account Auction Bid

**Request**

* path: /v1/account/auction/bid
  * bid a account during dutch auction period
* param:

```json
{
  "account":"michaeltest1.bit",
  "type":"blockchain",
  "key_info":{
    "coin_type":"60",
    "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
  }
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "",
    "sign_list": [
      {
        "sign_type": 5,
        "sign_msg": ""
      }
    ],
    "mm_json": {}
  }
}
```

**Usage**

```curl
curl --location 'http://127.0.0.1:8120/v1/account/auction/bid' \
--header 'Content-Type: application/json' \
--data '{
    "account":"michaeltest1.bit",
    "type":"blockchain",
    "key_info":{
        "coin_type":"60",
        "key":"0xd437b8e9cA16Fce24bF3258760c3567214213C5A"
    }
}'
```