-- Add provider-agnostic verification fields to custom_domains
ALTER TABLE custom_domains 
ADD COLUMN verification_provider VARCHAR(20) DEFAULT 'ses' CHECK (verification_provider IN ('ses', 'resend'));

-- Add provider-specific verification status (keeping SES fields for backward compatibility)
ALTER TABLE custom_domains 
ADD COLUMN provider_verification_status VARCHAR(50),
ADD COLUMN provider_domain_id VARCHAR(255);

-- Create index for verification provider
CREATE INDEX idx_custom_domains_verification_provider ON custom_domains(verification_provider);

-- Update existing records to set verification_provider to 'ses'
UPDATE custom_domains SET verification_provider = 'ses' WHERE verification_provider IS NULL;