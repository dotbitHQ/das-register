   * [Prerequisites](#prerequisites)
   * [Install &amp; Run](#install--run)
   * [Usage](#usage)
      * [Register](#register)
      * [Set Reverse Record](#set-reverse-record)
      * [Others](#others)
   * [Documents](#documents)

# das-register

Backend of DAS registration service. You can use this repo to build your own DAS registration website, just like https://da.systems do

## Prerequisites

* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* GO version >= 1.15.0
* [CKB Node](https://github.com/nervosnetwork/ckb)
* [CKB Indexer](https://github.com/nervosnetwork/ckb-indexer)
* [das-database](https://github.com/DeAccountSystems/das-database)
* [das-pay](https://github.com/DeAccountSystems/das-pay)


## Install & Run

```bash
# get the code
git clone https://github.com/DeAccountSystems/das-register.git

# edit config/config.yaml before init mysql database
mysql -uroot -p
> source das-register/tables/das_register_db.sql;
> quit;

# compile and run
cd das-register
make register
./das_register --config=conf/config.yaml
```

## Usage
You need to run [das-pay](https://github.com/DeAccountSystems/das-pay) before you can run this service
### Register
* Use [register API](https://github.com/DeAccountSystems/das-register/blob/main/API.md#account-order-register) get the `order ID`
* The server `das-pay` is monitoring the balance change of the receiving address on chain, and wait for user to pay with the `order ID` attached to the payment
* `Das-pay` will notify the `das-register` to start the registration process after the user's payment is completed
* Wait for `das-register` to complete the entire registration process

```
   +---------+                 +----------------+        +-----------+
   |   user  |                 |  das_register  |        |  das pay  |
   +----+----+                 +-------+--------+        +-----+-----+
        |                              |                       |
        |                              |                       |
        +----- Get order id ---------->+                       |
        |                              |                       |
        |                              |                       |
        +<---- Return order id --------+                       |
        |                              |                       |
        |                              |                       |
Pay for the order                      |                       |
      on chain                         |                       |
        |                              |            Update the order status
        |                              |                       |
        |                              |                       |
        |                  Continue the registration           |
        |                              |                       |
        |                              |                       |
        |                              |                       |
        +                              +                       +

```

### Set Reverse Record
`Das-register` will use user's das balance to set reverse record via API [reverse declare](https://github.com/DeAccountSystems/das-register/blob/main/API.md#reverse-declare)

### Others
More APIs see [API.md](https://github.com/DeAccountSystems/das-register/blob/main/API.md)

## Documents
* [What is DAS](https://github.com/DeAccountSystems/das-contracts/blob/master/docs/en/Overview-of-DAS.md)
* [What is a DAS transaction on CKB](https://github.com/DeAccountSystems/das-contracts/blob/master/docs/en/Data-Structure-and-Protocol/Transaction-Structure.md)

