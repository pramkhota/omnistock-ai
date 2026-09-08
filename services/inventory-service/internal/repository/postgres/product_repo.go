package postgres

import (
	"database/sql"
	"errors"

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
	query := `SELECT id, sku, name, width, length, height, weight, qty_available, qty_reserved, created_at, updated_at FROM products`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product

	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID, &p.SKU, &p.Name,
			&p.Width, &p.Length, &p.Height, &p.Weight,
			&p.QtyAvailable, &p.QtyReserved,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	return products, nil
}

func (r *productRepository) UpdateStock(id int, qty int) error {
	return nil
}

// ReserveStock atomically deducts available quantity and increases reserved quantity
func (r *productRepository) ReserveStock(sku string, quantity int) error {
	// Execute SQL to deduct available stock and add to reserved stock
	// The WHERE clause ensures we only update if there is enough stock available
	query := `
		UPDATE products 
		SET qty_available = qty_available - $1, 
		    qty_reserved = qty_reserved + $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE sku = $2 AND qty_available >= $1
	`

	result, err := r.db.Exec(query, quantity, sku)
	if err != nil {
		return err
	}

	// Check if any row was actually updated (if 0, it means out of stock or SKU not found)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("insufficient stock or product not found")
	}

	return nil
}
