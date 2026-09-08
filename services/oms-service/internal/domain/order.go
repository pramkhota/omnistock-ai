package domain

import "time"

// Order represents a customer's purchase order
type Order struct {
	ID              int         `json:"id"`
	OrderRef        string      `json:"order_ref"`
	CustomerName    string      `json:"customer_name"`
	ShippingAddress string      `json:"shipping_address"`
	Status          string      `json:"status"`
	TrackingNumber  string      `json:"tracking_number,omitempty"`
	Items           []OrderItem `json:"items"` // One order can have multiple items
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// OrderItem represents a specific product and quantity within an order
type OrderItem struct {
	ID         int    `json:"id"`
	OrderID    int    `json:"order_id"`
	ProductSKU string `json:"product_sku"`
	Quantity   int    `json:"quantity"`
}

// OrderRepository defines the database interactions for orders
type OrderRepository interface {
	CreateOrder(order *Order) error
	GetOrderByRef(orderRef string) (*Order, error)
	UpdateStatus(orderID int, status string, trackingNo string) error
}

// OrderUsecase defines the business logic for managing orders
type OrderUsecase interface {
	PlaceOrder(order *Order) error
	GetOrderDetails(orderRef string) (*Order, error)
}
