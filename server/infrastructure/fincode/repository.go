package fincode

import (
	"bytes"
	"context"
	"encoding/json"
	"fincode-token-practice/server/domain"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
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

func (r *Repository) request(ctx context.Context, method, path string, body any, extraHeaders map[string]string) (*http.Response, error) {
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
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	return r.httpClient.Do(req)
}

func readErrorBody(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

func (r *Repository) CreateCustomer(ctx context.Context, customerID, email string) error {
	resp, err := r.request(ctx, http.MethodPost, "/v1/customers", map[string]string{
		"id":    customerID,
		"email": email,
	}, nil)
	if err != nil {
		return fmt.Errorf("FincodeRepository.CreateCustomer: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FincodeRepository.CreateCustomer: status %d body %s", resp.StatusCode, readErrorBody(resp))
	}
	return nil
}

func (r *Repository) DeleteCard(ctx context.Context, customerID, cardID string) error {
	resp, err := r.request(ctx, http.MethodDelete, "/v1/customers/"+customerID+"/cards/"+cardID, nil, nil)
	if err != nil {
		return fmt.Errorf("FincodeRepository.DeleteCard: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FincodeRepository.DeleteCard: status %d body %s", resp.StatusCode, readErrorBody(resp))
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
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("FincodeRepository.RegisterCard: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FincodeRepository.RegisterCard: status %d body %s", resp.StatusCode, readErrorBody(resp))
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

type createPaymentResp struct {
	ID       string `json:"id"`
	AccessID string `json:"access_id"`
}

func (r *Repository) CreatePayment(ctx context.Context) (*domain.FincodePayment, error) {
	resp, err := r.request(ctx, http.MethodPost, "/v1/payments", map[string]any{
		"pay_type": "Card",
		"job_code": "CAPTURE",
		"amount":   "500",
		"tds_type": "2",
	}, map[string]string{
		"Idempotency-Key": uuid.New().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("FincodeRepository.CreatePayment: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FincodeRepository.CreatePayment: status %d body %s", resp.StatusCode, readErrorBody(resp))
	}
	var raw createPaymentResp
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("FincodeRepository.CreatePayment: %w", err)
	}
	return &domain.FincodePayment{
		ID:       raw.ID,
		AccessID: raw.AccessID,
	}, nil
}

type executePaymentResp struct {
	RedirectURL string `json:"redirect_url"`
}

func (r *Repository) ExecutePayment(ctx context.Context, id, accessID, customerID, cardID, method, returnURL, returnURLOnFailure string) (string, error) {
	resp, err := r.request(ctx, http.MethodPut, "/v1/payments/"+id, map[string]any{
		"pay_type":              "Card",
		"access_id":             accessID,
		"customer_id":           customerID,
		"card_id":               cardID,
		"method":                method,
		"return_url":            returnURL,
		"return_url_on_failure": returnURLOnFailure,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("FincodeRepository.ExecutePayment: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("FincodeRepository.ExecutePayment: status %d body %s", resp.StatusCode, readErrorBody(resp))
	}
	var raw executePaymentResp
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", fmt.Errorf("FincodeRepository.ExecutePayment: %w", err)
	}
	return raw.RedirectURL, nil
}
