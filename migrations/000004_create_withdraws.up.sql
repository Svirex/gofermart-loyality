CREATE TABLE IF NOT EXISTS withdraws (
    id SERIAL PRIMARY KEY,
    order_num TEXT UNIQUE,
    sum NUMERIC(20, 10),
    processed_at TIMESTAMP DEFAULT NOW()
);

