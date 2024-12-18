CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    chat_id VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL,
    portfolio_url VARCHAR(255) NULL,
    specialization VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
