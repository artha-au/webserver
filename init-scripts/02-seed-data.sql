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

-- Create CRM-specific tables (optional - extend as needed)
CREATE TABLE IF NOT EXISTS crm_teams (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS crm_timesheets (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    team_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    hours DECIMAL(4,2) NOT NULL CHECK (hours >= 0 AND hours <= 24),
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    approved_by VARCHAR(255),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, team_id, date)
);

CREATE TABLE IF NOT EXISTS crm_rosters (
    id VARCHAR(255) PRIMARY KEY,
    team_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_date >= start_date)
);

CREATE TABLE IF NOT EXISTS crm_roster_shifts (
    id VARCHAR(255) PRIMARY KEY,
    roster_id VARCHAR(255) NOT NULL REFERENCES crm_rosters(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    shift_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CHECK (end_time > start_time)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_timesheets_user_team ON crm_timesheets(user_id, team_id);
CREATE INDEX IF NOT EXISTS idx_timesheets_date ON crm_timesheets(date);
CREATE INDEX IF NOT EXISTS idx_timesheets_status ON crm_timesheets(status);
CREATE INDEX IF NOT EXISTS idx_rosters_team ON crm_rosters(team_id);
CREATE INDEX IF NOT EXISTS idx_rosters_dates ON crm_rosters(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_roster_shifts_roster ON crm_roster_shifts(roster_id);
CREATE INDEX IF NOT EXISTS idx_roster_shifts_user ON crm_roster_shifts(user_id);
CREATE INDEX IF NOT EXISTS idx_roster_shifts_date ON crm_roster_shifts(shift_date);

-- Create update trigger for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update trigger to CRM tables
DROP TRIGGER IF EXISTS update_crm_teams_updated_at ON crm_teams;
CREATE TRIGGER update_crm_teams_updated_at
    BEFORE UPDATE ON crm_teams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_crm_timesheets_updated_at ON crm_timesheets;
CREATE TRIGGER update_crm_timesheets_updated_at
    BEFORE UPDATE ON crm_timesheets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_crm_rosters_updated_at ON crm_rosters;
CREATE TRIGGER update_crm_rosters_updated_at
    BEFORE UPDATE ON crm_rosters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Grant permissions on CRM tables
GRANT ALL ON ALL TABLES IN SCHEMA public TO crmuser;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO crmuser;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO crmuser;