CREATE TABLE IF NOT EXISTS orders (
    order_num BIGINT PRIMARY KEY,
    status INT NOT NULL,
    accrual decimal (12, 6) NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id)
);