package domain

import "context"

// FincodeCard はfincodeへのカード登録結果。
type FincodeCard struct {
	ID               string
	MaskedCardNumber string
	Expire           string
	Brand            string
}

// FincodeRepository はfincodeとのやり取りを抽象化する。
type FincodeRepository interface {
	CreateCustomer(ctx context.Context, customerID, email string) error
	RegisterCard(ctx context.Context, customerID, token string) (*FincodeCard, error)
}
