CREATE TABLE IF NOT EXISTS "business" (
    buss_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    gstno VARCHAR(15) NOT NULL,
    company_email citext NOT NULL,
    company_phone VARCHAR(20) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    zip_code VARCHAR(10) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    account_no VARCHAR(20) NOT NULL,
    ifsc VARCHAR(11) NOT NULL,
    bank_branch VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
