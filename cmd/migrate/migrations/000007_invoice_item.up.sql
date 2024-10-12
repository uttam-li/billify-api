CREATE TABLE IF NOT EXISTS "invoice_item" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inv_id UUID NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
    prod_id UUID NOT NULL REFERENCES product(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL
);
