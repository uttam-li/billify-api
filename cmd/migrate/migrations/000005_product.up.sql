CREATE TABLE IF NOT EXISTS "product" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buss_id UUID NOT NULL REFERENCES business(buss_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    tax_rate NUMERIC(5, 2) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    hsn_code VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (buss_id, name)
);
