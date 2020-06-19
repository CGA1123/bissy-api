CREATE TABLE IF NOT EXISTS auth_api_keys (
  id uuid NOT NULL,
  user_id uuid NOT NULL,
  name varchar(255) NOT NULL,
  last_used timestamp NOT NULL,
  created_at timestamp NOT NULL,
  key varchar(255) NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (user_id) REFERENCES auth_users(id) ON DELETE CASCADE ON UPDATE CASCADE
);
