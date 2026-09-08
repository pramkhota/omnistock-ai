package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	pb "omnistock-ai/proto/inventory/v1"
	"omnistock-ai/services/oms-service/internal/domain"
)

// --- 1. Mock Order Repository ---
type mockOrderRepo struct{}

func (m *mockOrderRepo) CreateOrder(order *domain.Order) error {
	order.ID = 100 // Mock assigned ID
	return nil
}
func (m *mockOrderRepo) GetOrderByRef(orderRef string) (*domain.Order, error) {
	return nil, nil
}
func (m *mockOrderRepo) UpdateStatus(orderID int, status string, trackingNo string) error {
	return nil
}

// --- 2. Mock Inventory gRPC Client ---
type mockInventoryClient struct {
	networkError    bool // Simulate gRPC server down
	outOfStockError bool // Simulate inventory saying "no stock"
}

func (m *mockInventoryClient) ReserveStock(ctx context.Context, in *pb.ReserveStockRequest, opts ...grpc.CallOption) (*pb.ReserveStockResponse, error) {
	if m.networkError {
		return nil, errors.New("connection refused")
	}
	if m.outOfStockError {
		return &pb.ReserveStockResponse{
			Success: false,
			Message: "insufficient stock",
		}, nil
	}
	// Success case
	return &pb.ReserveStockResponse{
		Success: true,
		Message: "reserved",
	}, nil
}

// --- 3. Test Cases ---

func TestPlaceOrder_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{}
	mockInv := &mockInventoryClient{}
	uc := NewOrderUsecase(mockRepo, mockInv)

	order := &domain.Order{
		OrderRef:        "ORD-001",
		CustomerName:    "John Doe",
		ShippingAddress: "BKK",
		Items: []domain.OrderItem{
			{ProductSKU: "IP15", Quantity: 1},
		},
	}

	err := uc.PlaceOrder(order)
	assert.NoError(t, err)
	assert.Equal(t, "PENDING", order.Status) // Should set status correctly
}

func TestPlaceOrder_MissingRef(t *testing.T) {
	uc := NewOrderUsecase(&mockOrderRepo{}, &mockInventoryClient{})
	order := &domain.Order{CustomerName: "John Doe"} // No OrderRef

	err := uc.PlaceOrder(order)
	assert.Error(t, err)
	assert.Equal(t, "order reference is required", err.Error())
}

func TestPlaceOrder_EmptyBasket(t *testing.T) {
	uc := NewOrderUsecase(&mockOrderRepo{}, &mockInventoryClient{})
	order := &domain.Order{
		OrderRef: "ORD-002", CustomerName: "John Doe", ShippingAddress: "BKK",
		Items: []domain.OrderItem{}, // Empty
	}

	err := uc.PlaceOrder(order)
	assert.Error(t, err)
	assert.Equal(t, "order must contain at least one item", err.Error())
}

func TestPlaceOrder_InventoryNetworkError(t *testing.T) {
	mockInv := &mockInventoryClient{networkError: true} // Simulate Server Down
	uc := NewOrderUsecase(&mockOrderRepo{}, mockInv)

	order := &domain.Order{
		OrderRef: "ORD-003", CustomerName: "John Doe", ShippingAddress: "BKK",
		Items: []domain.OrderItem{{ProductSKU: "IP15", Quantity: 1}},
	}

	err := uc.PlaceOrder(order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to contact inventory service")
}

func TestPlaceOrder_OutOfStock(t *testing.T) {
	mockInv := &mockInventoryClient{outOfStockError: true} // Simulate Item Sold Out
	uc := NewOrderUsecase(&mockOrderRepo{}, mockInv)

	order := &domain.Order{
		OrderRef: "ORD-004", CustomerName: "John Doe", ShippingAddress: "BKK",
		Items: []domain.OrderItem{{ProductSKU: "IP15", Quantity: 999}},
	}

	err := uc.PlaceOrder(order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reserve stock")
}
