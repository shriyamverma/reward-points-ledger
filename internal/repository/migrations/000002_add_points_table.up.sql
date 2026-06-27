CREATE TABLE IF NOT EXISTS points (
    point_id SERIAL PRIMARY KEY,
    point_type_id INT UNIQUE NOT NULL,
    point_code VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);