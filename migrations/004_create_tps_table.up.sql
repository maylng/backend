-- Create TPS (Third Party Software) table
CREATE TABLE tps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_address_id UUID NOT NULL REFERENCES email_addresses(id) ON DELETE CASCADE,
    service_name VARCHAR(100) NOT NULL,
    service_type VARCHAR(50) NOT NULL,
        service_url VARCHAR(500) NOT NULL,
    has_premium BOOLEAN DEFAULT FALSE,
    is_premium BOOLEAN DEFAULT FALSE,
    description TEXT,
    api_key TEXT,
    username VARCHAR(100),
    password TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'pending', 'failed', 'suspended')),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tps_email_address_id ON tps(email_address_id);
CREATE INDEX idx_tps_service_name ON tps(service_name);
CREATE INDEX idx_tps_status ON tps(status);

-- Trigger for updated_at
CREATE TRIGGER update_tps_updated_at BEFORE UPDATE ON tps FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
