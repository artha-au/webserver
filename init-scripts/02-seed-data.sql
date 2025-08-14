-- Seed data for CRM development and testing
-- This file is only executed on first database initialization

-- Note: The RBAC tables and initial permissions/roles are created by the Go application
-- This file can be used to add additional test data if needed

-- Example: Create some test namespaces (teams)
-- These will be created after the application initializes the RBAC tables
-- You can uncomment and modify as needed

/*
-- Wait for RBAC tables (this is just an example, the app will create these)
DO $$
BEGIN
    -- Check if namespaces table exists before inserting
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'namespaces') THEN
        -- Insert test teams
        INSERT INTO namespaces (id, type, name, created_at, updated_at)
        VALUES 
            ('team-engineering', 'team', 'Engineering Team', NOW(), NOW()),
            ('team-sales', 'team', 'Sales Team', NOW(), NOW()),
            ('team-support', 'team', 'Support Team', NOW(), NOW())
        ON CONFLICT (id) DO NOTHING;
    END IF;
END $$;
*/

-- Create a simple audit log table for the CRM
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    user_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

-- CRM tables will be created by the application's migration system
-- This ensures consistency between the init script and the application

-- Create update trigger function for updated_at columns (used by CRM migrations)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Grant permissions on CRM tables
GRANT ALL ON ALL TABLES IN SCHEMA public TO crmuser;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO crmuser;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO crmuser;