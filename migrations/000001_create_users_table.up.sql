-- Filename: migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
  user_id bigserial PRIMARY KEY,
  name text NOT NULL,
  email citext UNIQUE NOT NULL,
  password_hash bytea NOT NULL,
  activated bool NOT NUll,
  created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
