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
	var m Customer
	err := r.db.WithContext(ctx).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("CustomerRepository.Get: %w", err)
	}
	return toCustomer(m), nil
}

func (r *CustomerRepository) Save(ctx context.Context, c *domain.Customer) error {
	m := Customer{
		ID:                c.ID,
		FincodeCustomerID: c.FincodeCustomerID,
		CreatedAt:         c.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("CustomerRepository.Save: %w", err)
	}
	return nil
}

func toCustomer(m Customer) *domain.Customer {
	return &domain.Customer{
		ID:                m.ID,
		FincodeCustomerID: m.FincodeCustomerID,
		CreatedAt:         m.CreatedAt,
	}
}
