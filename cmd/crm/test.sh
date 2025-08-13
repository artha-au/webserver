#!/bin/bash

# CRM API Test Script
# This script demonstrates how to interact with the CRM API

BASE_URL="http://localhost:8080"
TOKEN=""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}CRM API Test Script${NC}"
echo "========================"

# Health check
echo -e "\n${GREEN}1. Health Check${NC}"
curl -s "$BASE_URL/health" | jq '.'

# Get auth token (you'll need to implement actual login)
# For testing, you can manually generate a token for the admin user
echo -e "\n${GREEN}2. Authentication${NC}"
echo "Note: You'll need to implement actual login endpoint or generate a token manually"
echo "Example: POST $BASE_URL/auth/token with credentials"

# For demo purposes, set a dummy token (replace with actual token)
TOKEN="your-jwt-token-here"

# Admin APIs (require admin role)
echo -e "\n${GREEN}3. Admin APIs${NC}"
echo "List all teams (admin only):"
echo "curl -H \"Authorization: Bearer \$TOKEN\" $BASE_URL/api/v1/admin/teams"

echo -e "\nCreate a new team (admin only):"
echo "curl -X POST -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" \\"
echo "  -d '{\"name\":\"New Team\",\"description\":\"Test team\"}' \\"
echo "  $BASE_URL/api/v1/admin/teams"

# Regular user APIs
echo -e "\n${GREEN}4. Regular User APIs${NC}"
echo "List user's teams:"
echo "curl -H \"Authorization: Bearer \$TOKEN\" $BASE_URL/api/v1/teams"

echo -e "\nList team members:"
echo "curl -H \"Authorization: Bearer \$TOKEN\" $BASE_URL/api/v1/teams/{teamID}/members"

echo -e "\nCreate a timesheet:"
echo "curl -X POST -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" \\"
echo "  -d '{\"date\":\"2024-01-15\",\"hours\":8,\"description\":\"Daily work\"}' \\"
echo "  $BASE_URL/api/v1/teams/{teamID}/timesheets"

# Team Leader APIs
echo -e "\n${GREEN}5. Team Leader APIs${NC}"
echo "List member timesheets (team leader only):"
echo "curl -H \"Authorization: Bearer \$TOKEN\" \\"
echo "  $BASE_URL/api/v1/teams/{teamID}/members/{memberID}/timesheets"

echo -e "\nApprove a timesheet (team leader only):"
echo "curl -X POST -H \"Authorization: Bearer \$TOKEN\" \\"
echo "  $BASE_URL/api/v1/teams/{teamID}/members/{memberID}/timesheets/{timesheetID}/approve"

echo -e "\nCreate a roster (team leader only):"
echo "curl -X POST -H \"Authorization: Bearer \$TOKEN\" -H \"Content-Type: application/json\" \\"
echo "  -d '{\"name\":\"Week 4 Roster\",\"start_date\":\"2024-01-22\",\"end_date\":\"2024-01-28\"}' \\"
echo "  $BASE_URL/api/v1/teams/{teamID}/rosters"

echo -e "\n${BLUE}Complete!${NC}"
echo "Replace {teamID}, {memberID}, {timesheetID} with actual IDs"
echo "Remember to set a valid JWT token in the Authorization header"