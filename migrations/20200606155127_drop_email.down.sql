ALTER TABLE auth_users
ADD COLUMN IF NOT EXISTS email varchar(255) NOT NULL;
