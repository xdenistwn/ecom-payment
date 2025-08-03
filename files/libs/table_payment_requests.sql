CREATE TABLE payment_requests (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    amount NUMERIC,
    user_email varchar(255),
    status varchar(50),
    retry_count INT,
    notes text,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP
)