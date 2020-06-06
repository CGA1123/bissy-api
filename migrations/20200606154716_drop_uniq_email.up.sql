DROP INDEX IF EXISTS auth_users_uniq_email_idx;
CREATE UNIQUE INDEX auth_users_uniq_github_id_idx on auth_users (github_id);
