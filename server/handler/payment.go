package handler

import (
	"fincode-token-practice/server/domain"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	customerRepo domain.CustomerRepository
	cardRepo     domain.CardRepository
	paymentRepo  domain.PaymentRepository
	fincodeRepo  domain.FincodeRepository
	baseURL      string
}

func NewPaymentHandler(
	customerRepo domain.CustomerRepository,
	cardRepo domain.CardRepository,
	paymentRepo domain.PaymentRepository,
	fincodeRepo domain.FincodeRepository,
	baseURL string,
) *PaymentHandler {
	return &PaymentHandler{
		customerRepo: customerRepo,
		cardRepo:     cardRepo,
		paymentRepo:  paymentRepo,
		fincodeRepo:  fincodeRepo,
		baseURL:      baseURL,
	}
}

func (h *PaymentHandler) List(c *gin.Context) {
	payments, err := h.paymentRepo.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payments"})
		return
	}
	type item struct {
		Amount    int                  `json:"amount"`
		Status    domain.PaymentStatus `json:"status"`
		CreatedAt time.Time            `json:"created_at"`
	}
	res := make([]item, len(payments))
	for i, p := range payments {
		res[i] = item{Amount: p.Amount, Status: p.Status, CreatedAt: p.CreatedAt}
	}
	c.JSON(http.StatusOK, res)
}

type purchaseRequest struct {
	Method string `json:"method" binding:"required,oneof=1 2"`
}

func (h *PaymentHandler) Purchase(c *gin.Context) {
	var req purchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	card, err := h.cardRepo.FindActive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active card"})
		return
	}
	if card == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active card"})
		return
	}

	customer, err := h.customerRepo.Get(ctx)
	if err != nil || customer == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get customer"})
		return
	}

	fincodePayment, err := h.fincodeRepo.CreatePayment(ctx)
	if err != nil {
		log.Printf("CreatePayment error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}

	redirectURL, err := h.fincodeRepo.ExecutePayment(ctx,
		fincodePayment.ID,
		fincodePayment.AccessID,
		customer.FincodeCustomerID,
		card.FincodeCardID,
		req.Method,
		h.baseURL+"/api/payments/callback",
		h.baseURL+"/api/payments/failure",
	)
	if err != nil {
		log.Printf("ExecutePayment error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to execute payment"})
		return
	}

	payment := &domain.Payment{
		ID:               uuid.New(),
		CardID:           card.ID,
		FincodePaymentID: fincodePayment.ID,
		FincodeAccessID:  fincodePayment.AccessID,
		Amount:           500,
		Status:           domain.PaymentStatusUnprocessed,
		CreatedAt:        time.Now(),
	}
	if err := h.paymentRepo.Save(ctx, payment); err != nil {
		log.Printf("PaymentRepository.Save error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect_url": redirectURL})
}
