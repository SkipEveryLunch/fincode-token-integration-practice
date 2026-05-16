package main

import (
	"fincode-token-practice/server/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	r.Run(":8080")
}
