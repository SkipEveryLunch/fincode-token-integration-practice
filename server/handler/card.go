package handler

import (
	"fincode-token-practice/server/domain"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardHandler struct {
	customerRepo domain.CustomerRepository
	cardRepo     domain.CardRepository
	fincodeRepo  domain.FincodeRepository
}

func NewCardHandler(
	customerRepo domain.CustomerRepository,
	cardRepo domain.CardRepository,
	fincodeRepo domain.FincodeRepository,
) *CardHandler {
	return &CardHandler{
		customerRepo: customerRepo,
		cardRepo:     cardRepo,
		fincodeRepo:  fincodeRepo,
	}
}

type registerCardRequest struct {
	Token string `json:"token" binding:"required"`
}

type registerCardResponse struct {
	MaskedCardNumber string `json:"masked_card_number"`
	Expire           string `json:"expire"`
	Brand            string `json:"brand"`
}

func (h *CardHandler) GetActive(c *gin.Context) {
	ctx := c.Request.Context()
	card, err := h.cardRepo.FindActive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active card"})
		return
	}
	if card == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active card"})
		return
	}
	c.JSON(http.StatusOK, registerCardResponse{
		MaskedCardNumber: card.MaskedCardNumber,
		Expire:           card.Expire,
		Brand:            card.Brand,
	})
}

func (h *CardHandler) Register(c *gin.Context) {
	var req registerCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 1. カスタマーシングルトン upsert（既存なら取得）
	customer, err := h.customerRepo.Get(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get customer"})
		return
	}
	if customer == nil {
		fincodeCustomerID := uuid.New().String()
		if err := h.fincodeRepo.CreateCustomer(ctx, fincodeCustomerID, "test@example.com"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create fincode customer"})
			return
		}
		customer = &domain.Customer{
			ID:                uuid.New(),
			FincodeCustomerID: fincodeCustomerID,
			CreatedAt:         time.Now(),
		}
		if err := h.customerRepo.Save(ctx, customer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save customer"})
			return
		}
	}

	// 2. 既存アクティブカードを取得（後で削除するため）
	oldCard, err := h.cardRepo.FindActive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active card"})
		return
	}

	// 3. fincodeにカード登録
	fincodeCard, err := h.fincodeRepo.RegisterCard(ctx, customer.FincodeCustomerID, req.Token)
	if err != nil {
		log.Printf("RegisterCard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register card in fincode"})
		return
	}

	// 4. 旧カードをfincodeから削除（新カード登録成功後）
	if oldCard != nil {
		if err := h.fincodeRepo.DeleteCard(ctx, customer.FincodeCustomerID, oldCard.FincodeCardID); err != nil {
			log.Printf("DeleteCard error (non-fatal): %v", err)
		}
	}

	// 5. 既存カードを無効化して新規カード保存
	if err := h.cardRepo.DeactivateAll(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate cards"})
		return
	}
	card := &domain.Card{
		ID:               uuid.New(),
		CustomerID:       customer.ID,
		FincodeCardID:    fincodeCard.ID,
		MaskedCardNumber: fincodeCard.MaskedCardNumber,
		Expire:           fincodeCard.Expire,
		Brand:            fincodeCard.Brand,
		IsAlive:          true,
		CreatedAt:        time.Now(),
	}
	if err := h.cardRepo.Save(ctx, card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save card"})
		return
	}

	c.JSON(http.StatusOK, registerCardResponse{
		MaskedCardNumber: fincodeCard.MaskedCardNumber,
		Expire:           fincodeCard.Expire,
		Brand:            fincodeCard.Brand,
	})
}
