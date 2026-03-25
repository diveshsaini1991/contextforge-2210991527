package api

import (
	"net/http"
	"strconv"

	"github.com/example/sample-project/pkg/calculator"
	"github.com/gin-gonic/gin"
)

// CalculatorRequest represents a calculation request
type CalculatorRequest struct {
	A int `json:"a" binding:"required"`
	B int `json:"b" binding:"required"`
}

// CalculatorResponse represents a calculation response
type CalculatorResponse struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

// AddHandler handles addition requests
func AddHandler(c *gin.Context) {
	var req CalculatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := calculator.Add(req.A, req.B)
	c.JSON(http.StatusOK, CalculatorResponse{Result: float64(result)})
}

// SubtractHandler handles subtraction requests
func SubtractHandler(c *gin.Context) {
	var req CalculatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := calculator.Subtract(req.A, req.B)
	c.JSON(http.StatusOK, CalculatorResponse{Result: float64(result)})
}

// DivideHandler handles division requests
func DivideHandler(c *gin.Context) {
	var req CalculatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := calculator.Divide(req.A, req.B)
	if err != nil {
		c.JSON(http.StatusBadRequest, CalculatorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, CalculatorResponse{Result: result})
}

// FactorialHandler handles factorial requests
func FactorialHandler(c *gin.Context) {
	nStr := c.Param("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number"})
		return
	}

	result, err := calculator.Factorial(n)
	if err != nil {
		c.JSON(http.StatusBadRequest, CalculatorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, CalculatorResponse{Result: float64(result)})
}

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/add", AddHandler)
		v1.POST("/subtract", SubtractHandler)
		v1.POST("/divide", DivideHandler)
		v1.GET("/factorial/:n", FactorialHandler)
	}
}
