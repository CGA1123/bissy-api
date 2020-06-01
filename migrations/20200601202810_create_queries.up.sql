CREATE TABLE IF NOT EXISTS querycache_queries (
    id uuid,
    adapter_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    last_refresh timestamp NOT NULL,
    lifetime varchar(255) NOT NULL,
    query text NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (adapter_id) REFERENCES querycache_adapters(id) ON DELETE CASCADE ON UPDATE CASCADE
);

