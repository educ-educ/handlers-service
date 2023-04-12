CREATE TABLE IF NOT EXISTS handlers (
    id VARCHAR(128) PRIMARY KEY,
    socket_address TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS methods (
    id SERIAL PRIMARY KEY,
    handler_id VARCHAR(128) REFERENCES handlers (id) ON DELETE CASCADE NOT NULL,
    path_part TEXT,
    method_type TEXT
);