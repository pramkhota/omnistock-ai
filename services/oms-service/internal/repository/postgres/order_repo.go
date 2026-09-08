package postgres

import (
	"database/sql"

	"omnistock-ai/services/oms-service/internal/domain"
)

type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new postgres repository for orders
func NewOrderRepository(db *sql.DB) domain.OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// CreateOrder saves the order and its items securely using a Database Transaction
func (r *orderRepository) CreateOrder(order *domain.Order) error {
	// 1. Begin the Transaction (Start a safe zone)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// 2. Insert the main order (Header)
	orderQuery := `
		INSERT INTO orders (order_ref, customer_name, shipping_address, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(orderQuery,
		order.OrderRef,
		order.CustomerName,
		order.ShippingAddress,
		order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	// If header fails, Rollback (Cancel everything)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Loop and Insert all order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_sku, quantity)
		VALUES ($1, $2, $3)
		RETURNING id`

	for i := range order.Items {
		order.Items[i].OrderID = order.ID // Link the item to the new order ID

		err = tx.QueryRow(itemQuery,
			order.Items[i].OrderID,
			order.Items[i].ProductSKU,
			order.Items[i].Quantity,
		).Scan(&order.Items[i].ID)

		// If ANY item fails, Rollback everything (Header is also deleted)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 4. Everything is successful, Commit to save permanently!
	return tx.Commit()
}

// GetOrderByRef and UpdateStatus are left empty for now
func (r *orderRepository) GetOrderByRef(orderRef string) (*domain.Order, error) {
	return nil, nil
}

func (r *orderRepository) UpdateStatus(orderID int, status string, trackingNo string) error {
	return nil
}
