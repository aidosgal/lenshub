CREATE TABLE responses (
    id SERIAL PRIMARY KEY,
    user_id INT,
    order_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
