package main

import (
	"fmt"
	"log"

	"github.com/example/sample-project/pkg/api"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Setup API routes
	api.SetupRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "sample-calculator-api",
		})
	})

	fmt.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
