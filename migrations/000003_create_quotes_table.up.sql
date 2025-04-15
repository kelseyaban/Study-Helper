-- Filename: migrations/000003_create_quotes_table.up.sql
CREATE TABLE quotes(
quote_id bigserial PRIMARY KEY,
content text NOT NULL,
created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
