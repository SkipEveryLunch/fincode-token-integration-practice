package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID                uuid.UUID
	FincodeCustomerID string
	CreatedAt         time.Time
}

type CustomerRepository interface {
	// Get はシングルトンのカスタマーを返す。存在しない場合は nil, nil。
	Get(ctx context.Context) (*Customer, error)
	Save(ctx context.Context, customer *Customer) error
}
