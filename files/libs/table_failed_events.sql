CREATE TABLE failed_events (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    external_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    failed_type integer NOT NULL,
    notes TEXT,
    status integer NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP
)