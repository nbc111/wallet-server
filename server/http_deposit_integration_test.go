package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lmxdawn/wallet/config"
	"github.com/lmxdawn/wallet/engine"
)

func newDepositTestRouter(engines []*engine.ConCurrentEngine) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SetEngine(engines...))
	auth := r.Group("/api", AuthRequired())
	auth.POST("/createWallet", CreateWallet)
	return r
}

func testEngineConfig(t *testing.T, coinName, protocol, contract, contractType, rpc string) config.EngineConfig {
	t.Helper()
	return config.EngineConfig{
		CoinName:              coinName,
		Protocol:              protocol,
		Contract:              contract,
		ContractType:          contractType,
		Rpc:                   rpc,
		File:                  filepath.Join(t.TempDir(), coinName),
		WalletPrefix:          "wallet-",
		HashPrefix:            "hash-",
		BlockInit:             0,
		BlockAfterTime:        60,
		ReceiptCount:          1,
		ReceiptAfterTime:      60,
		CollectionAfterTime:   60,
		CollectionCount:       1,
		CollectionMax:         "0",
		CollectionAddress:     "0x0000000000000000000000000000000000000001",
		Confirms:              1,
		RechargeNotifyUrl:     "http://127.0.0.1:9/recharge",
		WithdrawNotifyUrl:     "http://127.0.0.1:9/withdraw",
		WithdrawPrivateKey:    "61774dacba914e5675eef6c616df85c61d7c8917f56ee77f547a140f8f982d31",
		Network:               "",
		User:                  "",
		Pass:                  "",
	}
}

// TestIntegration_HTTP_CreateWallet_MultiChain 通过 httptest 走真实 HTTP 路由，验证与 engine 包单元测试一致的多链 createWallet。
func TestIntegration_HTTP_CreateWallet_MultiChain(t *testing.T) {
	if testing.Short() {
		t.Skip("链上集成：go test 不带 -short")
	}

	cases := []struct {
		name     string
		cfg      config.EngineConfig
		protocol string
		coinName string
		okAddr   func(addr string) bool
	}{
		{
			name:     "以太坊_ETH",
			cfg:      testEngineConfig(t, "ETH", "eth", "", "", "https://cloudflare-eth.com"),
			protocol: "eth",
			coinName: "ETH",
			okAddr: func(addr string) bool {
				return strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) == 42
			},
		},
		{
			name: "以太坊_USDT_ERC20",
			cfg:  testEngineConfig(t, "USDT-ERC20", "eth", "0xdAC17F958D2ee523a2206206994597C13D831ec7", "erc20", "https://cloudflare-eth.com"),
			protocol: "eth",
			coinName: "USDT-ERC20",
			okAddr: func(addr string) bool {
				return strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) == 42
			},
		},
		{
			name: "币安_USDT_BEP20",
			cfg:  testEngineConfig(t, "USDT-BEP20", "eth", "0x55d398326f99059fF775485246999027B3197955", "erc20", "https://bsc-dataseed.binance.org/"),
			protocol: "eth",
			coinName: "USDT-BEP20",
			okAddr: func(addr string) bool {
				return strings.HasPrefix(strings.ToLower(addr), "0x") && len(addr) == 42
			},
		},
		{
			name: "波场_USDT_TRC20",
			cfg:  testEngineConfig(t, "USDT-TRC20", "trx", "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", "trc20", "grpc.trongrid.io:50051"),
			protocol: "trx",
			coinName: "USDT-TRC20",
			okAddr: func(addr string) bool {
				return strings.HasPrefix(addr, "T") && len(addr) >= 26
			},
		},
		{
			name: "比特币_BTC",
			cfg: func() config.EngineConfig {
				c := testEngineConfig(t, "BTC", "btc", "", "", "127.0.0.1:18332")
				c.Network = "MainNet"
				c.User = "user"
				c.Pass = "pass"
				return c
			}(),
			protocol: "btc",
			coinName: "BTC",
			okAddr: func(addr string) bool {
				return strings.HasPrefix(addr, "1") || strings.HasPrefix(addr, "3") || strings.HasPrefix(addr, "bc1")
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			eng, err := engine.NewEngine(tc.cfg)
			if err != nil {
				t.Skipf("无法连接链节点: %v", err)
			}
			defer eng.Close()

			r := newDepositTestRouter([]*engine.ConCurrentEngine{eng})
			body := map[string]string{"protocol": tc.protocol, "coinName": tc.coinName}
			raw, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/api/createWallet", bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-token", "test-token")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
			}
			var resp Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Code != 0 {
				t.Fatalf("业务 code=%d msg=%s", resp.Code, resp.Message)
			}
			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("data 类型: %T", resp.Data)
			}
			addr, _ := data["address"].(string)
			if addr == "" || !tc.okAddr(addr) {
				t.Fatalf("地址异常: %q", addr)
			}
		})
	}
}

// TestHTTP_CreateWallet_NoToken 验证鉴权：无 x-token 时拒绝。
func TestHTTP_CreateWallet_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SetEngine())
	auth := r.Group("/api", AuthRequired())
	auth.POST("/createWallet", CreateWallet)

	req := httptest.NewRequest(http.MethodPost, "/api/createWallet", bytes.NewReader([]byte(`{"protocol":"eth","coinName":"ETH"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP %d", w.Code)
	}
	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 10002 {
		t.Fatalf("期望 token 错误码 10002，得到 %d", resp.Code)
	}
}
