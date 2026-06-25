-- Migration for rentals
CREATE TABLE rentals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id),
    product_id UUID NOT NULL REFERENCES products(id),
    due_date TIMESTAMPTZ,
    returned_at TIMESTAMPTZ,
    deposit_cents INTEGER NOT NULL DEFAULT 0
);
