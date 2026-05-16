package fincode

import (
	"bytes"
	"context"
	"encoding/json"
	"fincode-token-practice/server/domain"
	"fmt"
	"net/http"
)

const baseURL = "https://api.test.fincode.jp"

type Repository struct {
	secretKey  string
	httpClient *http.Client
}

func NewRepository(secretKey string) *Repository {
	return &Repository{
		secretKey:  secretKey,
		httpClient: &http.Client{},
	}
}

func (r *Repository) request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.secretKey)
	req.Header.Set("Content-Type", "application/json")
	return r.httpClient.Do(req)
}

func (r *Repository) CreateCustomer(ctx context.Context, customerID, email string) error {
	resp, err := r.request(ctx, http.MethodPost, "/v1/customers", map[string]string{
		"id":    customerID,
		"email": email,
	})
	if err != nil {
		return fmt.Errorf("FincodeRepository.CreateCustomer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FincodeRepository.CreateCustomer: status %d", resp.StatusCode)
	}
	return nil
}

type registerCardResp struct {
	ID     string `json:"id"`
	CardNo string `json:"card_no"`
	Expire string `json:"expire"`
	Brand  string `json:"brand"`
}

func (r *Repository) RegisterCard(ctx context.Context, customerID, token string) (*domain.FincodeCard, error) {
	resp, err := r.request(ctx, http.MethodPost, "/v1/customers/"+customerID+"/cards", map[string]string{
		"token":        token,
		"default_flag": "1",
	})
	if err != nil {
		return nil, fmt.Errorf("FincodeRepository.RegisterCard: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FincodeRepository.RegisterCard: status %d", resp.StatusCode)
	}
	var raw registerCardResp
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("FincodeRepository.RegisterCard: %w", err)
	}
	return &domain.FincodeCard{
		ID:               raw.ID,
		MaskedCardNumber: raw.CardNo,
		Expire:           raw.Expire,
		Brand:            raw.Brand,
	}, nil
}
