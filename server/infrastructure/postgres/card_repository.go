package postgres

import (
	"context"
	"errors"
	"fincode-token-practice/server/domain"
	"fmt"

	"gorm.io/gorm"
)

type CardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) FindActive(ctx context.Context) (*domain.Card, error) {
	var record Card
	err := r.db.WithContext(ctx).Where("is_alive = ?", true).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("CardRepository.FindActive: %w", err)
	}
	return toCard(record), nil
}

func (r *CardRepository) DeactivateAll(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Model(&Card{}).Where("is_alive = ?", true).Update("is_alive", false).Error; err != nil {
		return fmt.Errorf("CardRepository.DeactivateAll: %w", err)
	}
	return nil
}

func (r *CardRepository) Save(ctx context.Context, c *domain.Card) error {
	record := Card{
		ID:               c.ID,
		CustomerID:       c.CustomerID,
		FincodeCardID:    c.FincodeCardID,
		MaskedCardNumber: c.MaskedCardNumber,
		Expire:           c.Expire,
		Brand:            c.Brand,
		IsAlive:          c.IsAlive,
		CreatedAt:        c.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("CardRepository.Save: %w", err)
	}
	return nil
}

func toCard(record Card) *domain.Card {
	return &domain.Card{
		ID:               record.ID,
		CustomerID:       record.CustomerID,
		FincodeCardID:    record.FincodeCardID,
		MaskedCardNumber: record.MaskedCardNumber,
		Expire:           record.Expire,
		Brand:            record.Brand,
		IsAlive:          record.IsAlive,
		CreatedAt:        record.CreatedAt,
	}
}
