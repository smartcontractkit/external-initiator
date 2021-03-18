package services

import (
	"math/big"
	"net/url"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNewFluxMonitor(t *testing.T) {
	cryptoapis, _ := url.Parse("http://localhost:8081")
	coingecko, _ := url.Parse("http://localhost:8082")
	amberdata, _ := url.Parse("http://localhost:8083")
	NewFluxMonitor([]url.URL{*cryptoapis, *coingecko, *amberdata}, "BTC", "USD", new(big.Int), decimal.NewFromFloat(0.01), decimal.NewFromInt(0), 15*time.Second, 5*time.Second)
	// time.Sleep(20 * time.Second)
	// fm.stop()
}
