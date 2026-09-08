package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "omnistock-ai/proto/inventory/v1"
	"omnistock-ai/services/oms-service/internal/domain"
)

type orderUsecase struct {
	repo            domain.OrderRepository
	inventoryClient pb.InventoryServiceClient
}

// NewOrderUsecase is the constructor function for the order usecase
func NewOrderUsecase(repo domain.OrderRepository, invClient pb.InventoryServiceClient) domain.OrderUsecase {
	return &orderUsecase{
		repo:            repo,
		inventoryClient: invClient,
	}
}

// PlaceOrder validates the incoming order data before saving it
func (u *orderUsecase) PlaceOrder(order *domain.Order) error {
	// 1. Basic Validations
	if order.OrderRef == "" {
		return errors.New("order reference is required")
	}
	if order.CustomerName == "" || order.ShippingAddress == "" {
		return errors.New("customer name and shipping address are required")
	}
	if len(order.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	// 2. Loop through items to validate and RESERVE STOCK via gRPC
	for _, item := range order.Items {
		if item.ProductSKU == "" {
			return errors.New("product SKU is required for all items")
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than zero")
		}

		// --- THE MAGIC HAPPENS HERE: Call Inventory Service via gRPC ---
		// We set a 5-second timeout. If Inventory doesn't respond, we fail the order.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := u.inventoryClient.ReserveStock(ctx, &pb.ReserveStockRequest{
			ProductSku: item.ProductSKU,
			Quantity:   int32(item.Quantity),
		})

		// If the gRPC network call fails
		if err != nil {
			return fmt.Errorf("failed to contact inventory service for SKU %s: %v", item.ProductSKU, err)
		}

		// If the Inventory Service replies with "Success: false" (e.g. out of stock)
		if !resp.Success {
			return fmt.Errorf("failed to reserve stock for SKU %s: %s", item.ProductSKU, resp.Message)
		}
	}

	// 3. Set default status and save to DB
	order.Status = "PENDING"
	return u.repo.CreateOrder(order)
}

// GetOrderDetails fetches an order and its items by order reference
func (u *orderUsecase) GetOrderDetails(orderRef string) (*domain.Order, error) {
	return u.repo.GetOrderByRef(orderRef)
}
