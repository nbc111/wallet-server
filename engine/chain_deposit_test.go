// Package engine 中的本文件实现「多链资产存入（理财入口）」的单元测试与链上集成自检。
//
// # 需求与用例对应关系
//
// 业务需求：
//  1. 支持链：波场（TRC20）、币安智能链（BEP20）、以太坊（ERC20）—— 通过 protocol / contract / rpc 区分；
//  2. 资产类型：USDT、BTC、ETH 等 —— 通过 coin_name + 合约（或主币空合约）区分；
//  3. 用户存入后资金归集至财务指定地址 —— 每条 engines 配置独立的 collection_address（每链/每资产一条引擎时可配置为「每链一个财务地址」）。
//
// 测试分层：
//   - 无链上 RPC：归集逻辑、未达标不转账等纯单元测试；
//   - 需外网 RPC（默认跑 make test / go test 时会执行，可用 -short 跳过）：各链 CreateWallet 地址格式。
package engine

import (
	"math/big"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lmxdawn/wallet/config"
	"github.com/lmxdawn/wallet/types"
)

// stubWorker 模拟链 Worker，仅记录 Transfer 的收款方，用于验证归集是否指向财务地址。
type stubWorker struct {
	balance        *big.Int
	lastTransferTo string
	transferCalls  int
}

func (s *stubWorker) GetNowBlockNum() (uint64, error) { return 0, nil }
func (s *stubWorker) GetTransaction(uint64) ([]types.Transaction, uint64, error) {
	return nil, 0, nil
}
func (s *stubWorker) GetTransactionReceipt(*types.Transaction) error { return nil }
func (s *stubWorker) GetBalance(string) (*big.Int, error)            { return s.balance, nil }
func (s *stubWorker) CreateWallet() (*types.Wallet, error)           { return nil, nil }
func (s *stubWorker) Transfer(_ string, toAddress string, _ *big.Int, _ uint64) (string, string, uint64, error) {
	s.transferCalls++
	s.lastTransferTo = toAddress
	return "", "", 0, nil
}

// TestDeposit_Collection_SendsToFinanceAddressPerChain 验证需求「用户存入资产经归集后转入财务提供的指定地址」。
//
// 实现要点：每条引擎配置 collection_address；归集时 Worker.Transfer 的 to 必须为该地址。
// 多链场景下应为不同引擎分别配置财务地址（例如以太坊一条、BSC 一条、波场一条），本测试用表驱动模拟三种典型地址形态。
func TestDeposit_Collection_SendsToFinanceAddressPerChain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		financeAddr   string // 模拟「该链财务收款地址」
		userAddr      string
		minCollection *big.Int
	}{
		{
			name:          "以太坊主网/ERC20_财务地址为0x",
			financeAddr:   "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			userAddr:      "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			minCollection: big.NewInt(50),
		},
		{
			name:          "币安链_BEP20_财务地址同为EVM格式",
			financeAddr:   "0xcccccccccccccccccccccccccccccccccccccccc",
			userAddr:      "0xdddddddddddddddddddddddddddddddddddddddd",
			minCollection: big.NewInt(1),
		},
		{
			name:          "波场_TRC20_财务地址为Base58_T开头",
			financeAddr:   "TYBpNWMNHPyk1hBbY1SRr2RQebF39iFk7z",
			userAddr:      "TXYZopYRdj2D9XRtbDm411ZedZH4Wf3U7E",
			minCollection: big.NewInt(10),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stub := &stubWorker{balance: big.NewInt(100)}
			eng := ConCurrentEngine{
				Worker: stub,
				config: config.EngineConfig{
					CollectionAddress: tc.financeAddr,
				},
			}
			got, err := eng.collection(tc.userAddr, "priv", tc.minCollection)
			if err != nil {
				t.Fatal(err)
			}
			if got.Cmp(big.NewInt(100)) != 0 {
				t.Fatalf("归集成功时应返回当前可转余额: got %s", got)
			}
			if stub.lastTransferTo != tc.financeAddr {
				t.Fatalf("归集收款方必须为该引擎配置的财务地址: want %q got %q", tc.financeAddr, stub.lastTransferTo)
			}
			if stub.transferCalls != 1 {
				t.Fatalf("应只发起一笔归集转账: calls=%d", stub.transferCalls)
			}
		})
	}
}

// TestDeposit_Collection_SkipsWhenBalanceBelowThreshold 验证：余额未达到归集门槛时不调用 Transfer（避免无效上链）。
func TestDeposit_Collection_SkipsWhenBalanceBelowThreshold(t *testing.T) {
	stub := &stubWorker{balance: big.NewInt(10)}
	eng := ConCurrentEngine{
		Worker: stub,
		config: config.EngineConfig{
			CollectionAddress: "0x1111111111111111111111111111111111111111",
		},
	}
	got, err := eng.collection("0xuser", "priv", big.NewInt(100))
	if err != nil {
		t.Fatal(err)
	}
	if got.Sign() != 0 {
		t.Fatalf("未达门槛应返回 0: got %s", got)
	}
	if stub.transferCalls != 0 {
		t.Fatalf("未达门槛不应发起 Transfer: calls=%d", stub.transferCalls)
	}
}

func defaultTestEngineConfig(t *testing.T, coinName, protocol, contract, contractType, rpc string) config.EngineConfig {
	t.Helper()
	return config.EngineConfig{
		CoinName:            coinName,
		Protocol:            protocol,
		Contract:            contract,
		ContractType:        contractType,
		Rpc:                 rpc,
		File:                filepath.Join(t.TempDir(), coinName),
		WalletPrefix:        "wallet-",
		HashPrefix:          "hash-",
		BlockInit:           0,
		BlockAfterTime:      60,
		ReceiptCount:        1,
		ReceiptAfterTime:    60,
		CollectionAfterTime: 60,
		CollectionCount:     1,
		CollectionMax:       "0",
		CollectionAddress:   "0x0000000000000000000000000000000000000001",
		Confirms:            1,
		RechargeNotifyUrl:   "http://127.0.0.1:9/recharge",
		WithdrawNotifyUrl:   "http://127.0.0.1:9/withdraw",
		WithdrawPrivateKey:  "61774dacba914e5675eef6c616df85c61d7c8917f56ee77f547a140f8f982d31",
		Network:             "",
		User:                "",
		Pass:                "",
	}
}

// TestIntegration_Deposit_CreateWallet_MultiChain 验证「各链可生成用户充值地址」，对应理财入口侧为用户分配链上地址。
//
// 覆盖：
//   - 以太坊链 ERC20（USDT 主网合约示例）
//   - 币安链 BEP20（USDT BSC 合约 + BSC RPC；与 ERC20 共用 eth Worker）
//   - 波场链 TRC20（USDT 主网合约 + Tron gRPC）
//   - 以太坊主币 ETH
//   - 比特币主网 BTC
//
// 依赖公网 RPC / gRPC；CI 或离线请使用：go test -short ./engine/
func TestIntegration_Deposit_CreateWallet_MultiChain(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过链上集成测试（无 RPC）：go test 不带 -short 时再跑")
	}

	// 子用例命名与需求字段对齐，便于测试报告与需求追溯。
	cases := []struct {
		name string
		// requirement 说明该用例对应的需求条目（文档/验收用）
		requirement string
		cfg         config.EngineConfig
		ok          func(addr string) bool
	}{
		{
			name:        "以太坊链_主币_ETH",
			requirement: "资产类型：ETH；链：以太坊",
			cfg:         defaultTestEngineConfig(t, "ETH", "eth", "", "", "https://cloudflare-eth.com"),
			ok: func(addr string) bool {
				return strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) == 42
			},
		},
		{
			name:        "以太坊链_ERC20_USDT",
			requirement: "资产类型：USDT；链：以太坊；标准：ERC20",
			cfg:         defaultTestEngineConfig(t, "USDT-ERC20", "eth", "0xdAC17F958D2ee523a2206206994597C13D831ec7", "erc20", "https://cloudflare-eth.com"),
			ok: func(addr string) bool {
				return strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) == 42
			},
		},
		{
			name:        "币安链_BEP20_USDT",
			requirement: "资产类型：USDT；链：币安智能链；标准：BEP20（protocol=eth + BSC RPC）",
			cfg:         defaultTestEngineConfig(t, "USDT-BEP20", "eth", "0x55d398326f99059fF775485246999027B3197955", "erc20", "https://bsc-dataseed.binance.org/"),
			ok: func(addr string) bool {
				return strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) == 42
			},
		},
		{
			name:        "波场链_TRC20_USDT",
			requirement: "资产类型：USDT；链：波场；标准：TRC20",
			cfg:         defaultTestEngineConfig(t, "USDT-TRC20", "trx", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", "trc20", "grpc.trongrid.io:50051"),
			ok: func(addr string) bool {
				return strings.HasPrefix(addr, "T") && len(addr) >= 26
			},
		},
		{
			name:        "比特币_BTC",
			requirement: "资产类型：BTC（Bitcoin 主网，非 EVM 上封装 BTC）",
			cfg: func() config.EngineConfig {
				c := defaultTestEngineConfig(t, "BTC", "btc", "", "", "127.0.0.1:18332")
				c.Network = "MainNet"
				c.User = "user"
				c.Pass = "pass"
				return c
			}(),
			ok: func(addr string) bool {
				return strings.HasPrefix(addr, "1") || strings.HasPrefix(addr, "3") || strings.HasPrefix(addr, "bc1")
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("需求: %s", tc.requirement)
			eng, err := NewEngine(tc.cfg)
			if err != nil {
				t.Skipf("无法连接链节点（网络/防火墙）: %v", err)
			}
			defer eng.db.Close()
			addr, err := eng.CreateWallet()
			if err != nil {
				t.Fatalf("CreateWallet: %v", err)
			}
			if !tc.ok(addr) {
				t.Fatalf("充值地址格式不符合该链预期: %q", addr)
			}
		})
	}
}
