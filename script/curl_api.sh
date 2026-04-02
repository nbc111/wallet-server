#!/usr/bin/env bash
# 对照 script/api.md 用 curl 冒烟测试各接口（需已启动服务，默认 config/dev.yml 端口 10001）。
# 用法：先另开终端执行 `go run . -c config/dev.yml`，再执行：
#   bash script/curl_api.sh
# 环境变量：API_BASE（默认 http://127.0.0.1:10001）、API_TOKEN、PROTOCOL、COIN_NAME
set -euo pipefail

API_BASE="${API_BASE:-http://127.0.0.1:10001}"
API_TOKEN="${API_TOKEN:-dev}"
PROTOCOL="${PROTOCOL:-eth}"
COIN_NAME="${COIN_NAME:-ETH}"
HDR_TOKEN=(-H "x-token: ${API_TOKEN}")
HDR_JSON=(-H "Content-Type: application/json")

json_addr() {
  local json=$1
  if command -v jq >/dev/null 2>&1; then
    echo "${json}" | jq -r '.data.address // empty'
    return
  fi
  if command -v python3 >/dev/null 2>&1; then
    echo "${json}" | python3 -c "import sys,json; d=json.load(sys.stdin); print((d.get('data') or {}).get('address') or '')"
    return
  fi
  echo "${json}" | sed -n 's/.*\"address\":\"\\([^\"]*\\)\".*/\\1/p' | head -1
}

need_http_200() {
  local name=$1
  local body=$2
  if ! echo "${body}" | grep -q '"code"'; then
    echo "FAIL ${name}: 非 JSON 或缺少 code: ${body}" >&2
    return 1
  fi
  echo "OK   ${name}"
}

echo "=== api.md 接口验证（${API_BASE}，protocol=${PROTOCOL} coinName=${COIN_NAME}）==="
echo "（需已启动服务，例如：go run . -c config/dev.yml）"

# 4 创建钱包 POST /api/createWallet
echo "--- §4 createWallet ---"
RESP=$(curl -s -X POST "${API_BASE}/api/createWallet" "${HDR_JSON[@]}" "${HDR_TOKEN[@]}" \
  -d "{\"protocol\":\"${PROTOCOL}\",\"coinName\":\"${COIN_NAME}\"}")
echo "${RESP}"
ADDR=$(json_addr "${RESP}")
need_http_200 "createWallet" "${RESP}"
if [[ -z "${ADDR}" ]]; then
  echo "未解析到 address，后续依赖地址的接口将跳过" >&2
fi

# 6 获取交易结果 GET /api/getTransactionReceipt（JSON Body）
# 交易不存在或未上链：code=0，data.status=0（非错误）
echo "--- §6 getTransactionReceipt ---"
FAKE_HASH="0x0000000000000000000000000000000000000000000000000000000000000000"
REC=$(curl -s -X GET "${API_BASE}/api/getTransactionReceipt" "${HDR_JSON[@]}" "${HDR_TOKEN[@]}" \
  -d "{\"protocol\":\"${PROTOCOL}\",\"coinName\":\"${COIN_NAME}\",\"hash\":\"${FAKE_HASH}\"}")
echo "${REC}"
need_http_200 "getTransactionReceipt" "${REC}"

# 3 归集 POST /api/collection
# 成功：code=0 且 data.balance 为实际归集量；余额未达 max 时常为 "0"。
# 钱包不存在：code=10004（没有数据）。公网 RPC 偶发失败：code=10001，可重试。
echo "--- §3 collection ---"
if [[ -n "${ADDR}" ]]; then
  COL=$(curl -s -X POST "${API_BASE}/api/collection" "${HDR_JSON[@]}" "${HDR_TOKEN[@]}" \
    -d "{\"protocol\":\"${PROTOCOL}\",\"coinName\":\"${COIN_NAME}\",\"address\":\"${ADDR}\",\"max\":\"1\"}")
  echo "${COL}"
  need_http_200 "collection" "${COL}"
else
  echo "SKIP collection（无 address）"
fi

# 7 提现 POST /api/withdraw（热钱包需有主币/ Gas；测试环境常见 code=10001，属预期）
echo "--- §7 withdraw ---"
WITH=$(curl -s -X POST "${API_BASE}/api/withdraw" "${HDR_JSON[@]}" "${HDR_TOKEN[@]}" \
  -d "{\"protocol\":\"${PROTOCOL}\",\"coinName\":\"${COIN_NAME}\",\"orderId\":\"curl-smoke-$$\",\"address\":\"0x0000000000000000000000000000000000000001\",\"value\":1}")
echo "${WITH}"
need_http_200 "withdraw" "${WITH}"

# 5 删除钱包 POST /api/delWallet
echo "--- §5 delWallet ---"
if [[ -n "${ADDR}" ]]; then
  DEL=$(curl -s -X POST "${API_BASE}/api/delWallet" "${HDR_JSON[@]}" "${HDR_TOKEN[@]}" \
    -d "{\"protocol\":\"${PROTOCOL}\",\"coinName\":\"${COIN_NAME}\",\"address\":\"${ADDR}\"}")
  echo "${DEL}"
  need_http_200 "delWallet" "${DEL}"
else
  echo "SKIP delWallet（无 address）"
fi

echo "=== 完成 ==="
echo "code=0：业务成功。§6 假哈希：status=0 表示链上尚无回执。"
echo "§3 若 code=10001 多为公网 RPC 瞬时错误；§7 无热钱包余额时常为非 0。"
