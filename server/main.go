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

	customerRepo := infrapostgres.NewCustomerRepository(db.DB)
	cardRepo := infrapostgres.NewCardRepository(db.DB)
	paymentRepo := infrapostgres.NewPaymentRepository(db.DB)
	fincodeRepo := infrafincode.NewRepository(secretKey)

	cardHandler := handler.NewCardHandler(customerRepo, cardRepo, fincodeRepo)
	paymentHandler := handler.NewPaymentHandler(customerRepo, cardRepo, paymentRepo, fincodeRepo, baseURL)

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
		api.POST("/payments", paymentHandler.Purchase)
	}

	r.Run(":8080")
}
