package ports

import "v3-trading-bot/internal/core/domain"

// driving port içeriye giren akıs
// signalservice uygulamanın beyni, http'den sinyali alır ve işler.
type SignalService interface {
	ProcessSignal(signal *domain.Signal) error
	GetAllSignals() ([]domain.Signal, error)
}

// driven port dısarıya cıkan akıs
// signal repository veritabanı işlemlerini yapar
type SignalRepository interface {
	// sinyal işlemleri
	SaveSignal(signal *domain.Signal) error
	UpdateSignalStatus(id string, status string) error // sinyalin akibetidir, işlem durumunu güncelleriz, daha sonrasında hangi sinyalleri degerlendirip degerlendirmedigine bakarız

	// emir işlemleri, ürettigimiz emri ve durumunu DB'de sakalamak için
	SaveOrder(order *domain.Order) error
	UpdateOrderStatus(id string, status string) error // emrin akibetidir, order basladıgında işlem open olarak yazılır, tamamlanınca closed yazarız, hata olusursa failed yazarız.

	GetAllSignals() ([]domain.Signal, error)
}

// exchangeport ccxt (borsa/ paper trading) iletişim arayüzü

type ExchangePort interface {
	//hesaplanmıs order nesnesini borsaya iletir, işlemin basarılı olma durumunda,
	// exchangereqıd'si dolmus güncel order'ı geri döner
	ExecuteOrder(order *domain.Order) (*domain.Order, error)

	// işlem öncesi paper trading testnet cüzdanındaki güncel bakiyeyi sorgular, daha sonra executeorder' içerisindeki order hesaplanır.
	CheckBalance(asset string) (float64, error)
}
