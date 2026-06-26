-- Migration for client_history
CREATE TABLE IF NOT EXISTS client_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    notes TEXT,
    formula_details JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
