-- Initialize database with required extensions and default data
-- This script runs automatically when PostgreSQL starts for the first time

-- Enable UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create indexes for better performance (these will be created by migrations too)
-- This is just to ensure they exist if migrations haven't run yet

-- Note: The actual schema will be created by the migration service
-- This file is mainly for extensions and initial setup
