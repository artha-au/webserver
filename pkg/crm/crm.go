package crm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// CRM represents the main CRM service
type CRM struct {
	store  Store
	logger *log.Logger
}

// Config holds CRM configuration options
type Config struct {
	AutoMigrate bool // Whether to automatically run migrations on startup
}

// DefaultConfig returns default CRM configuration
func DefaultConfig() *Config {
	return &Config{
		AutoMigrate: true,
	}
}

// New creates a new CRM service instance
func New(db *sql.DB, config *Config, logger *log.Logger) (*CRM, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if logger == nil {
		logger = log.New(log.Writer(), "[CRM] ", log.LstdFlags)
	}

	store := NewSQLStore(db)

	crm := &CRM{
		store:  store,
		logger: logger,
	}

	// Run migrations if enabled
	if config.AutoMigrate {
		migrator := NewMigrator(db, logger)
		if err := migrator.Init(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to run CRM migrations: %w", err)
		}
	}

	logger.Println("CRM service initialized successfully")
	return crm, nil
}

// Store returns the underlying store interface
func (c *CRM) Store() Store {
	return c.store
}

// ValidateTeamAccess checks if a user has access to a team
func (c *CRM) ValidateTeamAccess(ctx context.Context, userID, teamID string) error {
	isMember, err := c.store.IsTeamMember(ctx, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to check team membership: %w", err)
	}
	
	if !isMember {
		return fmt.Errorf("user is not a member of this team")
	}
	
	return nil
}

// ValidateTeamLeaderAccess checks if a user is a team leader
func (c *CRM) ValidateTeamLeaderAccess(ctx context.Context, userID, teamID string) error {
	isLeader, err := c.store.IsTeamLeader(ctx, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to check team leadership: %w", err)
	}
	
	if !isLeader {
		return fmt.Errorf("user is not a leader of this team")
	}
	
	return nil
}

// ValidateTimesheetAccess checks if a user can access a timesheet
func (c *CRM) ValidateTimesheetAccess(ctx context.Context, userID, timesheetID string) (*Timesheet, error) {
	timesheet, err := c.store.GetTimesheet(ctx, timesheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timesheet: %w", err)
	}
	
	// Users can access their own timesheets
	if timesheet.UserID == userID {
		return timesheet, nil
	}
	
	// Team leaders can access team member timesheets
	isLeader, err := c.store.IsTeamLeader(ctx, timesheet.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check team leadership: %w", err)
	}
	
	if !isLeader {
		return nil, fmt.Errorf("user cannot access this timesheet")
	}
	
	return timesheet, nil
}

// ValidateRosterAccess checks if a user can access a roster
func (c *CRM) ValidateRosterAccess(ctx context.Context, userID, rosterID string) (*Roster, error) {
	roster, err := c.store.GetRoster(ctx, rosterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roster: %w", err)
	}
	
	// Check if user is a team member
	isMember, err := c.store.IsTeamMember(ctx, roster.TeamID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check team membership: %w", err)
	}
	
	if !isMember {
		return nil, fmt.Errorf("user cannot access this roster")
	}
	
	return roster, nil
}

// Business logic methods

// CreateTeamWithValidation creates a team with business validation
func (c *CRM) CreateTeamWithValidation(ctx context.Context, req *CreateTeamRequest, createdBy string) (*Team, error) {
	// Validate request
	if err := c.validateCreateTeamRequest(req); err != nil {
		return nil, err
	}
	
	// Create the team
	team, err := c.store.CreateTeam(ctx, req, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}
	
	c.logger.Printf("Created team %s (%s) by user %s", team.Name, team.ID, createdBy)
	return team, nil
}

// CreateRosterWithValidation creates a roster with business validation
func (c *CRM) CreateRosterWithValidation(ctx context.Context, teamID string, req *CreateRosterRequest, createdBy string) (*Roster, error) {
	// Validate team leader access
	if err := c.ValidateTeamLeaderAccess(ctx, createdBy, teamID); err != nil {
		return nil, err
	}
	
	// Validate request
	if err := c.validateCreateRosterRequest(req); err != nil {
		return nil, err
	}
	
	// Check for overlapping rosters
	if err := c.validateRosterOverlap(ctx, teamID, req.StartDate, req.EndDate, ""); err != nil {
		return nil, err
	}
	
	roster, err := c.store.CreateRoster(ctx, teamID, req, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create roster: %w", err)
	}
	
	c.logger.Printf("Created roster %s (%s) for team %s by user %s", roster.Name, roster.ID, teamID, createdBy)
	return roster, nil
}

// CreateTimesheetWithValidation creates a timesheet with business validation
func (c *CRM) CreateTimesheetWithValidation(ctx context.Context, userID, teamID string, req *CreateTimesheetRequest) (*Timesheet, error) {
	// Validate team membership
	if err := c.ValidateTeamAccess(ctx, userID, teamID); err != nil {
		return nil, err
	}
	
	// Validate request
	if err := c.validateCreateTimesheetRequest(req); err != nil {
		return nil, err
	}
	
	// Check if timesheet already exists for this date
	existing, _, _ := c.store.ListUserTimesheets(ctx, userID, teamID, 1, 0)
	if len(existing) > 0 {
		for _, ts := range existing {
			if ts.Date.Format("2006-01-02") == req.Date.Format("2006-01-02") {
				return nil, fmt.Errorf("timesheet already exists for this date")
			}
		}
	}
	
	timesheet, err := c.store.CreateTimesheet(ctx, userID, teamID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create timesheet: %w", err)
	}
	
	c.logger.Printf("Created timesheet %s for user %s on %s", timesheet.ID, userID, req.Date.Format("2006-01-02"))
	return timesheet, nil
}

// SubmitTimesheetWithValidation submits a timesheet for review with validation
func (c *CRM) SubmitTimesheetWithValidation(ctx context.Context, userID, timesheetID string) error {
	timesheet, err := c.ValidateTimesheetAccess(ctx, userID, timesheetID)
	if err != nil {
		return err
	}
	
	// Can only submit own timesheets
	if timesheet.UserID != userID {
		return fmt.Errorf("can only submit own timesheets")
	}
	
	// Can only submit draft timesheets
	if timesheet.Status != TimesheetStatusDraft {
		return fmt.Errorf("can only submit draft timesheets")
	}
	
	if err := c.store.SubmitTimesheet(ctx, timesheetID); err != nil {
		return fmt.Errorf("failed to submit timesheet: %w", err)
	}
	
	c.logger.Printf("User %s submitted timesheet %s for review", userID, timesheetID)
	return nil
}

// ReviewTimesheetWithValidation reviews a timesheet with validation
func (c *CRM) ReviewTimesheetWithValidation(ctx context.Context, reviewerID, timesheetID string, req *ReviewTimesheetRequest) error {
	timesheet, err := c.ValidateTimesheetAccess(ctx, reviewerID, timesheetID)
	if err != nil {
		return err
	}
	
	// Must be a team leader to review
	if err := c.ValidateTeamLeaderAccess(ctx, reviewerID, timesheet.TeamID); err != nil {
		return err
	}
	
	// Can only review submitted timesheets
	if timesheet.Status != TimesheetStatusSubmitted {
		return fmt.Errorf("can only review submitted timesheets")
	}
	
	// Validate review request
	if err := c.validateReviewTimesheetRequest(req); err != nil {
		return err
	}
	
	if err := c.store.ReviewTimesheet(ctx, timesheetID, req, reviewerID); err != nil {
		return fmt.Errorf("failed to review timesheet: %w", err)
	}
	
	c.logger.Printf("User %s reviewed timesheet %s: %s", reviewerID, timesheetID, req.Status)
	return nil
}

// Validation methods

func (c *CRM) validateCreateTeamRequest(req *CreateTeamRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("team name is required")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("team name must be 255 characters or less")
	}
	if strings.TrimSpace(req.LeaderID) == "" {
		return fmt.Errorf("leader ID is required")
	}
	if len(req.Description) > 1000 {
		return fmt.Errorf("description must be 1000 characters or less")
	}
	return nil
}

func (c *CRM) validateCreateRosterRequest(req *CreateRosterRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("roster name is required")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("roster name must be 255 characters or less")
	}
	if len(req.Description) > 1000 {
		return fmt.Errorf("description must be 1000 characters or less")
	}
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}
	if req.StartDate.Before(time.Now().AddDate(0, 0, -1)) {
		return fmt.Errorf("start date cannot be in the past")
	}
	return nil
}

func (c *CRM) validateCreateTimesheetRequest(req *CreateTimesheetRequest) error {
	if req.Hours < 0 || req.Hours > 24 {
		return fmt.Errorf("hours must be between 0 and 24")
	}
	if strings.TrimSpace(req.Description) == "" {
		return fmt.Errorf("description is required")
	}
	if len(req.Description) > 1000 {
		return fmt.Errorf("description must be 1000 characters or less")
	}
	// Don't allow future dates beyond tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1)
	if req.Date.After(tomorrow) {
		return fmt.Errorf("cannot create timesheets for future dates")
	}
	// Don't allow very old dates (more than 1 year ago)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	if req.Date.Before(oneYearAgo) {
		return fmt.Errorf("cannot create timesheets for dates more than 1 year ago")
	}
	return nil
}

func (c *CRM) validateReviewTimesheetRequest(req *ReviewTimesheetRequest) error {
	if req.Status != TimesheetStatusApproved && req.Status != TimesheetStatusRejected {
		return fmt.Errorf("review status must be 'approved' or 'rejected'")
	}
	if len(req.ReviewNotes) > 1000 {
		return fmt.Errorf("review notes must be 1000 characters or less")
	}
	return nil
}

func (c *CRM) validateRosterOverlap(ctx context.Context, teamID string, startDate, endDate time.Time, excludeRosterID string) error {
	// Get existing rosters for the team
	rosters, _, err := c.store.ListTeamRosters(ctx, teamID, 100, 0) // Reasonable limit for checking overlaps
	if err != nil {
		return fmt.Errorf("failed to check roster overlaps: %w", err)
	}
	
	for _, roster := range rosters {
		// Skip if this is the roster we're updating
		if excludeRosterID != "" && roster.ID == excludeRosterID {
			continue
		}
		
		// Skip archived rosters
		if roster.Status == RosterStatusArchived {
			continue
		}
		
		// Check for overlap
		if startDate.Before(roster.EndDate) && endDate.After(roster.StartDate) {
			return fmt.Errorf("roster dates overlap with existing roster '%s'", roster.Name)
		}
	}
	
	return nil
}

// Close closes the CRM service and its resources
func (c *CRM) Close() error {
	return c.store.Close()
}