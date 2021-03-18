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
	NewFluxMonitor([]url.URL{*cryptoapis, *coingecko, *amberdata}, "BTC", "USD", new(big.Int), decimal.NewFromInt(1), decimal.NewFromInt(0), 1*time.Minute, 5*time.Second)
}
