-- Filename: migrations/000001_create_study_sessions_table.up.sql
CREATE TABLE IF NOT EXISTS study_sessions (
 session_id bigserial PRIMARY KEY,
 title text NOT NULL,
 description text,
 subject text,
 start_date DATE NOT NULL,
 end_date DATE NOT NULL,
 is_completed boolean DEFAULT 'false',
 created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
