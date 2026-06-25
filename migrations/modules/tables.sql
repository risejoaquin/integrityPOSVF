CREATE TABLE tables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    capacity INTEGER,
    status TEXT NOT NULL DEFAULT 'available'
        CHECK (status IN ('available','occupied','reserved')),
    metadata JSONB DEFAULT '{}'
);

ALTER TABLE orders ADD COLUMN table_id UUID REFERENCES tables(id);
