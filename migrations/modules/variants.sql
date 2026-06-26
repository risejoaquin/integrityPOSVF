-- Migration for variants
CREATE TABLE IF NOT EXISTS retail_variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    size TEXT,
    color TEXT,
    stock INTEGER NOT NULL DEFAULT 0
);
