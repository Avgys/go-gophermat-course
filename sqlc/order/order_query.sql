-- name: GetUnproccessedOrders :many
WITH picked AS (
	SELECT order_num
	FROM orders
	WHERE status IN (0, 1)
	ORDER BY status DESC, created_at
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
UPDATE orders o
SET status = 1
FROM picked p
WHERE o.order_num = p.order_num
RETURNING o.order_num, o.status, o.accrual, o.user_id, o.created_at;

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

-- name: UpdateOrder :exec
UPDATE public.orders
	SET status = $2, accrual = $3
	WHERE order_num = $1;