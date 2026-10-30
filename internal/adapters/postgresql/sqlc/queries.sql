-- name: ListProducts :many
SELECT
  *
FROM
  products;

-- name: FindProductByID :one
SELECT
  *
FROM
  products
WHERE
  id = $1;

-- name: CreateOrder :one
INSERT INTO orders (
  customer_id
) VALUES ($1) RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price_cents)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListOrders :many
SELECT
  *
FROM
  orders
ORDER BY
  id;

-- name: DecrementProductStock :one
UPDATE products
SET quantity = quantity - $2
WHERE id = $1 AND quantity >= $2
RETURNING price_in_cents;

-- name: ListOrderItems :many
SELECT
  *
FROM
  order_items
ORDER BY
  id;
