CREATE TABLE IF NOT EXISTS auth_users (
    id uuid NOT NULL,
    email varchar(255) NOT NULL,
    created_at timestamp NOT NULL,
    PRIMARY KEY (id)
);
