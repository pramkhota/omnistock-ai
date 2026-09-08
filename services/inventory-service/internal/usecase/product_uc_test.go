package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"omnistock-ai/services/inventory-service/internal/domain"
)

// mockProductRepo is a manual mock for the ProductRepository
type mockProductRepo struct {
	shouldFailReserve bool // Used to simulate an out-of-stock scenario
}

// FIXED 1: Changed from CreateProduct to Create to match the interface
func (m *mockProductRepo) Create(product *domain.Product) error {
	product.ID = 1 // Mock auto-increment ID
	return nil
}

func (m *mockProductRepo) GetByID(id int) (*domain.Product, error) {
	// Dummy implementation
	return nil, nil
}

func (m *mockProductRepo) UpdateStock(id int, qty int) error {
	// Dummy implementation
	return nil
}

func (m *mockProductRepo) ListAll() ([]*domain.Product, error) {
	// Return empty list for now
	return nil, nil
}

func (m *mockProductRepo) ReserveStock(sku string, quantity int) error {
	if m.shouldFailReserve {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}

// --- Existing Tests (Create Product) ---

func TestCreateProduct_Success(t *testing.T) {
	mockRepo := &mockProductRepo{}
	uc := NewProductUsecase(mockRepo)

	product := &domain.Product{
		SKU:          "TEST-SKU",
		Name:         "Test Product",
		Width:        10,
		Length:       10,
		Height:       10,
		Weight:       1,
		QtyAvailable: 100,
	}

	err := uc.CreateProduct(product)

	assert.NoError(t, err)
	assert.Equal(t, 1, product.ID)
}

func TestCreateProduct_EmptySKU(t *testing.T) {
	mockRepo := &mockProductRepo{}
	uc := NewProductUsecase(mockRepo)

	product := &domain.Product{
		Name: "No SKU Product",
	}

	err := uc.CreateProduct(product)

	assert.Error(t, err)
	assert.Equal(t, "SKU cannot be empty", err.Error())
}

// --- New Tests (Reserve Stock) ---

func TestReserveStock_Success(t *testing.T) {
	mockRepo := &mockProductRepo{shouldFailReserve: false}
	uc := NewProductUsecase(mockRepo)

	err := uc.ReserveStock("IP15-PRO-256", 2)

	assert.NoError(t, err)
}

func TestReserveStock_InvalidInput(t *testing.T) {
	mockRepo := &mockProductRepo{}
	uc := NewProductUsecase(mockRepo)

	// Test missing SKU
	err := uc.ReserveStock("", 2)
	assert.Error(t, err)
	assert.Equal(t, "invalid reservation request", err.Error())

	// FIXED 2: Changed := to =
	err = uc.ReserveStock("IP15-PRO-256", 0)
	assert.Error(t, err)
	assert.Equal(t, "invalid reservation request", err.Error())
}

func TestReserveStock_FailInsufficientStock(t *testing.T) {
	mockRepo := &mockProductRepo{shouldFailReserve: true}
	uc := NewProductUsecase(mockRepo)

	err := uc.ReserveStock("IP15-PRO-256", 999) // Ordering more than available

	assert.Error(t, err)
	assert.Equal(t, "insufficient stock or product not found", err.Error())
}
