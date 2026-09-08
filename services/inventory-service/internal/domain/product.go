package domain

import (
	"time"
)

// Product represents an item in our warehouse inventory.
type Product struct {
	ID           int       `json:"id"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	Width        float64   `json:"width"`         // in cm (required for AI packaging calculation)
	Length       float64   `json:"length"`        // in cm
	Height       float64   `json:"height"`        // in cm
	Weight       float64   `json:"weight"`        // in kg
	QtyAvailable int       `json:"qty_available"` // Stock available for sale
	QtyReserved  int       `json:"qty_reserved"`  // Stock reserved in active orders (not yet packed)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProductRepository defines the contract for database operations.
type ProductRepository interface {
	Create(product *Product) error
	GetByID(id int) (*Product, error)
	ListAll() ([]*Product, error)
	UpdateStock(id int, qty int) error
	ReserveStock(sku string, quantity int) error
}

// ProductUsecase defines the contract for business logic operations.
type ProductUsecase interface {
	CreateProduct(product *Product) error
	GetProduct(id int) (*Product, error)
	GetAllProducts() ([]*Product, error)
	ReserveStock(sku string, quantity int) error
}
