CREATE TABLE IF NOT EXISTS feedback (
    id bigserial PRIMARY KEY,
    fullname text NOT NULL,
    subject text NOT NULL,
    message text NOT NULL,
    email citext NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);