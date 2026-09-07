-- Create the products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    
    -- Dimensions and weight are stored as DECIMAL(10,2) to prevent precision loss
    width DECIMAL(10, 2) NOT NULL,
    length DECIMAL(10, 2) NOT NULL,
    height DECIMAL(10, 2) NOT NULL,
    weight DECIMAL(10, 2) NOT NULL,
    
    -- Inventory tracking
    qty_available INT NOT NULL DEFAULT 0,
    qty_reserved INT NOT NULL DEFAULT 0,
    
    -- Audit timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create an index on SKU for faster lookups
CREATE INDEX idx_products_sku ON products(sku);