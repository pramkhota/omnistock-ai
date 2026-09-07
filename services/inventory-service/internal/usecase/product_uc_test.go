package usecase

import (
	"testing"

	"omnistock-ai/services/inventory-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------
// 1. Create Mock Database (Mock Repository)
// ---------------------------------------------------
type mockProductRepo struct{}

func (m *mockProductRepo) Create(product *domain.Product) error {
	product.ID = 1
	return nil
}
func (m *mockProductRepo) GetByID(id int) (*domain.Product, error) { return nil, nil }
func (m *mockProductRepo) ListAll() ([]*domain.Product, error)     { return nil, nil }
func (m *mockProductRepo) UpdateStock(id int, qty int) error       { return nil }

// ---------------------------------------------------
// 2. Test Cases
// ---------------------------------------------------
func TestCreateProduct_Success(t *testing.T) {

	mockRepo := &mockProductRepo{}
	uc := NewProductUsecase(mockRepo)

	newProduct := &domain.Product{
		SKU:    "TEST-01",
		Name:   "Test Product",
		Width:  10,
		Length: 10,
		Height: 10,
		Weight: 1,
	}

	err := uc.CreateProduct(newProduct)

	assert.NoError(t, err)
}

func TestCreateProduct_EmptySKU(t *testing.T) {
	mockRepo := &mockProductRepo{}
	uc := NewProductUsecase(mockRepo)

	invalidProduct := &domain.Product{
		SKU:    "",
		Name:   "Test Product",
		Width:  10,
		Length: 10,
		Height: 10,
		Weight: 1,
	}

	err := uc.CreateProduct(invalidProduct)

	assert.Error(t, err)
	assert.Equal(t, "SKU cannot be empty", err.Error())
}
