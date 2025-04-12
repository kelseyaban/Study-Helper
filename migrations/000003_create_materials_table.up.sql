-- Filename: migrations/000003_create_materials_table.up.sql
CREATE TABLE materials (
   material_id bigserial PRIMARY KEY,
   session_id integer,
   title text NOT NULL,
   file_path text,
   uploaded_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
   FOREIGN KEY (session_id) REFERENCES study_sessions(session_id)
);


