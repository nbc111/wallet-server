#!/usr/bin/env bash
# 多链存入能力自检：TRC20 / BEP20 / ERC20 / ETH / BTC + 归集目标为财务地址
set -euo pipefail
cd "$(dirname "$0")/.."
# 抑制 macOS 上 rjeczalik/notify（CGO）的 FSEvents 弃用告警
export CGO_CFLAGS="-Wno-deprecated-declarations ${CGO_CFLAGS:-}"
go test -count=1 -timeout 10m ./engine/ -run 'TestDeposit_|TestIntegration_Deposit' "$@"
