package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusUnprocessed PaymentStatus = "UNPROCESSED"
	PaymentStatusCaptured    PaymentStatus = "CAPTURED"
	PaymentStatusFailed      PaymentStatus = "FAILED"
)

type Payment struct {
	ID               uuid.UUID
	CardID           uuid.UUID
	FincodePaymentID string
	FincodeAccessID  string
	Amount           int
	Status           PaymentStatus
	CreatedAt        time.Time
}

type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) error
	FindAll(ctx context.Context) ([]*Payment, error)
	FindByFincodePaymentID(ctx context.Context, fincodePaymentID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error
}
