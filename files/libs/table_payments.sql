CREATE TABLE payments (
    id BIGSERIAL,
    order_id BIGINT,
    user_id BIGINT,
    external_id TEXT UNIQUE NOT NULL,
    amount NUMERIC,
    status VARCHAR,
    expired_time TIMESTAMP,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP
)