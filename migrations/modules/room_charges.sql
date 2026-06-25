-- Migration for room_charges
ALTER TABLE orders ADD COLUMN room_id UUID REFERENCES rooms(id);
