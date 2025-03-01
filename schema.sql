CREATE TABLE IF NOT EXISTS authors (
    id SERIAL PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    category_id INT,
    created_at TIMESTAMPTZ,
    message TEXT,
    author_id INT,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FOREIGN KEY (author_id) REFERENCES authors(id)
);