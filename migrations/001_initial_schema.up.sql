-- Create accounts table
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_hash VARCHAR(255) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'starter',
    email_limit_per_month INTEGER DEFAULT 1000,
    email_address_limit INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create email_addresses table
CREATE TABLE email_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('temporary', 'persistent')),
    prefix VARCHAR(100),
    domain VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'disabled')),
    expires_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create sent_emails table
CREATE TABLE sent_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    from_email_id UUID NOT NULL REFERENCES email_addresses(id),
    to_recipients JSONB NOT NULL,
    cc_recipients JSONB,
    bcc_recipients JSONB,
    subject VARCHAR(998) NOT NULL,
    text_content TEXT,
    html_content TEXT,
    attachments JSONB,
    headers JSONB,
    thread_id UUID,
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'queued' CHECK (status IN ('queued', 'sent', 'delivered', 'failed', 'scheduled')),
    provider_message_id VARCHAR(255),
    failure_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create email_analytics table
CREATE TABLE email_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_id UUID NOT NULL REFERENCES sent_emails(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('delivered', 'bounced', 'opened', 'clicked', 'complained', 'unsubscribed')),
    event_data JSONB,
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create rate_limits table
CREATE TABLE rate_limits (
    key VARCHAR(255) PRIMARY KEY,
    count INTEGER DEFAULT 1,
    window_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- Create indexes for performance
CREATE INDEX idx_email_addresses_account_id ON email_addresses(account_id);
CREATE INDEX idx_email_addresses_email ON email_addresses(email);
CREATE INDEX idx_email_addresses_expires_at ON email_addresses(expires_at);

CREATE INDEX idx_sent_emails_account_id ON sent_emails(account_id);
CREATE INDEX idx_sent_emails_from_email_id ON sent_emails(from_email_id);
CREATE INDEX idx_sent_emails_status ON sent_emails(status);
CREATE INDEX idx_sent_emails_scheduled_at ON sent_emails(scheduled_at);
CREATE INDEX idx_sent_emails_thread_id ON sent_emails(thread_id);

CREATE INDEX idx_email_analytics_email_id ON email_analytics(email_id);
CREATE INDEX idx_email_analytics_event_type ON email_analytics(event_type);
CREATE INDEX idx_email_analytics_occurred_at ON email_analytics(occurred_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_email_addresses_updated_at BEFORE UPDATE ON email_addresses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sent_emails_updated_at BEFORE UPDATE ON sent_emails FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
