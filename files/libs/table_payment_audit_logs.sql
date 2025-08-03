CREATE TABLE payment_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    payment_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    external_id TEXT,
    event TEXT,
    actor TEXT,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);