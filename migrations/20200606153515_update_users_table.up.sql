ALTER TABLE auth_users
ADD COLUMN IF NOT EXISTS github_id varchar(255) NOT NULL,
ADD COLUMN IF NOT EXISTS name varchar(255) NOT NULL;
