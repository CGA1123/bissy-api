CREATE TABLE IF NOT EXISTS querycache_adapters (
    id uuid NOT NULL,
    type varchar(255) NOT NULL,
    options varchar(255) NOT NULL,
    name varchar(255) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY (id)
);
