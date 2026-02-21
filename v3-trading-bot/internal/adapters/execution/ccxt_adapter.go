package execution

import (
	"fmt"
	"log"
	"v3-trading-bot/internal/core/domain"
	"v3-trading-bot/internal/core/ports"

	ccxt "github.com/ccxt/ccxt/go/v4"
)

type ccxtAdapter struct {
	exchange *ccxt.Binance
}

func NewCCXTAdapter(apiKey, secretKey string, isTestnet bool) (ports.ExchangePort, error) {
	// 1. Konfigürasyon
	config := map[string]interface{}{
		"apiKey": apiKey,
		"secret": secretKey,
	}

	// 2. Binance Başlatma
	exchange := ccxt.NewBinance(config)
	if exchange == nil {
		return nil, fmt.Errorf("binance nesnesi oluşturulamadı")
	}

	// 3. Testnet Modu
	if isTestnet {
		exchange.SetSandboxMode(true)
		log.Println("[CCXT] Sandbox (Testnet) modu aktif!")
	}

	// 4. Piyasaları Yükle (Koruma Kalkanı ile)
	// LoadMarkets ağ isteği yaptığı için panic olma ihtimaline karşı koruyoruz.
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("LoadMarkets sırasında kritik hata (panic): %v", r)
			}
		}()
		_, err = exchange.LoadMarkets()
	}()

	if err != nil { // hem panik hem normaal hatayı yakalamak için yazdık bunu.
		return nil, fmt.Errorf("piyasalar yüklenemedi: %w", err)
	}

	return &ccxtAdapter{
		exchange: exchange,
	}, nil
}

// CheckBalance - Çelik Yelekli Versiyon
// (Named Return Parameters kullanıyoruz ki recover bloğundan müdahale edebilelim: balance, err)
func (a *ccxtAdapter) CheckBalance(asset string) (balance float64, err error) {
	// AIRBAG: Olası bir panic durumunda programı kapatma, hatayı yakala.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ CCXT CheckBalance Panic Yakalandı: %v", r)
			err = fmt.Errorf("bakiye sorgusunda kütüphane hatası: %v", r)
			balance = 0
		}
	}()

	params := map[string]interface{}{}

	// 1. Bakiyeyi Çek
	balances, err := a.exchange.FetchBalance(params)
	if err != nil {
		return 0, fmt.Errorf("bakiye çekilemedi: %w", err)
	}

	// 2. Pointer Kontrolü (Senin kodun aynısı)
	if amount, ok := balances.Free[asset]; ok {
		if amount != nil {
			return *amount, nil
		}
	}

	return 0, nil
}

// ExecuteOrder
func (a *ccxtAdapter) ExecuteOrder(order *domain.Order) (result *domain.Order, err error) {
	// AIRBAG: Emir girerken kütüphane çökse bile sunucu ayakta kalsın.
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ CCXT ExecuteOrder Panic Yakalandı: %v", r)
			err = fmt.Errorf("emir gönderiminde kütüphane hatası: %v", r)
			result = nil
		}
	}()

	// (4 Parametreli Yapı Korundu)
	// options veya params göndermiyoruz, çünkü kütüphane versiyonun bunu desteklemiyor.
	response, err := a.exchange.CreateOrder(
		order.Symbol,
		order.OrderType,
		order.Side,
		order.Amount,
	)

	if err != nil {
		return nil, fmt.Errorf("CCXT emir hatası: %w", err)
	}

	if response.Id != nil {
		order.ExchangeReqID = *response.Id
	}

	if response.Status != nil {
		order.Status = *response.Status
	} else {
		order.Status = "OPEN"
	}

	return order, nil
}
