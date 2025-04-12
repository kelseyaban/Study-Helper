-- Filename: migrations/000002_create_study_sessions_table.up.sql
CREATE TABLE IF NOT EXISTS study_sessions (
   session_id bigserial PRIMARY KEY,
   user_id integer NOT NULL,
   title text NOT NULL,
   description text,
   subject text,
   start_time TIMESTAMP,
   end_time TIMESTAMP,
   is_completed boolean DEFAULT 'false',
   created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
   FOREIGN KEY (user_id) REFERENCES users(user_id)
);


