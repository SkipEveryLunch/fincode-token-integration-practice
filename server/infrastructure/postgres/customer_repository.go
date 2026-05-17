package postgres

import (
	"context"
	"errors"
	"fincode-token-practice/server/domain"
	"fmt"

	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Get(ctx context.Context) (*domain.Customer, error) {
	var record Customer
	err := r.db.WithContext(ctx).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("CustomerRepository.Get: %w", err)
	}
	return toCustomer(record), nil
}

func (r *CustomerRepository) Save(ctx context.Context, c *domain.Customer) error {
	record := Customer{
		ID:                c.ID,
		FincodeCustomerID: c.FincodeCustomerID,
		CreatedAt:         c.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("CustomerRepository.Save: %w", err)
	}
	return nil
}

func toCustomer(record Customer) *domain.Customer {
	return &domain.Customer{
		ID:                record.ID,
		FincodeCustomerID: record.FincodeCustomerID,
		CreatedAt:         record.CreatedAt,
	}
}
