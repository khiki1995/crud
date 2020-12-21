
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
    customer_id BIGINT NOT NULL REFERENCES customers,
    expire TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE managers
(
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    salary INTEGER NOT NULL CHECK (salary > 0),
    plan INTEGER NOT NULL CHECK (salary > 0),
    boss_id BIGINT REFERENCES managers,
    department TEXT,
    login TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE managers_tokens
(
    token TEXT NOT NULL,
    manager_id BIGINT NOT NULL REFERENCES managers,
    expire TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)