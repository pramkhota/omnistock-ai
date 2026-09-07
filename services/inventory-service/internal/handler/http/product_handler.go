package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omnistock-ai/services/inventory-service/internal/domain"
)

type ProductHandler struct {
	usecase domain.ProductUsecase
}

// NewProductHandler creates a new handler and registers the routes
func NewProductHandler(router *gin.Engine, us domain.ProductUsecase) {
	handler := &ProductHandler{
		usecase: us,
	}

	// Registering the route (Method POST)
	router.POST("/api/v1/products", handler.CreateProduct)
}

// CreateProduct handles the HTTP POST request for creating a new product
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product domain.Product

	// 1. Convert JSON from the request body into our Go struct
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// 2. Send the product to the Usecase layer for validation and saving
	if err := h.usecase.CreateProduct(&product); err != nil {
		// If business rules fail, return 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Return a success response with status 201 (Created)
	c.JSON(http.StatusCreated, gin.H{
		"message": "product created successfully",
		"data":    product,
	})
}
