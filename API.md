 * [Query API LIST](#query-api-list)
    * [Token List](#token-list)
    * [Config Info](#config-info)
    * [Account List](#account-list)
    * [Account Mine](#account-mine)
    * [Account Detail](#account-detail)
    * [Account Records](#account-records)
    * [Reverse Latest](#reverse-latest)
    * [Reverse List](#reverse-list)
    * [Transaction Status](#transaction-status)
    * [Balance Info](#balance-info)
    * [Transaction List](#transaction-list)
    * [Rewards Mine](#rewards-mine)
    * [Withdraw List](#withdraw-list)
    * [Account Search](#account-search)
    * [Account Registering List](#account-registering-list)
    * [Account Order Detail](#account-order-detail)
 * [OPERATE API LIST](#operate-api-list)
    * [Reverse Declare](#reverse-declare)
    * [Reverse Redeclare](#reverse-redeclare)
    * [Reverse Retract](#reverse-retract)
    * [Transaction Send](#transaction-send)
    * [Balance Withdraw](#balance-withdraw)
    * [Balance Transfer](#balance-transfer)
    * [Account Edit Manager](#account-edit-manager)
    * [Account Edit Owner](#account-edit-owner)
    * [Account Edit Records](#account-edit-records)
    * [Account Order Renew](#account-order-renew)
    * [Balance Pay](#balance-pay)
    * [Account Order Register](#account-order-register)
    * [Account Order Change](#account-order-change)
    * [Account Order Pay Hash](#account-order-pay-hash)
    * [Account Register](#account-register)  
 * [NODE RPC](#node-rpc)
    * [Node Ckb Rpc](#node-ckb-rpc)

### Query API LIST

#### Token List

**Request**

* path: /token/list
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

* path: /config/info
* param: none

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "reverse_record_capacity": 20100000000,
    "min_change_capacity": 11600000000,
    "sale_cell_capacity": 20100000000,
    "min_sell_price": 20000000000,
    "account_expiration_grace_period": 2592000,
    "min_ttl": 300,
    "profit_rate_of_inviter": "0.1",
    "inviter_discount": "0.05",
    "min_account_len": 4,
    "max_account_len": 42,
    "edit_records_throttle": 300,
    "edit_manager_throttle": 300,
    "transfer_throttle": 300,
    "income_cell_min_transfer_value": 11600000000,
    "premium": "0.1", 
    "timestamp_on_chain": 1647589995
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/config/info
```

#### Account List

**Request**

* path: /account/list
    * get user's not on sale accounts
* param:

```json
{
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
        "registered_at": 1624347772000
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

* path: /account/mine
    * get user's accounts by pagination
* param:

```json

{
  "chain_type": 1,
  "address": "0xc9f53b1d85356B60453F867610888D89a0B667Ad",
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
    "total": 47,
    "list": [
      {
        "account": "0001.bit",
        "status": 6,
        "expired_at": 1822199174000,
        "registered_at": 1632983174000
      },
      {
        "account": "10086.bit",
        "status": 8,
        "expired_at": 1662546730000,
        "registered_at": 1629196330000
      }
    ]
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/mine -d'{"chain_type":1,"address":"0xc9f53b1d85356B60453F867610888D89a0B667Ad","page":1,"size":2}'
```

#### Account Detail

**Request**

* path: /account/detail
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
    "account": "king.bit",
    "owner": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
    "owner_chain_type": 1,
    "manager": "0x3919a8eb619ccae32fba88d333829929db2f4324",
    "manager_chain_type": 1,
    "registered_at": 1632983024000,
    "expired_at": 1664519024000,
    "status": 6,
    "account_price": "10",
    "base_amount": "3.89",
    "confirm_proposal_hash": "0xec7bec47a4d3ad467253925a7e097f311e0738d625d55f8b3420cabaaa9b5201"
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/detail -d'{"account":"king.bit"}'
```


#### Account Records

**Request**

* path: /account/records
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

#### Reverse Latest

**Request**

* path: /reverse/latest
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

#### Reverse List

**Request**

* path: /reverse/list
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

* path: /transaction/status
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "actions": [
    8,
    9
  ]
}
```

**Response**

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

* path: /balance/info
* param:

```json
{
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

* path: /transaction/list
* param:

```json
{
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

* path: /rewards/mine
* param:

```json
{
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
        "reward": "5149000000"
      },
      {
        "invitee": "9caaaaaa.bit",
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

* path: /withdraw/list
* param:

```json
{
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

* path: /account/search
* param:

```json
{
  "account": "aaaa.bit",
  "chain_type": 1,
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

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "status": -1,
    "account": "aaaa.bit",
    "account_price": "0",
    "base_amount": "0",
    "is_self": false,
    "register_tx_map": {}
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/search -d'{"account":"aaaa.bit","chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account_char_str":[{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}]}'
```


#### Account Registering List

**Request**

* path: /account/registering/list
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
    "registering_accounts": [
      {
        "account": "",
        "status": 1
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

* path: /account/order/detail
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "xasdaaxaaa.bit",
  "action": "apply_register"
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "order_id": "780bb68a7dd3b0554d95d6e0b3ca3ef3",
    "account": "asxasadasx.bit",
    "action": "apply_register",
    "status": 0,
    "timestamp": 1642059562457,
    "pay_token_id": "ckb_das",
    "pay_amount": "50502165739",
    "pay_type": "",
    "receipt_address": "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q",
    "inviter_account": "",
    "channel_account": "",
    "register_years": 1,
    "code_url": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/detail -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"xasdaaxaaa.bit","action":"apply_register"}'
```

### OPERATE API LIST

#### Reverse Declare

**Request**

* path: /reverse/declare
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

#### Reverse Redeclare

**Request**

* path: /reverse/redeclare
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

#### Reverse Retract

**Request**

* path: /reverse/retract
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

* path: /transaction/send
* param:

```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 5,
      "sign_msg": ""
    }
  ]
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

#### Balance Withdraw

**Request**

* path: /balance/withdraw
* param:

```json
{
  "evm_chain_id": 5,
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "receiver_chain_type": 0,
  "receiver_address": "ckt1qsexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6q7f75a3mpf4ddsy20uxwcgg3rvf5zmx0tgre86nk8v9x44kq3flsemppzyd3xstveadktaqfa",
  "withdraw_all": false,
  "amount": "20000000000"
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

* path: /balance/transfer
* param:

```json
{
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

#### Account Edit Manager

**Request**

* path: /account/edit/manager
* param:

```json
{
  "evm_chain_id": 5,
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "0001.bit",
  "raw_param": {
    "manager_chain_type": 1,
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

* path: /account/edit/owner
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "0001.bit",
  "evm_chain_id": 5,
  "raw_param": {
    "receiver_chain_type": 1,
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

* path: /account/edit/records
* param:

```json
{
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
        "action": "add"
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

* path: /account/order/renew
* param:

```json
{
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
    "code_url": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/renew -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"0001.bit","pay_chain_type":0,"pay_token_id":"ckb_das","pay_type":"","pay_address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","renew_years":1}'
```


#### Balance Pay

**Request**

* path: /balance/pay
* param:

```json
{
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


#### Account Order Register

**Request**

* path: /account/order/register
* param:

```json
{
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
  ]
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
    "code_url": "",
    "pay_type": ""
  }
}
```

**Usage**

```curl
curl -X POST http://127.0.0.1:8120/v1/account/order/register -d'{"chain_type":1,"address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","account":"asxasadasx.bit","pay_chain_type":0,"pay_token_id":"ckb_das","pay_address":"0xc9f53b1d85356b60453f867610888d89a0b667ad","pay_type":"","register_years":1,"inviter_account":"","channel_account":"","account_char_str":[{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"x"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"d"},{"char_set_name":2,"char":"a"},{"char_set_name":2,"char":"s"},{"char_set_name":2,"char":"x"},{"char_set_name":2,"char":"."},{"char_set_name":2,"char":"b"},{"char_set_name":2,"char":"i"},{"char_set_name":2,"char":"t"}]}'
```


#### Account Order Change

**Request**

* path: /account/order/change
* param:

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "asxasadasx.bit",
  "pay_chain_type": 0,
  "pay_token_id": "ckb_das",
  "pay_address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "pay_type": "",
  "register_years": 2,
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

* path: /account/order/pay/hash
* param:

```json
{
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
* path: /account/register
* param:


```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "account": "1234.bit",
  "account_char_str": [
     {"char_set_name":1,"char":"1"},
     {"char_set_name":1,"char":"2"},
     {"char_set_name":1,"char":"3"},
     {"char_set_name":1,"char":"4"},
     {"char_set_name":2,"char":"."},
     {"char_set_name":2,"char":"b"},
     {"char_set_name":2,"char":"i"},
     {"char_set_name":2,"char":"t"}
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

### NODE RPC

#### Node Ckb Rpc

**Request**

* path: /node/ckb/rpc
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

