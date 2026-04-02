# macOS：间接依赖 gotron-sdk → rjeczalik/notify 使用 CGO，会触发 FSEvents 弃用告警，用 CGO_CFLAGS 关闭。
export CGO_CFLAGS ?= -Wno-deprecated-declarations

.PHONY: build run test test-http curl-api test-btc-integration

build:
	go build -o wallet .

test:
	go test ./...

# 启动 HTTP（config/http-integration.yml），curl 验证多链 createWallet；需公网 RPC
test-http:
	bash script/test_http.sh

# 对照 script/api.md 各接口（需已启动 dev 等服务，默认 10001）
curl-api:
	bash script/curl_api.sh

# 需本地 bitcoind；仅调试 BTC RPC 时使用
test-btc-integration:
	go test -tags=integration -count=1 ./btc/

run:
	go run .
