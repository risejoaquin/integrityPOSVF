-- Migration for tips
ALTER TABLE orders ADD COLUMN tip_cents INTEGER NOT NULL DEFAULT 0;
