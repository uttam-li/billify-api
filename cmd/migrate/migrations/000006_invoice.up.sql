CREATE TABLE IF NOT EXISTS "invoice" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inv_no INT NOT NULL,
    buss_id UUID NOT NULL REFERENCES business(buss_id) ON DELETE CASCADE,
    cust_id UUID NOT NULL REFERENCES customer(id) ON DELETE CASCADE,
    total_amount NUMERIC(10, 2) NOT NULL,
    inv_date TIMESTAMPTZ NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    paid_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (buss_id, inv_no)
);
