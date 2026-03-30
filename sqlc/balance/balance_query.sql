-- name: GetWithdrawals :many
SELECT order_num, withdraw_amount, user_id, created_at
	FROM public.withdraw_history
	where user_id = $1;

-- name: GetBalance :one
SELECT id, balance, user_id
	FROM public.balance
	where user_id = $1;

-- name: TryDecreaseBalance :one
WITH updated AS (
	UPDATE public.balance
	SET balance = balance - $2
	WHERE user_id = $1
	  AND balance >= $2
	RETURNING 1
)
SELECT EXISTS (SELECT 1 FROM updated) AS decreased;

-- name: CreateBalance :exec
INSERT INTO public.balance(id, balance, user_id)
	VALUES ($1, $2, $3);

-- name: InsertWithdrawal :exec
INSERT INTO public.withdraw_history(order_num, withdraw_amount, user_id)
	VALUES ($1, $2, $3);