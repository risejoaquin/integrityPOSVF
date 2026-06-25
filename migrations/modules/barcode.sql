-- Migration for barcode
ALTER TABLE products ADD COLUMN barcode TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode) WHERE barcode IS NOT NULL;
