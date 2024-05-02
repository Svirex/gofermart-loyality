CREATE TABLE IF NOT EXISTS balance (
    id SERIAL PRIMARY KEY,
    uid INT REFERENCES users (id) UNIQUE,
    current NUMERIC(20, 10) CHECK (current >= 0) DEFAULT 0.00,
    withdrawn NUMERIC(20, 10) DEFAULT 0.00
);