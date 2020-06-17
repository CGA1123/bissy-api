ALTER TABLE querycache_queries
ADD COLUMN IF NOT EXISTS user_id uuid NOT NULL;
