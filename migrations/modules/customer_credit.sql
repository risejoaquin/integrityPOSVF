-- Migration for customer_credit
ALTER TABLE customers ADD COLUMN credit_limit_cents INTEGER NOT NULL DEFAULT 0;
ALTER TABLE customers ADD COLUMN current_balance_cents INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS customer_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    amount_cents INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('charge', 'payment')),
    order_id UUID REFERENCES orders(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
