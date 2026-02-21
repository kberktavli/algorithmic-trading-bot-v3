package repositories

import (
	"fmt"
	"v3-trading-bot/internal/core/domain"
	"v3-trading-bot/internal/core/ports"

	"gorm.io/gorm"
)

// postgresRepository, dış dünyadaki postgresql veritabanımızla iletişim kuran adaptörümüzdür.
type postgresRepository struct {
	db *gorm.DB
}

// newpostgresrepository, gorm veritabanı baglantısını (db) alıp, service'in bekledigi signalrepository sözlesmesini döner.
func NewPostgresRepository(db *gorm.DB) ports.SignalRepository {
	return &postgresRepository{
		db: db,
	}
}

// ------ sinyal işlemleri ------
func (r *postgresRepository) SaveSignal(signal *domain.Signal) error {
	// gorm, struct'ı alır ve otomatik olarak "signals" tablosuna insert into yapar
	if err := r.db.Create(signal).Error; err != nil {
		return fmt.Errorf("sinyal db'ye kaydedilmedi: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateSignalStatus(id string, status string) error {
	// update signals set status = ? where id = ?
	if err := r.db.Model(&domain.Signal{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("sinyal durumu güncellenemedi: %w", err)
	}
	return nil
}

// ----- emir işlemleri -----
func (r *postgresRepository) SaveOrder(order *domain.Order) error {
	// gorm otomatik olarak "orders" tablosuna kaydeder.
	if err := r.db.Create(order).Error; err != nil {
		return fmt.Errorf("emir db'ye kaydedilmedi: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateOrderStatus(id string, status string) error {
	// update orders set status = ? where id = ?
	if err := r.db.Model(&domain.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("order durumu güncellenmedi: %w", err)
	}
	return nil
}

// GetAllSignals
func (r *postgresRepository) GetAllSignals() ([]domain.Signal, error) {
	var signals []domain.Signal
	err := r.db.Order("created_at desc").Limit(100).Find(&signals).Error // sinyalleri yeniden eskiye doğru sıralayarak getirir.
	return signals, err
}
