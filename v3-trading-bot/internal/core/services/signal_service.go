package services

import (
	"errors"
	"fmt"
	"log"
	"time"
	"v3-trading-bot/internal/core/domain"
	"v3-trading-bot/internal/core/ports"

	"github.com/google/uuid"
)

type signalService struct {
	repo     ports.SignalRepository // db ile konusacagımız sözleşme
	exchange ports.ExchangePort     // ccxt ile konusacagımız sözleşme
}

func NewSignalService(repo ports.SignalRepository, exchange ports.ExchangePort) ports.SignalService {
	return &signalService{
		repo:     repo,
		exchange: exchange,
	}
}

func (s *signalService) ProcessSignal(signal *domain.Signal) error {
	// uuid
	signal.ID = uuid.New().String()

	log.Printf("[SERVICE] Yeni Sinyal Geldi: %s %s -Fiyat: %.2f - Skor: %.2f\n", signal.Action, signal.Symbol, signal.Price, signal.Score)

	// 1. Adım, sinyali db'ye "pending"-"bekliyor" olarak kaydet.
	signal.Status = "PENDING"
	if err := s.repo.SaveSignal(signal); err != nil {
		return fmt.Errorf("sinyal kaydedilmedi: %w", err)
	}

	// 2. Adım, risk ve strateji kontrol - iş mantıgı
	// sinyalden gelen skor pozitif fakat zayıfsa işleme girmemek adına bir filtre
	threshold := 0.15

	isWeakBuy := signal.Action == "BUY" && signal.Score < threshold
	isWeakSell := signal.Action == "SELL" && signal.Score > -threshold

	if isWeakBuy || isWeakSell {
		log.Printf("[SERVICE] Sinyal skoru zayıf (Aksiyon: %s, Skor: %.4f), işlem pas geçiliyor.\n", signal.Action, signal.Score)
		_ = s.repo.UpdateSignalStatus(signal.ID, "IGNORED")
		return nil
	}

	// 3. Adım, cüzdan bakiye kontrolü, switch kullanılabilir bu kısımda, revize edilecek
	baseCoin := "BTC"       // satmak için gerekli olan
	quoteCoin := "USDT"     // almak için gerekli olan
	var orderAmount float64 //ccxt'e gönderecegimiz miktar.
	if signal.Action == "BUY" {
		// buy işlemi, quotecoin yani usdt bakiyesini kontrol etmeliyiz.
		balance, err := s.exchange.CheckBalance(quoteCoin)
		if err != nil {
			_ = s.repo.UpdateSignalStatus(signal.ID, "FAILED")
			return fmt.Errorf("usdt bakiye kontrolü başarısız: %w", err)
		}
		if balance < 10.0 { // 10 usdt yoksa alım yapamayız.
			log.Println("[SERVICE] yetersiz usdt bakiyesi, işlem reddedildi.")
			_ = s.repo.UpdateSignalStatus(signal.ID, "REJECTED_NO_FUNDS")
			return errors.New("yetersiz bakiye")
		}
		// kural: usdt bakiyemizin maksimum %10 ile işlem acalım
		tradeAmountUSDT := (balance * 10) / 100
		if balance < 1000 {
			tradeAmountUSDT = balance / 2
		}

		// kaç adet btc alınacagının hesaplanması
		orderAmount = tradeAmountUSDT / signal.Price

	} else if signal.Action == "SELL" {
		// sell işlemi: basecoin yani btc bakiyesini kontrol etmeliyiz
		balance, err := s.exchange.CheckBalance(baseCoin)
		if err != nil {
			_ = s.repo.UpdateSignalStatus(signal.ID, "FAILED")
			return fmt.Errorf("btc bakiye kontrolü başarısız : %w", err)
		}
		// en azından satacak küsüratlı bir miktar btc'miz var mı ?
		if balance <= 0.0001 {
			log.Println("[SERVICE] elde satacak yeterli btc yok, işlem reddedildi.")
			_ = s.repo.UpdateSignalStatus(signal.ID, "REJECTED_NO_FUNDS")
			return errors.New("yetersiz coin bakiyesi")
		}

		// kural: elimizdeki tüm btc'yi satalım
		orderAmount = balance
	}

	// 4. Adım, ml botun ürettigi sinyali, kesin karara "Order" dönüştürme.
	order := &domain.Order{
		ID:        generateUUID(),
		SignalID:  signal.ID,
		Symbol:    signal.Symbol,
		Side:      signal.Action,
		OrderType: "market", //anında işlem atmak için market fiyatından emri veriyoruz
		Amount:    orderAmount,
		Price:     signal.Price,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	// 5. Adım, emri "order" db'ye kaydet
	if err := s.repo.SaveOrder(order); err != nil {
		return fmt.Errorf("emir kaydedilmedi: %w", err)
	}
	// 6. Adım, emri borsaya "ccxt" ile gönder, dış dünyayı tetikliyoruz.
	log.Printf("[SERVICE] ccxt'ye emir iletiliyor: %s %f adet %s\n", order.Side, orderAmount, order.Symbol)
	executedOrder, err := s.exchange.ExecuteOrder(order)
	if err != nil {
		// borsa hata verirse db'yi güncelle
		_ = s.repo.UpdateOrderStatus(executedOrder.ID, "FAILED")
		_ = s.repo.UpdateSignalStatus(signal.ID, "FAILED")
		return fmt.Errorf("borsada işlem başarısız oldu: %w", err)
	}
	// 7. Adım, herşey tamamsa statusleri güncelleriyoruz
	_ = s.repo.UpdateOrderStatus(executedOrder.ID, "COMPLETED")
	_ = s.repo.UpdateSignalStatus(signal.ID, "COMPLETED")

	log.Printf("[SERVICE] işlem başarıyla tamamlandı !, borsa fiş no : %s\n", executedOrder.ExchangeReqID)
	return nil
}

// generateuuıd
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (s *signalService) GetAllSignals() ([]domain.Signal, error) {
	return s.repo.GetAllSignals()
}
