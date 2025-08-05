-- Create custom_domains table
CREATE TABLE custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'failed', 'disabled')),
    verification_token VARCHAR(255),
    dkim_tokens JSONB,
    dns_records JSONB,
    ses_verification_status VARCHAR(50),
    ses_dkim_verification_status VARCHAR(50),
    verification_attempted_at TIMESTAMP,
    verified_at TIMESTAMP,
    failure_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(account_id, domain)
);

-- Add custom_domain_id to email_addresses table
ALTER TABLE email_addresses 
ADD COLUMN custom_domain_id UUID REFERENCES custom_domains(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX idx_custom_domains_account_id ON custom_domains(account_id);
CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX idx_custom_domains_status ON custom_domains(status);
CREATE INDEX idx_email_addresses_custom_domain_id ON email_addresses(custom_domain_id);

-- Create trigger for updated_at on custom_domains
CREATE TRIGGER update_custom_domains_updated_at 
    BEFORE UPDATE ON custom_domains 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Update email_addresses status enum to include 'verification_pending'
ALTER TABLE email_addresses 
DROP CONSTRAINT email_addresses_status_check;

ALTER TABLE email_addresses 
ADD CONSTRAINT email_addresses_status_check 
CHECK (status IN ('active', 'expired', 'disabled', 'verification_pending'));
