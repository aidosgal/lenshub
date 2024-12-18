CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NULL,
    description VARCHAR(10000) NULL,
    specialization VARCHAR(255) NOT NULL,
    location VARCHAR(255) NULL,
    user_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
