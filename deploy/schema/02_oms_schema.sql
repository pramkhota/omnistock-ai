-- Table: orders (Header for each customer order)
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_ref VARCHAR(100) UNIQUE NOT NULL, -- External reference e.g., from Shopee/TikTok
    customer_name VARCHAR(255) NOT NULL,
    shipping_address TEXT NOT NULL,
    
    -- Status pipeline: PENDING -> PICKING -> SHIPPED
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    tracking_number VARCHAR(100), -- Assigned after packing is complete
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: order_items (Items contained within a specific order)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_sku VARCHAR(50) NOT NULL, -- Using SKU instead of ID for loosely coupled microservices
    quantity INT NOT NULL CHECK (quantity > 0)
);

-- Indexes for performance optimization
CREATE INDEX idx_orders_ref ON orders(order_ref);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);