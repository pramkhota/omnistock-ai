package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omnistock-ai/services/oms-service/internal/domain"
)

type OrderHandler struct {
	usecase domain.OrderUsecase
}

// NewOrderHandler initializes the handler and routes
func NewOrderHandler(router *gin.Engine, us domain.OrderUsecase) {
	handler := &OrderHandler{
		usecase: us,
	}

	// Register the POST route for placing an order
	router.POST("/api/v1/orders", handler.PlaceOrder)
}

// PlaceOrder handles incoming HTTP requests to create a new order
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var order domain.Order

	// 1. Bind JSON body to struct
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	// 2. Call Usecase
	if err := h.usecase.PlaceOrder(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Return Success
	c.JSON(http.StatusCreated, gin.H{
		"message": "order placed successfully",
		"data":    order,
	})
}
