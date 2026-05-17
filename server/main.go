package main

import (
	"fincode-token-practice/server/db"
	"fincode-token-practice/server/handler"
	infrafincode "fincode-token-practice/server/infrastructure/fincode"
	infrapostgres "fincode-token-practice/server/infrastructure/postgres"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	secretKey := os.Getenv("FINCODE_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("FINCODE_SECRET_KEY is not set")
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL is not set")
	}
	frontBaseURL := os.Getenv("FRONT_BASE_URL")
	if frontBaseURL == "" {
		log.Fatal("FRONT_BASE_URL is not set")
	}
	webhookSecret := os.Getenv("FINCODE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Fatal("FINCODE_WEBHOOK_SECRET is not set")
	}

	customerRepo := infrapostgres.NewCustomerRepository(db.DB)
	cardRepo := infrapostgres.NewCardRepository(db.DB)
	paymentRepo := infrapostgres.NewPaymentRepository(db.DB)
	fincodeRepo := infrafincode.NewRepository(secretKey)

	cardHandler := handler.NewCardHandler(customerRepo, cardRepo, fincodeRepo)
	paymentHandler := handler.NewPaymentHandler(customerRepo, cardRepo, paymentRepo, fincodeRepo, baseURL, frontBaseURL, webhookSecret)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	})

	api := r.Group("/api")
	{
		api.GET("/cards/active", cardHandler.GetActive)
		api.POST("/cards", cardHandler.Register)
		api.GET("/payments", paymentHandler.List)
		api.POST("/payments", paymentHandler.Purchase)
		api.POST("/payments/callback", paymentHandler.Callback)
		api.GET("/payments/callback", paymentHandler.Callback)
		api.POST("/payments/failure", paymentHandler.Failure)
		api.GET("/payments/failure", paymentHandler.Failure)
		api.POST("/payments/webhook", paymentHandler.Webhook)
	}

	r.Run(":8080")
}
