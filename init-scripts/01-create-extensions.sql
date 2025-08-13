-- Enable useful PostgreSQL extensions for the CRM database

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Cryptographic functions (useful for password hashing if needed)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Case-insensitive text
CREATE EXTENSION IF NOT EXISTS "citext";

-- Create application schema
CREATE SCHEMA IF NOT EXISTS crm;

-- Set search path
ALTER DATABASE crmdb SET search_path TO public, crm;

-- Grant permissions
GRANT ALL ON SCHEMA crm TO crmuser;
GRANT ALL ON SCHEMA public TO crmuser;