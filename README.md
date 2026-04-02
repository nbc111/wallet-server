# wallet

> 虚拟币钱包服务，转账/提现/充值/归集
>
>
> 完全实现与业务服务隔离，使用http服务相互调用


## 多链与主流币种（TRC20 / BEP20 / ERC20）

每种链上的每种资产需要 **单独一条** `engines` 配置；HTTP 接口用 `protocol` + `coinName` 区分引擎（与 `coin_name` 一致）。

| 链 | 代币标准 | `protocol` | `contract` | `contract_type` | `rpc` |
|----|-----------|--------------|------------|-----------------|--------|
| 波场 | TRC20（如 USDT） | `trx` | T 开头的合约地址 | `trc20` | **gRPC** 地址（如 `grpc.trongrid.io:50051`）；与仅提供 **JSON-RPC** 的 HTTP 网关（如部分 Tatum 地址）**不兼容** |
| 币安智能链 | BEP20（如 USDT） | `eth` | `0x` 合约地址 | `erc20`（示例里可写） | **BSC** 的 JSON-RPC 地址 |
| 以太坊 | ERC20（如 USDT） | `eth` | `0x` 合约地址 | `erc20` | **以太坊** JSON-RPC |
| 以太坊 | 主币 ETH | `eth` | 留空 | 留空 | 以太坊 JSON-RPC |
| 比特币 | 主网 BTC | `btc` | 留空 | 留空 | 本地 bitcoind（`host:port`）或 **HTTPS** JSON-RPC 网关（如 Tatum：`https://bitcoin-mainnet.gateway.tatum.io`，`user`/`pass` 按网关要求填 API Key）；`network` 填 `MainNet` 等 |

说明：**BEP20 与 ERC20 共用 `protocol: eth`**，通过 `rpc` 指向 BSC 或以太坊即可区分链；`coin_name` 建议用业务可读的名称（如 `USDT`、`USDT-BSC`、`USDT-TRC`），避免同名冲突。

# 接口文档

`script/api.md`（curl 冒烟：`make curl-api` 或 `bash script/curl_api.sh`，需先启动服务）

# 下载-打包

```shell
# 拉取代码
$ git clone https://github.com/lmxdawn/wallet.git
$ cd wallet

# 打包 (-tags "doc") 可选，加上可以运行swagger
$ go build [-tags "doc"]

# 本地运行（默认使用 config/dev.yml：以太坊公开 RPC，无需 bitcoind）
$ go mod download
$ go run .

# 或使用完整示例配置（默认含 BTC，需本地 bitcoind 时区块同步才正常）
$ go run . -c config/config-example.yml

```

## 本地运行说明

1. **依赖**：Go 1.17+；`go mod download` 拉取模块。
2. **默认端口**：`10001`（见 `app.port`）。
3. **API 鉴权**：请求头需带 **`x-token`**（任意非空字符串即可，中间件未校验具体值）。
4. **验证**：`go run .` 后执行：
   ```shell
   curl -s -X POST http://127.0.0.1:10001/api/createWallet \
     -H "Content-Type: application/json" -H "x-token: dev" \
     -d '{"protocol":"eth","coinName":"ETH"}'
   ```
   返回 `code:0` 且含 `address` 即服务正常。
5. **`config/config-example.yml`** 中启用了 **BTC** 引擎并指向 `127.0.0.1:18332`；未启动 bitcoind 时区块监听会重试，但 HTTP 仍可启动（`createWallet` 等部分接口仍可用）。

## 多链存入自检

验证 **TRC20 / BEP20 / ERC20**、**USDT / BTC / ETH** 创建地址，以及归集时资金发往配置的 **`collection_address`（财务地址）**：

```shell
bash script/test_multi_chain.sh
```

等价：`go test -count=1 ./engine/ -run 'TestDeposit_|TestIntegration_Deposit'`（多链存入/归集财务地址相关）。集成测试需访问公网 RPC；仅跑单元测试可用 `go test -short ./engine/`（会跳过链上集成用例）。

## HTTP 服务验证（与单元测试等价场景）

在**真实 HTTP** 上验证 `POST /api/createWallet`（`x-token` + `protocol` + `coinName`），覆盖与 `engine/chain_deposit_test.go` 相同的多链用例：

| 方式 | 说明 |
|------|------|
| **Shell** | `make test-http` 或 `bash script/test_http.sh`：读取 `config/http-integration.yml`（默认端口 **10009**），启动进程后对 ETH / USDT-ERC20 / USDT-BEP20 / USDT-TRC20 / BTC 依次 curl。 |
| **Go httptest** | `go test ./server/ -run 'TestIntegration_HTTP|TestHTTP_'`（集成用例需外网 RPC；`TestHTTP_CreateWallet_NoToken` 可随时跑）。 |

环境变量（可选）：`HTTP_TEST_PORT`、`HTTP_TEST_CONFIG`、`HTTP_TEST_TOKEN`。

# 重新生成配置

```shell
# 生成配置文件
$ vim config.yml
$ wallet -c config.yml

```

# 配置文件参数解释

|  参数名   | 描述  |
|  ----  | ----  |
| coin_name  | 币种名称 |
| contract  | 合约地址（为空表示主币） |
| contract_type  | 合约类型（波场需要区分是TRC20还是TRC10） |
| protocol  | 协议名称 |
| network  | 网络名称（暂时BTC协议有用{MainNet：主网，TestNet：测试网，TestNet3：测试网3，SimNet：测试网}） |
| rpc  | rpc配置 |
| user  | rpc用户名（没有则为空） |
| pass  | rpc密码（没有则为空） |
| file  | db文件路径配置 |
| wallet_prefix  | 钱包的存储前缀 |
| hash_prefix  | 交易哈希的存储前缀 |
| block_init  | 初始块（默认读取最新块） |
| block_after_time  | 获取最新块的等待时间 |
| receipt_count  | 交易凭证worker数量 |
| receipt_after_time  | 获取交易信息的等待时间 |
| collection_after_time  | 归集等待时间 |
| collection_count  | 归集发送worker数量 |
| collection_max  | 最大的归集数量（满足多少才归集，为0表示不自动归集） |
| collection_address  | 归集地址 |
| confirms  | 确认数量 |
| recharge_notify_url  | 充值通知回调地址 |
| withdraw_notify_url  | 提现通知回调地址 |
| withdraw_private_key  | 提现的私钥地址 |

> 启动后访问： `http://localhost:10009/swagger/index.html`


# Swagger

> 把 swag cmd 包下载 `go get -u github.com/swaggo/swag/cmd/swag`

> 这时会在 bin 目录下生成一个 `swag.exe` ，把这个执行文件放到 `$GOPATH/bin` 下面

> 执行 `swag init` 注意，一定要和main.go处于同一级目录

> 启动时加上 `-tags "doc"` 才会启动swagger。 这里主要为了正式环境去掉 swagger，这样整体编译的包小一些

> 启动后访问： `http://ip:prot/swagger/index.html`

# 第三方库依赖

> log 日志 `github.com/rs/zerolog`

> 命令行工具 `github.com/urfave/cli`

> 配置文件 `github.com/jinzhu/configor`

# 环境依赖

> go 1.16+

> **macOS 编译告警**：若出现 `github.com/rjeczalik/notify` 与 `FSEventStreamScheduleWithRunLoop` 的弃用提示，来自间接依赖（gotron-sdk → keystore）的 CGO，**可忽略**。请优先使用 **`make build` / `make test`**（Makefile 已设置 `CGO_CFLAGS=-Wno-deprecated-declarations`）；若直接执行 `go test ./...`，可先：`export CGO_CFLAGS=-Wno-deprecated-declarations`。
>
> **Bitcoin 包测试**：`btc/*_test.go` 为需本地 **bitcoind** 的集成测试，已用 **`//go:build integration`** 默认不编译。跑全量测试请用 **`make test`**。仅在有节点且需测 RPC 时：`go test -tags=integration ./btc/`（个别 macOS 版本若仍报 `dyld ... LC_UUID`，属工具链/系统已知问题，可跳过该包测试）。

> Redis 3

> MySQL 5.7

# 其它

> `script/Generate MyPOJOs.groovy` 生成数据库Model

# 合约相关

> `solcjs.cmd --version` 查看版本
>
> `solcjs.cmd --abi erc20.sol`
>
> `abigen --abi=erc20_sol_IERC20.abi --pkg=eth --out=erc20.go`

# 准备

要实现这些功能首先得摸清楚我们需要完成些什么东西

1. 获取最新区块
2. 获取区块内部的交易记录
3. 通过交易哈希获取交易的完成状态
4. 获取某个地址的余额
5. 创建一个地址
6. 签名并发送luo交易
7. 定义接口如下

```go
type Worker interface {
getNowBlockNum() (uint64, error)
getTransaction(uint64) ([]types.Transaction, uint64, error)
getTransactionReceipt(*types.Transaction) error
getBalance(address string) (*big.Int, error)
createWallet() (*types.Wallet, error)
sendTransaction(string, string, *big.Int) (string, error)
}
```

# 实现

> 创建一个地址后把地址和私钥保存下来

## 进

通过一个无限循环的服务不停的去获取最新块的交易数据，并且把交易数据都一一验证是否完成 ，这里判断数据的接收地址（to）是否属于本服务创建的钱包地址，如果是本服务的创建过的地址则判断为充值成功，**（这时逻辑服务里面需要做交易哈希做幂等）**

## 出

用户发起一笔提出操作，用户发起提出时通过服务配置的私钥来打包并签名luo交易。（私钥转到用户输入的提出地址），这里把提交的luo交易的哈希记录到服务 通过一个无限循环的服务不停的去获取最新块的交易数据，并且把交易数据都一一验证是否完成
，这里判断交易数据的哈希是否存在于服务，如果存在则处理**（这时逻辑服务里面需要做交易哈希做幂等）**

## 归集

通过定期循环服务创建的地址去转账到服务配置的归集地址里面，这里需要注意归集数量的限制，当满足固定的数量时才去归集（减少gas费）

# 一个简单的示例

github地址： [golang 实现加密货币的充值/提现/归集服务](https://github.com/lmxdawn/wallet)

# 特别说明

> 创建钱包的方式可以用 create2 创建合约，这样可以实现不用批量管理私钥，防止私钥丢失或者被盗。

- solidity

```solidity
// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import "./ERC20.sol";

contract Wallet {
    address internal token = 0xDA0bab807633f07f013f94DD0E6A4F96F8742B53;
    address internal hotWallet = 0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2;

    constructor() {
        // send all tokens from this contract to hotwallet
        IERC20(token).transfer(
            hotWallet,
            IERC20(token).balanceOf(address(this))
        );
        // selfdestruct to receive gas refund and reset nonce to 0
        selfdestruct(payable(hotWallet));
    }
}

contract Fabric {
    function createContract(uint256 salt) public returns (address newAddr){
        // get wallet init_code
        bytes memory bytecode = type(Wallet).creationCode;
        assembly {
            let codeSize := mload(bytecode) // get size of init_bytecode
            newAddr := create2(
                0, // 0 wei
                add(bytecode, 32), // the bytecode itself starts at the second slot. The first slot contains array length
                codeSize, // size of init_code
                salt // salt from function arguments
            )
        }
    }
    function getAddress(uint _salt)
        public
        view
        returns (address)
    {
        bytes memory bytecode = type(Wallet).creationCode;
        bytes32 hash = keccak256(
            abi.encodePacked(bytes1(0xff), address(this), _salt, keccak256(bytecode))
        );

        // NOTE: cast last 20 bytes of hash to address
        return address(uint160(uint(hash)));
    }

    function getBytecode() public pure returns (bytes memory) {
        bytes memory bytecode = type(Wallet).creationCode;

        return bytecode;
    }

    
    function getBytecode1() public pure returns (bytes1) {
        
        return bytes1(0xff);
    }

    
    function getBytecode3(uint256 s) public pure returns (bytes memory) {
        
        return abi.encodePacked(s);
    }
    
    
    function getBytecode2() public pure returns (bytes32) {
        bytes memory bytecode = type(Wallet).creationCode;

        return keccak256(bytecode);
    }
}
```

- go

```go
code := "6080604052600080546001600160a01b031990811673da0bab807633f07f013f94dd0e6a4f96f8742b53179091556001805490911673ab8483f64d9c6d1ecf9b849ae677dd3315835cb217905534801561005857600080fd5b506000546001546040516370a0823160e01b81523060048201526001600160a01b039283169263a9059cbb92169083906370a0823190602401602060405180830381865afa1580156100ae573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906100d29190610150565b6040516001600160e01b031960e085901b1681526001600160a01b03909216600483015260248201526044016020604051808303816000875af115801561011d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906101419190610169565b506001546001600160a01b0316ff5b60006020828403121561016257600080fd5b5051919050565b60006020828403121561017b57600080fd5b8151801515811461018b57600080fd5b939250505056fe"
codeB := common.Hex2Bytes(code)
codeHash := crypto.Keccak256Hash(codeB)
fmt.Println(codeHash)

address := common.HexToAddress("0x7EF2e0048f5bAeDe046f6BF797943daF4ED8CB47")
fmt.Println(address)

fmt.Println(common.LeftPadBytes(big.NewInt(1).Bytes(), 32))
var buffer bytes.Buffer
buffer.Write(common.FromHex("0xff"))
buffer.Write(address.Bytes())
buffer.Write(common.Hex2Bytes("0x30"))
buffer.Write(codeHash.Bytes())

hash := crypto.Keccak256Hash([]byte{0xff}, address.Bytes(), common.LeftPadBytes(big.NewInt(1).Bytes(), 32), codeHash.Bytes())

//salt := common.LeftPadBytes(big.NewInt(1).Bytes(), 32)
//crypto.CreateAddress2(address, salt, codeHash.Bytes())

fmt.Println(common.BytesToAddress(hash[12:]))
```
