package postgres

import (
	"database/sql"

	"omnistock-ai/services/inventory-service/internal/domain"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) Create(product *domain.Product) error {
	query := `
		INSERT INTO products (sku, name, width, length, height, weight)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		product.SKU,
		product.Name,
		product.Width,
		product.Length,
		product.Height,
		product.Weight,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	return err
}

func (r *productRepository) GetByID(id int) (*domain.Product, error) {
	return nil, nil
}

func (r *productRepository) ListAll() ([]*domain.Product, error) {
	return nil, nil
}

func (r *productRepository) UpdateStock(id int, qty int) error {
	return nil
}
