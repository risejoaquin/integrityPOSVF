-- Migration for memberships
CREATE TABLE memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    plan_name TEXT NOT NULL,
    valid_until TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'active'
);
