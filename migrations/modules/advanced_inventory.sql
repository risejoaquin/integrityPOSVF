-- Migration for advanced_inventory
CREATE TABLE IF NOT EXISTS product_kits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_product_id UUID NOT NULL REFERENCES products(id),
    child_product_id UUID NOT NULL REFERENCES products(id),
    quantity_required INTEGER NOT NULL
);
