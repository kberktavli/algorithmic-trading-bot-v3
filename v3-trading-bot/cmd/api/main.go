package main

import (
	"fmt"
	"log"
	"os"

	"v3-trading-bot/internal/adapters/execution"
	"v3-trading-bot/internal/adapters/handlers"
	"v3-trading-bot/internal/adapters/repositories"
	"v3-trading-bot/internal/core/domain" // <--- BURASI EKLENDİ (Struct'ları tanımak için)
	"v3-trading-bot/internal/core/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. .env dosyasını yükle
	if err := godotenv.Load(); err != nil {
		log.Println("Uyarı: .env dosyası bulunamadı, sistem ortam değişkenleri kullanılacak.")
	}

	// 2. Veritabanı Bağlantısı (PostgreSQL)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Istanbul",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	log.Println("✅ PostgreSQL bağlantısı başarılı!")

	// --- EKLENEN KISIM: OTOMATİK MIGRATION ---
	log.Println("⏳ Tablolar veritabanında oluşturuluyor...")
	// GORM, Domain içindeki Signal ve Order struct'larına bakıp
	// senin yerine "CREATE TABLE signals..." ve "CREATE TABLE orders..." komutlarını çalıştırır.
	err = db.AutoMigrate(&domain.Signal{}, &domain.Order{})
	if err != nil {
		log.Fatalf("Tablo oluşturma hatası: %v", err)
	}
	log.Println("✅ Tablolar (signals, orders) başarıyla oluşturuldu!")
	// ------------------------------------------

	// 3. CCXT (Borsa) Adaptörü Başlatma
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")
	isTestnet := os.Getenv("USE_TESTNET") == "true"

	exchangeAdapter, err := execution.NewCCXTAdapter(apiKey, secretKey, isTestnet)
	if err != nil {
		log.Fatalf("CCXT Başlatılamadı: %v", err)
	}
	log.Println("✅ CCXT Borsa adaptörü hazır!")

	// 4. Bağımlılıkları Bağlama
	repo := repositories.NewPostgresRepository(db)
	service := services.NewSignalService(repo, exchangeAdapter)
	handler := handlers.NewSignalHandler(service)

	// 5. GoFiber Sunucusunu Başlatma
	app := fiber.New()
	app.Use(logger.New())
	app.Static("/", "./web")
	api := app.Group("/api/v1")
	api.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	api.Post("/signals", handler.HandlePostSignal)
	api.Get("/history", handler.GetTradeHistory)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	for _, r := range app.GetRoutes() {
		log.Println(r.Method, r.Path)
	}

	log.Printf("🚀 Trading Bot V3 %s portunda çalışıyor...", port)
	log.Fatal(app.Listen(":" + port))
}
