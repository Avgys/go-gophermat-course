CREATE TABLE IF NOT EXISTS balance (
    ID BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    amount decimal (12, 6) NOT NULL DEFAULT 0,
    withdrawn decimal (12, 6) NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL UNIQUE,
    CONSTRAINT balance_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS withdraw_history (
    ID BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    order_num BIGINT NOT NULL,
    withdraw_amount decimal (12, 6) NOT NULL,
    user_id BIGINT NOT NULL,    
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT withdraw_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id)
);