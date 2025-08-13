package crm

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Note: This is an example integration test
// In practice, you'd want to use a test database
func TestCRMIntegration(t *testing.T) {
	// Skip if no database available
	t.Skip("Integration test requires database setup")
	
	// This would be your test database connection
	db, err := sql.Open("postgres", "postgres://user:password@localhost/test_crm?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Create CRM service
	config := DefaultConfig()
	crmService, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("Failed to create CRM service: %v", err)
	}
	defer crmService.Close()

	ctx := context.Background()
	store := crmService.Store()

	// Test team creation
	teamReq := &CreateTeamRequest{
		Name:        "Test Team",
		Description: "A test team for integration testing",
		LeaderID:    "test-leader-123",
	}

	team, err := crmService.CreateTeamWithValidation(ctx, teamReq, "test-admin")
	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	if team.Name != "Test Team" {
		t.Errorf("Expected team name 'Test Team', got '%s'", team.Name)
	}

	// Test adding team member
	memberReq := &AddTeamMemberRequest{
		UserID: "test-member-456",
		Role:   RoleMember,
	}

	member, err := store.AddTeamMember(ctx, team.ID, memberReq, "test-admin")
	if err != nil {
		t.Fatalf("Failed to add team member: %v", err)
	}

	if member.Role != RoleMember {
		t.Errorf("Expected member role '%s', got '%s'", RoleMember, member.Role)
	}

	// Test roster creation
	rosterReq := &CreateRosterRequest{
		Name:        "Weekly Schedule",
		Description: "Test roster",
		StartDate:   time.Now().AddDate(0, 0, 1), // Tomorrow
		EndDate:     time.Now().AddDate(0, 0, 7), // Next week
	}

	roster, err := crmService.CreateRosterWithValidation(ctx, team.ID, rosterReq, "test-leader-123")
	if err != nil {
		t.Fatalf("Failed to create roster: %v", err)
	}

	if roster.Status != RosterStatusDraft {
		t.Errorf("Expected roster status '%s', got '%s'", RosterStatusDraft, roster.Status)
	}

	// Test timesheet creation
	timesheetReq := &CreateTimesheetRequest{
		Date:        time.Now().AddDate(0, 0, -1), // Yesterday
		Hours:       8.0,
		Description: "Worked on integration tests",
	}

	timesheet, err := crmService.CreateTimesheetWithValidation(ctx, "test-member-456", team.ID, timesheetReq)
	if err != nil {
		t.Fatalf("Failed to create timesheet: %v", err)
	}

	if timesheet.Hours != 8.0 {
		t.Errorf("Expected timesheet hours 8.0, got %f", timesheet.Hours)
	}

	// Test timesheet submission
	err = crmService.SubmitTimesheetWithValidation(ctx, "test-member-456", timesheet.ID)
	if err != nil {
		t.Fatalf("Failed to submit timesheet: %v", err)
	}

	// Verify timesheet status changed
	updatedTimesheet, err := store.GetTimesheet(ctx, timesheet.ID)
	if err != nil {
		t.Fatalf("Failed to get updated timesheet: %v", err)
	}

	if updatedTimesheet.Status != TimesheetStatusSubmitted {
		t.Errorf("Expected timesheet status '%s', got '%s'", 
			TimesheetStatusSubmitted, updatedTimesheet.Status)
	}

	// Test timesheet review
	reviewReq := &ReviewTimesheetRequest{
		Status:      TimesheetStatusApproved,
		ReviewNotes: "Good work on the integration tests!",
	}

	err = crmService.ReviewTimesheetWithValidation(ctx, "test-leader-123", timesheet.ID, reviewReq)
	if err != nil {
		t.Fatalf("Failed to review timesheet: %v", err)
	}

	// Verify final timesheet status
	finalTimesheet, err := store.GetTimesheet(ctx, timesheet.ID)
	if err != nil {
		t.Fatalf("Failed to get final timesheet: %v", err)
	}

	if finalTimesheet.Status != TimesheetStatusApproved {
		t.Errorf("Expected final timesheet status '%s', got '%s'", 
			TimesheetStatusApproved, finalTimesheet.Status)
	}

	if finalTimesheet.ReviewNotes != "Good work on the integration tests!" {
		t.Errorf("Expected review notes 'Good work on the integration tests!', got '%s'", 
			finalTimesheet.ReviewNotes)
	}

	t.Log("✅ All CRM integration tests passed successfully!")
}

// TestCRMValidation tests the validation logic without requiring a database
func TestCRMValidation(t *testing.T) {
	// Test team request validation
	tests := []struct {
		name    string
		req     *CreateTeamRequest
		wantErr bool
	}{
		{
			name: "valid team request",
			req: &CreateTeamRequest{
				Name:        "Valid Team",
				Description: "A valid team description",
				LeaderID:    "leader-123",
			},
			wantErr: false,
		},
		{
			name: "empty team name",
			req: &CreateTeamRequest{
				Name:     "",
				LeaderID: "leader-123",
			},
			wantErr: true,
		},
		{
			name: "empty leader ID",
			req: &CreateTeamRequest{
				Name:     "Team Name",
				LeaderID: "",
			},
			wantErr: true,
		},
		{
			name: "team name too long",
			req: &CreateTeamRequest{
				Name:     string(make([]byte, 256)), // 256 characters, too long
				LeaderID: "leader-123",
			},
			wantErr: true,
		},
	}

	crm := &CRM{} // No need for store/logger for validation tests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := crm.validateCreateTeamRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateTeamRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTimesheetValidation tests timesheet validation logic
func TestTimesheetValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateTimesheetRequest
		wantErr bool
	}{
		{
			name: "valid timesheet",
			req: &CreateTimesheetRequest{
				Date:        time.Now().AddDate(0, 0, -1), // Yesterday
				Hours:       8.0,
				Description: "Valid work description",
			},
			wantErr: false,
		},
		{
			name: "negative hours",
			req: &CreateTimesheetRequest{
				Date:        time.Now().AddDate(0, 0, -1),
				Hours:       -1.0,
				Description: "Invalid negative hours",
			},
			wantErr: true,
		},
		{
			name: "too many hours",
			req: &CreateTimesheetRequest{
				Date:        time.Now().AddDate(0, 0, -1),
				Hours:       25.0,
				Description: "Too many hours",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			req: &CreateTimesheetRequest{
				Date:        time.Now().AddDate(0, 0, -1),
				Hours:       8.0,
				Description: "",
			},
			wantErr: true,
		},
		{
			name: "future date",
			req: &CreateTimesheetRequest{
				Date:        time.Now().AddDate(0, 0, 7), // Next week
				Hours:       8.0,
				Description: "Future work",
			},
			wantErr: true,
		},
	}

	crm := &CRM{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := crm.validateCreateTimesheetRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateTimesheetRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}