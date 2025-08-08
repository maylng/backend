-- Add access_type column to email_addresses table
ALTER TABLE email_addresses 
ADD COLUMN access_type VARCHAR(20) DEFAULT 'agent' 
CHECK (access_type IN ('agent', 'individual'));

-- Update existing records to have 'agent' as default access type
UPDATE email_addresses SET access_type = 'agent' WHERE access_type IS NULL;

-- Make the column NOT NULL after setting default values
ALTER TABLE email_addresses ALTER COLUMN access_type SET NOT NULL;

-- Add index for access_type for potential filtering
CREATE INDEX idx_email_addresses_access_type ON email_addresses(access_type);
