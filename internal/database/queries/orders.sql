-- name: CreateOrder :one
INSERT INTO orders (user_id, status, total_amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1 LIMIT 1;

-- name: GetOrdersByUserID :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2
WHERE id = $1
RETURNING *;

-- name: CreateOrderProduct :one
INSERT INTO order_products (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrderProductsByOrderID :many
SELECT * FROM order_products
WHERE order_id = $1;

-- name: GetOrderWithProducts :one
SELECT 
    o.id,
    o.user_id,
    o.status,
    o.created_at,
    o.updated_at,
    json_agg(
        json_build_object(
            'id', op.id,
            'product_id', op.product_id,
            'quantity', op.quantity
        )
    ) FILTER (WHERE op.id IS NOT NULL) as products
FROM orders o
LEFT JOIN order_products op ON o.id = op.order_id
WHERE o.id = $1
GROUP BY o.id, o.user_id, o.status, o.created_at, o.updated_at;