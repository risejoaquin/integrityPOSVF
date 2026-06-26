-- Migration for pets
CREATE TABLE IF NOT EXISTS pets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES customers(id),
    name TEXT NOT NULL,
    species TEXT NOT NULL,
    breed TEXT,
    medical_history JSONB
);
