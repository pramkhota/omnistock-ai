package usecase

import (
	"errors"
	"omnistock-ai/services/inventory-service/internal/domain"
)

// productUsecase is the concrete implementation of domain.ProductUsecase
type productUsecase struct {
	repo domain.ProductRepository
}

// NewProductUsecase is the constructor function for the product usecase.
func NewProductUsecase(repo domain.ProductRepository) domain.ProductUsecase {
	return &productUsecase{
		repo: repo,
	}
}

// CreateProduct validates the business rules before passing data to the repository layer.
func (u *productUsecase) CreateProduct(product *domain.Product) error {
	// 1. Business Validation Rules
	if product.SKU == "" {
		return errors.New("SKU cannot be empty")
	}
	if product.Name == "" {
		return errors.New("product name cannot be empty")
	}
	if product.Width <= 0 || product.Length <= 0 || product.Height <= 0 || product.Weight <= 0 {
		return errors.New("dimensions and weight must be strictly positive")
	}

	// 2. Pass to the repository to interact with the database
	return u.repo.Create(product)
}

// GetProduct retrieves a product by its ID. (To be implemented)
func (u *productUsecase) GetProduct(id int) (*domain.Product, error) {
	return nil, nil
}

// GetAllProducts retrieves all products. (To be implemented)
func (u *productUsecase) GetAllProducts() ([]*domain.Product, error) {
	return nil, nil
}
