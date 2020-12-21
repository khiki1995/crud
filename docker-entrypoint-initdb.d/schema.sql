
CREATE TABLE customers
(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE customers_tokens
(
    token TEXT NOT NULL UNIQUE,
    customer_id BIGINT NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
    expire TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE managers
(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    salary INTEGER NOT NULL DEFAULT 0,
    plan INTEGER NOT NULL  DEFAULT 0,
    boss_id BIGINT REFERENCES managers,
    department TEXT,
    phone TEXT NOT NULL UNIQUE,
    password TEXT,
    roles   TEXT[] NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE managers_tokens
(
    token TEXT NOT NULL,
    manager_id BIGINT NOT NULL REFERENCES managers,
    expire TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products
(
    id    BIGSERIAL PRIMARY KEY,
    name  TEXT NOT NULL,
    price INTEGER NOT NULL DEFAULT 0,
    qty   INTEGER NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sales 
(
    id          BIGSERIAL PRIMARY KEY,
    manager_id  BIGINT NOT NULL REFERENCES managers,
    customer_id BIGINT NOT NULL,
    created     timestamp NOT NULL default current_timestamp 
);

CREATE TABLE sales_positions 
(
    id          BIGSERIAL PRIMARY KEY,
    product_id  BIGINT NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    sale_id     BIGINT NOT NULL REFERENCES sales (id) ON DELETE CASCADE,
    price       INTEGER NOT NULL,
    qty         INTEGER NOT NULL,
    created     timestamp NOT NULL default current_timestamp 
);