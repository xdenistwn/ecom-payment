CREATE TABLE payment_anomalies (
    id SERIAL PRIMARY KEY,
    order_id BIGINT,
    external_id TEXT,
    anomaly_type INTEGER,
    notes text,
    statis integer,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP
)