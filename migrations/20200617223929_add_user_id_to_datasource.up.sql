ALTER TABLE querycache_datasources
ADD COLUMN IF NOT EXISTS user_id uuid NOT NULL;
