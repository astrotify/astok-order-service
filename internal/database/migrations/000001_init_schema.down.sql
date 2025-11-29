-- Drop triggers
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_order_products_updated_at ON order_products;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Drop indexes
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_order_products_order_id;

-- Drop tables
DROP TABLE IF EXISTS order_products;
DROP TABLE IF EXISTS orders;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
