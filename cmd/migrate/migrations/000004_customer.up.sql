CREATE TABLE IF NOT EXISTS "customer" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buss_id UUID NOT NULL REFERENCES business(buss_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    gstno VARCHAR(15) NOT NULL,
    email citext NOT NULL,
    phone VARCHAR(20),
    baddress TEXT,
    saddress TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (buss_id, email, phone)
);