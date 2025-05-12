-- Filename: migrations/000003_create_daily_goals_table.up.sql
CREATE TABLE daily_goals (
goal_id bigserial PRIMARY KEY,
user_id integer NOT NULL,
goal_text varchar NOT NULL,
target_date DATE NOT NULL,
is_completed boolean DEFAULT 'false',
created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
