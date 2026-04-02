# Swagger Example API
[toc]

## 通用说明

- **Base URL**：默认 `http://127.0.0.1:10001`（以配置 `app.port` 为准）。
- **鉴权**：所有 `/api/*` 请求须在请求头携带 **`x-token`**（任意非空字符串；示例：`dev`）。
- **响应**：HTTP 状态码一般为 `200`，业务是否成功看 JSON 中的 **`code`**（`0` 表示成功）。

## 1	环境变量

### 默认环境1
| 参数名 | 字段值 |
| ------ | ------ |



## 2	Swagger Example API

##### 说明
> 



##### 联系方式
- **联系人：**API Support
- **邮箱：**support@swagger.io
- **网址：**/http://www.swagger.io/support/

##### 文档版本
```
1.0
```


## 3	归集

> POST  /api/collection
### 请求体(Request Body)
| 参数名称 | 数据类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| address|string||false|地址|
| coinName|string||false|币种名称|
| max|string||false|最大归集数量（满足当前值才会归集）|
| protocol|string||false|协议|
### 响应体
● 200 响应数据格式：JSON
| 参数名称 | 类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| data|object||false||
|⇥ balance|string||false|实际归集的数量|
| server.Response|object||false||
|⇥ code|int32||false|错误code码|
|⇥ data|object||false|成功时返回的对象|
|⇥ message|string||false|错误信息|

##### 接口描述
> 




## 4	创建钱包地址

> POST  /api/createWallet
### 请求体(Request Body)
| 参数名称 | 数据类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| coinName|string||false|币种名称|
| protocol|string||false|协议|
### 响应体
● 200 响应数据格式：JSON
| 参数名称 | 类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| data|object||false||
|⇥ address|string||false|生成的钱包地址|
| server.Response|object||false||
|⇥ code|int32||false|错误code码|
|⇥ data|object||false|成功时返回的对象|
|⇥ message|string||false|错误信息|

##### 接口描述
> 




## 5	删除钱包地址

> POST  /api/delWallet
### 请求体(Request Body)
| 参数名称 | 数据类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| address|string||false|地址|
| coinName|string||false|币种名称|
| protocol|string||false|协议|
### 响应体
● 200 响应数据格式：JSON
| 参数名称 | 类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| code|int32||false|错误code码|
| data|object||false|成功时返回的对象|
| message|string||false|错误信息|

##### 接口描述
> 




## 6	获取交易结果

> GET  /api/getTransactionReceipt

**说明**：路由注册为 **GET**，但 Handler 使用 **`ShouldBindJSON`**，请求需带 **JSON 请求体**（curl 示例见文末）。若使用浏览器或部分客户端，可改为与 Handler 一致的绑定方式（或封装为 POST）。

### 请求体(Request Body)
| 参数名称 | 数据类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| coinName|string||false|币种名称|
| hash|string||false|交易哈希|
| protocol|string||false|协议|
### 响应体
● 200 响应数据格式：JSON
| 参数名称 | 类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| data|object||false||
|⇥ status|int32||false|交易状态（0：未成功，1：已成功）|
| server.Response|object||false||
|⇥ code|int32||false|错误code码|
|⇥ data|object||false|成功时返回的对象|
|⇥ message|string||false|错误信息|

##### 接口描述
> 




## 7	提现

> POST  /api/withdraw
### 请求体(Request Body)
| 参数名称 | 数据类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| address|string||false|提现地址|
| coinName|string||false|币种名称|
| orderId|string||false|订单号|
| protocol|string||false|协议|
| value|int32||false|金额|
### 响应体
● 200 响应数据格式：JSON
| 参数名称 | 类型 | 默认值 | 不为空 | 描述 |
| ------ | ------ | ------ | ------ | ------ |
| data|object||false||
|⇥ hash|string||false|生成的交易hash|
| server.Response|object||false||
|⇥ code|int32||false|错误code码|
|⇥ data|object||false|成功时返回的对象|
|⇥ message|string||false|错误信息|

##### 接口描述
> 

## 8	curl 一键验证（与上文接口对应）

先启动服务，例如：

```shell
go run . -c config/dev.yml
```

另开终端执行（默认访问 `http://127.0.0.1:10001`，`eth` + `ETH` 引擎）：

```shell
bash script/curl_api.sh
```

或：`make curl-api`。

环境变量可选：`API_BASE`、`API_TOKEN`、`PROTOCOL`、`COIN_NAME`。

**单条 curl 示例**（创建钱包）：

```shell
curl -s -X POST "http://127.0.0.1:10001/api/createWallet" \
  -H "Content-Type: application/json" \
  -H "x-token: dev" \
  -d '{"protocol":"eth","coinName":"ETH"}'
```

**获取交易结果**（GET + JSON 体，与实现一致）：

```shell
curl -s -X GET "http://127.0.0.1:10001/api/getTransactionReceipt" \
  -H "Content-Type: application/json" \
  -H "x-token: dev" \
  -d '{"protocol":"eth","coinName":"ETH","hash":"0x0000000000000000000000000000000000000000000000000000000000000000"}'
```

**返回码说明（curl 冒烟常见）**

| 接口 | code=0 时 | 常见非 0 |
|------|-----------|----------|
| getTransactionReceipt | 能解析请求；`data.status`：1 成功，0 未上链/无回执 | 节点 RPC 异常等 |
| collection | 归集逻辑执行完毕；`data.balance` 为归集数量（未达门槛可为 `"0"`） | **10004**：本地无该充值地址；**10001**：多为公网 RPC 读余额失败（可重试） |
| withdraw | 链上已发起转账 | **10001**：热钱包无余额 / Gas 不足等 |

