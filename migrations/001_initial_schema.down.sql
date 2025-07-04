-- Drop triggers
DROP TRIGGER IF EXISTS update_sent_emails_updated_at ON sent_emails;
DROP TRIGGER IF EXISTS update_email_addresses_updated_at ON email_addresses;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_email_analytics_occurred_at;
DROP INDEX IF EXISTS idx_email_analytics_event_type;
DROP INDEX IF EXISTS idx_email_analytics_email_id;

DROP INDEX IF EXISTS idx_sent_emails_thread_id;
DROP INDEX IF EXISTS idx_sent_emails_scheduled_at;
DROP INDEX IF EXISTS idx_sent_emails_status;
DROP INDEX IF EXISTS idx_sent_emails_from_email_id;
DROP INDEX IF EXISTS idx_sent_emails_account_id;

DROP INDEX IF EXISTS idx_email_addresses_expires_at;
DROP INDEX IF EXISTS idx_email_addresses_email;
DROP INDEX IF EXISTS idx_email_addresses_account_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS rate_limits;
DROP TABLE IF EXISTS email_analytics;
DROP TABLE IF EXISTS sent_emails;
DROP TABLE IF EXISTS email_addresses;
DROP TABLE IF EXISTS accounts;
