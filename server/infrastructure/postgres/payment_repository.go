package postgres

import (
	"context"
	"errors"
	"fincode-token-practice/server/domain"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	record := Payment{
		ID:               p.ID,
		CardID:           p.CardID,
		FincodePaymentID: p.FincodePaymentID,
		FincodeAccessID:  p.FincodeAccessID,
		Amount:           p.Amount,
		Status:           p.Status,
		CreatedAt:        p.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("PaymentRepository.Save: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindAll(ctx context.Context) ([]*domain.Payment, error) {
	var records []Payment
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("PaymentRepository.FindAll: %w", err)
	}
	payments := make([]*domain.Payment, len(records))
	for i, record := range records {
		payments[i] = toPayment(record)
	}
	return payments, nil
}

func (r *PaymentRepository) FindByFincodePaymentID(ctx context.Context, fincodePaymentID string) (*domain.Payment, error) {
	var record Payment
	err := r.db.WithContext(ctx).Where("fincode_payment_id = ?", fincodePaymentID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PaymentRepository.FindByFincodePaymentID: %w", err)
	}
	return toPayment(record), nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	if err := r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("PaymentRepository.UpdateStatus: %w", err)
	}
	return nil
}

func toPayment(record Payment) *domain.Payment {
	return &domain.Payment{
		ID:               record.ID,
		CardID:           record.CardID,
		FincodePaymentID: record.FincodePaymentID,
		FincodeAccessID:  record.FincodeAccessID,
		Amount:           record.Amount,
		Status:           record.Status,
		CreatedAt:        record.CreatedAt,
	}
}
