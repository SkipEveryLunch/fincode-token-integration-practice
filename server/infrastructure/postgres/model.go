package postgres

import (
	"fincode-token-practice/server/domain"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FincodeCustomerID string    `gorm:"not null;uniqueIndex"`
	CreatedAt         time.Time
}

type Card struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID       uuid.UUID `gorm:"type:uuid;not null"`
	FincodeCardID    string    `gorm:"not null"`
	MaskedCardNumber string
	Expire           string
	Brand            string
	IsAlive          bool      `gorm:"not null;default:true"`
	CreatedAt        time.Time
}

type Payment struct {
	ID               uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CardID           uuid.UUID            `gorm:"type:uuid;not null"`
	FincodePaymentID string               `gorm:"not null"`
	FincodeAccessID  string               `gorm:"not null"`
	Amount           int                  `gorm:"not null;default:500"`
	Status           domain.PaymentStatus `gorm:"not null;default:UNPROCESSED"`
	CreatedAt        time.Time
}
