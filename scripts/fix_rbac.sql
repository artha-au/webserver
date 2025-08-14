-- Fix RBAC Initialization Script
-- Run this to properly set up roles and permissions

-- First, clean up any incorrect data
DELETE FROM user_roles WHERE role_id = 'role-admin' OR role_id = 'role-team-leader' OR role_id = 'role-team-member';
DELETE FROM role_permissions WHERE role_id = 'role-admin' OR role_id = 'role-team-leader' OR role_id = 'role-team-member';
DELETE FROM roles WHERE id IN ('role-admin', 'role-team-leader', 'role-team-member');

-- Create the roles with correct IDs
INSERT INTO roles (id, name, description, is_global, created_at)
VALUES 
    ('role-admin', 'Admin', 'Full system administrator', true, NOW()),
    ('role-team-leader', 'Team Leader', 'Team leader with management capabilities', false, NOW()),
    ('role-team-member', 'Team Member', 'Regular team member', false, NOW())
ON CONFLICT (id) DO NOTHING;

-- Attach permissions to admin role
INSERT INTO role_permissions (id, role_id, permission_id, created_at)
VALUES 
    ('rp-admin-teams-manage', 'role-admin', 'perm-teams-manage', NOW()),
    ('rp-admin-members-manage', 'role-admin', 'perm-members-manage', NOW()),
    ('rp-admin-timesheets-manage', 'role-admin', 'perm-timesheets-manage', NOW()),
    ('rp-admin-rosters-manage', 'role-admin', 'perm-rosters-manage', NOW())
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Attach permissions to team leader role
INSERT INTO role_permissions (id, role_id, permission_id, created_at)
VALUES 
    ('rp-leader-teams-read', 'role-team-leader', 'perm-teams-read', NOW()),
    ('rp-leader-teams-lead', 'role-team-leader', 'perm-teams-lead', NOW()),
    ('rp-leader-members-read', 'role-team-leader', 'perm-members-read', NOW()),
    ('rp-leader-timesheets-approve', 'role-team-leader', 'perm-timesheets-approve', NOW()),
    ('rp-leader-timesheets-create', 'role-team-leader', 'perm-timesheets-create', NOW()),
    ('rp-leader-timesheets-read', 'role-team-leader', 'perm-timesheets-read', NOW()),
    ('rp-leader-rosters-manage', 'role-team-leader', 'perm-rosters-manage', NOW()),
    ('rp-leader-rosters-read', 'role-team-leader', 'perm-rosters-read', NOW())
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Attach permissions to team member role
INSERT INTO role_permissions (id, role_id, permission_id, created_at)
VALUES 
    ('rp-member-teams-read', 'role-team-member', 'perm-teams-read', NOW()),
    ('rp-member-members-read', 'role-team-member', 'perm-members-read', NOW()),
    ('rp-member-timesheets-create', 'role-team-member', 'perm-timesheets-create', NOW()),
    ('rp-member-timesheets-read', 'role-team-member', 'perm-timesheets-read', NOW()),
    ('rp-member-rosters-read', 'role-team-member', 'perm-rosters-read', NOW())
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO user_roles (id, user_id, role_id, granted_by, granted_at)
VALUES ('ur-admin', 'user-admin', 'role-admin', 'user-admin', NOW())
ON CONFLICT (id) DO NOTHING;

-- Assign roles to demo users
INSERT INTO user_roles (id, user_id, role_id, granted_by, granted_at)
VALUES 
    ('ur-leader', 'user-leader', 'role-team-leader', 'user-admin', NOW()),
    ('ur-member', 'user-member', 'role-team-member', 'user-admin', NOW())
ON CONFLICT (id) DO NOTHING;

-- Verify the setup
SELECT 'Admin user roles:' as info;
SELECT ur.*, r.name as role_name 
FROM user_roles ur 
JOIN roles r ON ur.role_id = r.id 
WHERE ur.user_id = 'user-admin';

SELECT 'Admin role permissions:' as info;
SELECT rp.*, p.resource, p.action 
FROM role_permissions rp 
JOIN permissions p ON rp.permission_id = p.id 
WHERE rp.role_id = 'role-admin';