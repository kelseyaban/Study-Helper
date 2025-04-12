-- Filename: migrations/000004_create_feedbacks_table.up.sql
CREATE TABLE feedbacks (
   feedback_id bigserial PRIMARY KEY,
   session_id integer NOT NULL,
   user_id integer NOT NULL,
   rating integer,
   comment text,
   created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
   FOREIGN KEY (session_id) REFERENCES study_sessions(session_id),
   FOREIGN KEY (user_id) REFERENCES users(user_id)
);
