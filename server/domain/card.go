package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Card struct {
	ID               uuid.UUID
	CustomerID       uuid.UUID
	FincodeCardID    string
	MaskedCardNumber string
	Expire           string
	Brand            string
	IsAlive          bool
	CreatedAt        time.Time
}

type CardRepository interface {
	// FindActive は is_alive=true のカードを1件返す。存在しない場合は nil, nil。
	FindActive(ctx context.Context) (*Card, error)
	// DeactivateAll は全カードを is_alive=false にする。
	DeactivateAll(ctx context.Context) error
	Save(ctx context.Context, card *Card) error
}
