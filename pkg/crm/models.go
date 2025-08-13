package crm

import (
	"time"
)

// Team represents a work team in the CRM system
type Team struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	LeaderID    string                 `json:"leader_id" db:"leader_id"`
	Active      bool                   `json:"active" db:"active"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// TeamMember represents a team member relationship
type TeamMember struct {
	ID       string    `json:"id" db:"id"`
	TeamID   string    `json:"team_id" db:"team_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	Role     string    `json:"role" db:"role"` // "leader", "member"
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	Active   bool      `json:"active" db:"active"`
}

// Roster represents a work schedule for a team
type Roster struct {
	ID          string                 `json:"id" db:"id"`
	TeamID      string                 `json:"team_id" db:"team_id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	StartDate   time.Time              `json:"start_date" db:"start_date"`
	EndDate     time.Time              `json:"end_date" db:"end_date"`
	Status      string                 `json:"status" db:"status"` // "draft", "published", "archived"
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// RosterShift represents an individual shift in a roster
type RosterShift struct {
	ID        string    `json:"id" db:"id"`
	RosterID  string    `json:"roster_id" db:"roster_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	StartTime time.Time `json:"start_time" db:"start_time"`
	EndTime   time.Time `json:"end_time" db:"end_time"`
	Role      string    `json:"role" db:"role"`
	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Timesheet represents a timesheet entry for time tracking
type Timesheet struct {
	ID          string                 `json:"id" db:"id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	TeamID      string                 `json:"team_id" db:"team_id"`
	Date        time.Time              `json:"date" db:"date"`
	Hours       float64                `json:"hours" db:"hours"`
	Description string                 `json:"description" db:"description"`
	Status      string                 `json:"status" db:"status"` // "draft", "submitted", "approved", "rejected"
	SubmittedAt *time.Time             `json:"submitted_at" db:"submitted_at"`
	ReviewedAt  *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	ReviewedBy  *string                `json:"reviewed_by" db:"reviewed_by"`
	ReviewNotes string                 `json:"review_notes" db:"review_notes"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// CreateTeamRequest represents a request to create a new team
type CreateTeamRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	LeaderID    string                 `json:"leader_id" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTeamRequest represents a request to update a team
type UpdateTeamRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=1000"`
	LeaderID    *string                `json:"leader_id,omitempty"`
	Active      *bool                  `json:"active,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AddTeamMemberRequest represents a request to add a member to a team
type AddTeamMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=leader member"`
}

// CreateRosterRequest represents a request to create a new roster
type CreateRosterRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	StartDate   time.Time              `json:"start_date" validate:"required"`
	EndDate     time.Time              `json:"end_date" validate:"required,gtfield=StartDate"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateRosterRequest represents a request to update a roster
type UpdateRosterRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=1000"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	EndDate     *time.Time             `json:"end_date,omitempty"`
	Status      *string                `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRosterShiftRequest represents a request to create a roster shift
type CreateRosterShiftRequest struct {
	UserID    string    `json:"user_id" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	Role      string    `json:"role" validate:"required"`
	Notes     string    `json:"notes,omitempty" validate:"max=500"`
}

// CreateTimesheetRequest represents a request to create a timesheet
type CreateTimesheetRequest struct {
	Date        time.Time              `json:"date" validate:"required"`
	Hours       float64                `json:"hours" validate:"required,min=0,max=24"`
	Description string                 `json:"description" validate:"required,min=1,max=1000"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTimesheetRequest represents a request to update a timesheet
type UpdateTimesheetRequest struct {
	Date        *time.Time             `json:"date,omitempty"`
	Hours       *float64               `json:"hours,omitempty" validate:"omitempty,min=0,max=24"`
	Description *string                `json:"description,omitempty" validate:"omitempty,min=1,max=1000"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SubmitTimesheetRequest represents a request to submit a timesheet for review
type SubmitTimesheetRequest struct {
	// No additional fields needed - submission is an action
}

// ReviewTimesheetRequest represents a request to review a timesheet
type ReviewTimesheetRequest struct {
	Status      string `json:"status" validate:"required,oneof=approved rejected"`
	ReviewNotes string `json:"review_notes,omitempty" validate:"max=1000"`
}

// TeamMemberStats represents statistics for a team member
type TeamMemberStats struct {
	UserID          string  `json:"user_id"`
	TotalHours      float64 `json:"total_hours"`
	PendingHours    float64 `json:"pending_hours"`
	ApprovedHours   float64 `json:"approved_hours"`
	RejectedHours   float64 `json:"rejected_hours"`
	TimesheetCount  int     `json:"timesheet_count"`
	LastSubmission  *time.Time `json:"last_submission"`
}

// TeamStats represents statistics for a team
type TeamStats struct {
	TeamID         string             `json:"team_id"`
	MemberCount    int                `json:"member_count"`
	ActiveRosters  int                `json:"active_rosters"`
	PendingReviews int                `json:"pending_reviews"`
	MemberStats    []TeamMemberStats  `json:"member_stats"`
}

// Constants for status values
const (
	// Team member roles
	RoleLeader = "leader"
	RoleMember = "member"

	// Roster status
	RosterStatusDraft     = "draft"
	RosterStatusPublished = "published"
	RosterStatusArchived  = "archived"

	// Timesheet status
	TimesheetStatusDraft     = "draft"
	TimesheetStatusSubmitted = "submitted"
	TimesheetStatusApproved  = "approved"
	TimesheetStatusRejected  = "rejected"
)