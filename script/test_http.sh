#!/usr/bin/env bash
# 启动真实 HTTP 服务，用 curl 验证 POST /api/createWallet（与 engine/chain_deposit_test、server/http_deposit_integration_test 一致的多链场景）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export CGO_CFLAGS="-Wno-deprecated-declarations ${CGO_CFLAGS:-}"

PORT="${HTTP_TEST_PORT:-10009}"
CONFIG="${HTTP_TEST_CONFIG:-config/http-integration.yml}"
BASE="http://127.0.0.1:${PORT}/api"
TOKEN="${HTTP_TEST_TOKEN:-http-test-token}"

if command -v lsof >/dev/null 2>&1; then
  lsof -ti:"${PORT}" | xargs kill -9 2>/dev/null || true
fi

echo "启动服务: go run . -c ${CONFIG} (端口 ${PORT})"
go run . -c "${CONFIG}" &
PID=$!
trap 'kill "${PID}" 2>/dev/null || true' EXIT

for i in $(seq 1 90); do
  if curl -sf -o /dev/null -X POST "${BASE}/createWallet" \
    -H "Content-Type: application/json" \
    -H "x-token: ${TOKEN}" \
    -d '{"protocol":"eth","coinName":"ETH"}' 2>/dev/null; then
    break
  fi
  if ! kill -0 "${PID}" 2>/dev/null; then
    echo "进程已退出，启动失败"
    exit 1
  fi
  sleep 1
done

expect_ok() {
  local proto=$1 coin=$2 label=$3
  local resp
  resp=$(curl -s -X POST "${BASE}/createWallet" \
    -H "Content-Type: application/json" \
    -H "x-token: ${TOKEN}" \
    -d "{\"protocol\":\"${proto}\",\"coinName\":\"${coin}\"}")
  if echo "${resp}" | grep -q '"code":0'; then
    echo "OK  ${label}"
  else
    echo "FAIL ${label}: ${resp}"
    exit 1
  fi
}

expect_ok eth ETH "以太坊 ETH"
expect_ok eth USDT-ERC20 "以太坊 USDT ERC20"
expect_ok eth USDT-BEP20 "币安链 USDT BEP20"
expect_ok trx USDT-TRC20 "波场 USDT TRC20"
expect_ok btc BTC "比特币 BTC"

echo "全部 HTTP createWallet 检查通过。"
