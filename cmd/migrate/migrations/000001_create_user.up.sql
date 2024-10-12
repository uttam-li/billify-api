CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS "users" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name varchar(255),
    last_name varchar(255),
    email citext UNIQUE NOT NULL,
    password bytea,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT now()
);


