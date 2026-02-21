package handlers

import (
	"v3-trading-bot/internal/core/domain"
	"v3-trading-bot/internal/core/ports"

	"github.com/gofiber/fiber/v2"
)

// signalhandler, http isteklerini karsılayan adaptörümüz - kapıcımızdır.
type SignalHandler struct {
	service ports.SignalService
}

// newsignalhandler, handler'ı ayağa kaldıran constructor fonksiyonu.
func NewSignalHandler(service ports.SignalService) *SignalHandler {
	return &SignalHandler{
		service: service,
	}
}

// handlepostsignal, python botundan gelen "POST /api/v1/signals" istegini isler.
func (h *SignalHandler) HandlePostSignal(c *fiber.Ctx) error {
	// 1. gelen json'ı dolduracagımız boş bir domain.signal nesnesi yaratıyoruz
	var signal domain.Signal

	// 2. gofiber'ın bodyparser yetenegi ile json'ı struct'a çeviriyoruz (unmarshal)
	if err := c.BodyParser(&signal); err != nil {
		// eğer python botu bozuk bir json gönderdiyse içeri almadan kapıdan kovuyoruz.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Geçersiz JSON formatı",
		})
	}

	// 3. basit dogrulama (validation) : boş veri gelmesin
	if signal.Symbol == "" || signal.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Symbol ve Action alanları zorunludur",
		})
	}

	// 4. işi beyne devretme aşaması
	// handler risk hesabı yapmaz, db'ye gitmez, sadece pas verir.
	if err := h.service.ProcessSignal(&signal); err != nil {
		// eger servis katmanında bakiye yetersizliği veya ccxt hatası olduysa 500 dönsün
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// 5. her şey başarılıysa python botuna 200 OK ve "aldım, işledim" mesajı dönüyoruz.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Sinyal başarıyla alındı ve borsaya iletidildi.",
	})
}

func (s *SignalHandler) GetTradeHistory(c *fiber.Ctx) error {
	signals, err := s.service.GetAllSignals()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Veriler çekilemedi"})
	}
	return c.JSON(signals)
}
