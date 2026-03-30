-- name: GetUnproccessedOrders :many
SELECT order_num, status, accrual, user_id, created_at
	FROM orders
	where status IN (0, 1);

-- name: GetOrdersByUser :many
SELECT order_num, status, accrual, user_id, created_at
	FROM orders
	where user_id = $1;


-- name: GetOrAddOrder :one
WITH inserted AS (
	INSERT INTO public.orders(
		order_num, status, accrual, user_id)
		VALUES ($1, $2, $3, $4)
	ON CONFLICT (order_num) DO NOTHING
	RETURNING order_num, status, accrual, user_id, created_at
)
SELECT order_num, status, accrual, user_id, created_at, true as is_new
FROM inserted
UNION ALL
SELECT order_num, status, accrual, user_id, created_at, false as is_new
FROM orders
WHERE order_num = $1
  AND NOT EXISTS (SELECT 1 FROM inserted);
