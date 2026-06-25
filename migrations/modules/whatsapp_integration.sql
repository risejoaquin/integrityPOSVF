-- Migration for whatsapp_integration
-- (Note: whatsapp_messages table is already in core, this is a trigger/placeholder)
ALTER TABLE customers ADD COLUMN whatsapp_opt_in BOOLEAN NOT NULL DEFAULT false;
