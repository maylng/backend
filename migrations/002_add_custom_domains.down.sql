-- Drop the custom domain reference from email_addresses
ALTER TABLE email_addresses DROP COLUMN IF EXISTS custom_domain_id;

-- Revert email_addresses status constraint
ALTER TABLE email_addresses 
DROP CONSTRAINT email_addresses_status_check;

ALTER TABLE email_addresses 
ADD CONSTRAINT email_addresses_status_check 
CHECK (status IN ('active', 'expired', 'disabled'));

-- Drop custom_domains table
DROP TABLE IF EXISTS custom_domains;
