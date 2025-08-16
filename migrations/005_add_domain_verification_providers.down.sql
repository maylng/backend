-- Remove provider-agnostic verification fields from custom_domains
ALTER TABLE custom_domains 
DROP COLUMN IF EXISTS verification_provider,
DROP COLUMN IF EXISTS provider_verification_status,
DROP COLUMN IF EXISTS provider_domain_id;

-- Drop index for verification provider
DROP INDEX IF EXISTS idx_custom_domains_verification_provider;