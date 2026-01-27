CREATE TYPE event_status AS ENUM ('pending', 'delivered', 'failed');

ALTER TABLE events ADD COLUMN status event_status NOT NULL DEFAULT 'pending';
