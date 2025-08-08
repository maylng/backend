-- Remove access_type column from email_addresses table
DROP INDEX IF EXISTS idx_email_addresses_access_type;
ALTER TABLE email_addresses DROP COLUMN IF EXISTS access_type;
