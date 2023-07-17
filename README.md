   * [Prerequisites](#prerequisites)
   * [Install &amp; Run](#install--run)
      * [Source Compile](#source-compile)
      * [Docker](#docker)
   * [Usage](#usage)
      * [Register](#register)
      * [Set Reverse Record](#set-reverse-record)
      * [Others](#others)
   * [Documents](#documents)
   
# das-register

Backend of DAS registration service. You can use this repo to build your own DAS registration website (as like https://app.did.id)

## Prerequisites

* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* Redis >= 5.0 (for cache) 
* GO version >= 1.17.10
* [ckb-node](https://github.com/nervosnetwork/ckb) (Must be synced to latest height and add `Indexer` module to ckb.toml)
* [das-database](https://github.com/dotbitHQ/das-database)
* [unipay](https://github.com/dotbitHQ/unipay) (Payment service used for registered accounts)
* If the version of the dependency package is too low, please install `gcc-multilib` (apt install gcc-multilib)
* Machine configuration: 4c8g200G

## Install & Run

### Source Compile
```bash
# get the code
git clone https://github.com/dotbitHQ/das-register.git

# edit config/config.yaml before init mysql database
mysql -uroot -p
> source das-register/tables/das_register_db.sql;
> quit;

# compile and run
cd das-register
make register
./das_register --config=config/config.yaml
```

### Docker
* docker >= 20.10
* docker-compose >= 2.2.2

```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.2.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
docker-compose up -d
```

_if you already have a mysql installed, just run_
```bash
docker run -dp 8119-8120:8119-8120 -v $PWD/config/config.yaml:/app/config/config.yaml --name das-register-server dotbitteam/das-register:latest
```

## Usage
You need to run [das-pay](https://github.com/dotbitHQ/das-pay) before you can run this service
### Register
* Use [register API](https://github.com/dotbitHQ/das-register/blob/main/API.md#account-order-register) get the `order ID`
* The server `das-pay` is monitoring the balance change of the receiving address on chain, and wait for user to pay with the `order ID` attached to the payment
* `Das-pay` will notify the `das-register` to start the registration process after the user's payment is completed
* Wait for `das-register` to complete the entire registration process

```
   +---------+                 +----------------+        +-----------+
   |   user  |                 |  das_register  |        |  das pay  |
   +----|----+                 +-------|--------+        +-----|-----+
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
`Das-register` will use user's das balance to set reverse record via API [reverse declare](https://github.com/dotbitHQ/das-register/blob/main/API.md#reverse-declare)

### Others
More APIs see [API.md](https://github.com/dotbitHQ/das-register/blob/main/API.md)

## Documents
* [What is DAS](https://github.com/dotbitHQ/das-contracts/blob/master/docs/en/Overview-of-DAS.md)
* [What is a DAS transaction on CKB](https://github.com/dotbitHQ/das-contracts/blob/master/docs/en/Data-Structure-and-Protocol/Transaction-Structure.md)

