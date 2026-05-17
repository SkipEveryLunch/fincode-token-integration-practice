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
	m := Payment{
		ID:               p.ID,
		CardID:           p.CardID,
		FincodePaymentID: p.FincodePaymentID,
		FincodeAccessID:  p.FincodeAccessID,
		Amount:           p.Amount,
		Status:           p.Status,
		CreatedAt:        p.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("PaymentRepository.Save: %w", err)
	}
	return nil
}

func (r *PaymentRepository) FindAll(ctx context.Context) ([]*domain.Payment, error) {
	var ms []Payment
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("PaymentRepository.FindAll: %w", err)
	}
	payments := make([]*domain.Payment, len(ms))
	for i, m := range ms {
		payments[i] = toPayment(m)
	}
	return payments, nil
}

func (r *PaymentRepository) FindByFincodePaymentID(ctx context.Context, fincodePaymentID string) (*domain.Payment, error) {
	var m Payment
	err := r.db.WithContext(ctx).Where("fincode_payment_id = ?", fincodePaymentID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("PaymentRepository.FindByFincodePaymentID: %w", err)
	}
	return toPayment(m), nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	if err := r.db.WithContext(ctx).Model(&Payment{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("PaymentRepository.UpdateStatus: %w", err)
	}
	return nil
}

func toPayment(m Payment) *domain.Payment {
	return &domain.Payment{
		ID:               m.ID,
		CardID:           m.CardID,
		FincodePaymentID: m.FincodePaymentID,
		FincodeAccessID:  m.FincodeAccessID,
		Amount:           m.Amount,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
	}
}
