-- Filename: migrations/000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
   user_id bigserial PRIMARY KEY,
   username text NOT NULL,
   email citext NOT NULL,
   password_hash text NOT NULL,
   created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);


