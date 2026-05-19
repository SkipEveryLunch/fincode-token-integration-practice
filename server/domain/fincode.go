package domain

import "context"

// FincodeCard はfincodeへのカード登録結果。
type FincodeCard struct {
	ID               string
	MaskedCardNumber string
	Expire           string
	Brand            string
}

// FincodePayment はfincodeへの決済作成結果。
type FincodePayment struct {
	ID       string
	AccessID string
}

// FincodeRepository はfincodeとのやり取りを抽象化する。
type FincodeRepository interface {
	CreateCustomer(ctx context.Context, customerID, email string) error
	RegisterCard(ctx context.Context, customerID, token string) (*FincodeCard, error)
	DeleteCard(ctx context.Context, customerID, cardID string) error
	// CreatePayment は決済を作成してIDとアクセスIDを返す。
	CreatePayment(ctx context.Context) (*FincodePayment, error)
	// ExecutePayment は決済を実行してリダイレクト先URLを返す。tds_type=2 で3DS認証が自動で走る。
	ExecutePayment(ctx context.Context, id, accessID, customerID, cardID, method, returnURL, returnURLOnFailure string) (string, error)
}
