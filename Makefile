# macOS：间接依赖 gotron-sdk → rjeczalik/notify 使用 CGO，会触发 FSEvents 弃用告警，用 CGO_CFLAGS 关闭。
export CGO_CFLAGS ?= -Wno-deprecated-declarations

.PHONY: build run test test-btc-integration

build:
	go build -o wallet .

test:
	go test ./...

# 需本地 bitcoind；仅调试 BTC RPC 时使用
test-btc-integration:
	go test -tags=integration -count=1 ./btc/

run:
	go run .
