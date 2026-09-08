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
	router.GET("/api/v1/products", handler.GetAllProducts)
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

// GetAllProducts handles the HTTP GET request to list all products
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products, err := h.usecase.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    products,
	})
}
