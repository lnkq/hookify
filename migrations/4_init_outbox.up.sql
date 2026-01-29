CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL,
    webhook_id INT NOT NULL,
    payload JSONB NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_next_attempt_at ON outbox (next_attempt_at);