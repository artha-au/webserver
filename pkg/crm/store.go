package crm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Store defines the interface for CRM data operations
type Store interface {
	// Team operations
	CreateTeam(ctx context.Context, req *CreateTeamRequest, createdBy string) (*Team, error)
	GetTeam(ctx context.Context, teamID string) (*Team, error)
	ListTeams(ctx context.Context, limit, offset int) ([]*Team, int, error)
	UpdateTeam(ctx context.Context, teamID string, req *UpdateTeamRequest) (*Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
	ListUserTeams(ctx context.Context, userID string) ([]*Team, error)
	GetTeamStats(ctx context.Context, teamID string) (*TeamStats, error)

	// Team member operations
	AddTeamMember(ctx context.Context, teamID string, req *AddTeamMemberRequest, addedBy string) (*TeamMember, error)
	RemoveTeamMember(ctx context.Context, teamID, memberID string) error
	ListTeamMembers(ctx context.Context, teamID string) ([]*TeamMember, error)
	GetTeamMember(ctx context.Context, teamID, memberID string) (*TeamMember, error)
	UpdateTeamMember(ctx context.Context, teamID, memberID string, role string) (*TeamMember, error)
	IsTeamLeader(ctx context.Context, teamID, userID string) (bool, error)
	IsTeamMember(ctx context.Context, teamID, userID string) (bool, error)

	// Roster operations
	CreateRoster(ctx context.Context, teamID string, req *CreateRosterRequest, createdBy string) (*Roster, error)
	GetRoster(ctx context.Context, rosterID string) (*Roster, error)
	ListTeamRosters(ctx context.Context, teamID string, limit, offset int) ([]*Roster, int, error)
	UpdateRoster(ctx context.Context, rosterID string, req *UpdateRosterRequest) (*Roster, error)
	DeleteRoster(ctx context.Context, rosterID string) error
	PublishRoster(ctx context.Context, rosterID string) error

	// Roster shift operations
	CreateRosterShift(ctx context.Context, rosterID string, req *CreateRosterShiftRequest) (*RosterShift, error)
	GetRosterShift(ctx context.Context, shiftID string) (*RosterShift, error)
	ListRosterShifts(ctx context.Context, rosterID string) ([]*RosterShift, error)
	UpdateRosterShift(ctx context.Context, shiftID string, req *CreateRosterShiftRequest) (*RosterShift, error)
	DeleteRosterShift(ctx context.Context, shiftID string) error

	// Timesheet operations
	CreateTimesheet(ctx context.Context, userID, teamID string, req *CreateTimesheetRequest) (*Timesheet, error)
	GetTimesheet(ctx context.Context, timesheetID string) (*Timesheet, error)
	ListUserTimesheets(ctx context.Context, userID string, teamID string, limit, offset int) ([]*Timesheet, int, error)
	ListMemberTimesheets(ctx context.Context, memberID, teamID string, limit, offset int) ([]*Timesheet, int, error)
	UpdateTimesheet(ctx context.Context, timesheetID string, req *UpdateTimesheetRequest) (*Timesheet, error)
	DeleteTimesheet(ctx context.Context, timesheetID string) error
	SubmitTimesheet(ctx context.Context, timesheetID string) error
	ReviewTimesheet(ctx context.Context, timesheetID string, req *ReviewTimesheetRequest, reviewedBy string) error
	GetMemberStats(ctx context.Context, userID, teamID string) (*TeamMemberStats, error)

	// Utility operations
	Close() error
}

// SQLStore implements the Store interface using SQL database
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore creates a new SQL-based store
func NewSQLStore(db *sql.DB) Store {
	return &SQLStore{db: db}
}

// Team operations

func (s *SQLStore) CreateTeam(ctx context.Context, req *CreateTeamRequest, createdBy string) (*Team, error) {
	team := &Team{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		LeaderID:    req.LeaderID,
		Active:      true,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	metadataJSON, err := json.Marshal(team.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO crm_teams (id, name, description, leader_id, active, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, query,
		team.ID, team.Name, team.Description, team.LeaderID, team.Active,
		metadataJSON, team.CreatedAt, team.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add the leader as a team member
	memberReq := &AddTeamMemberRequest{
		UserID: req.LeaderID,
		Role:   RoleLeader,
	}
	_, err = s.addTeamMemberTx(ctx, tx, team.ID, memberReq, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to add team leader: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return team, nil
}

func (s *SQLStore) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	team := &Team{}
	var metadataJSON []byte

	query := `
		SELECT id, name, description, leader_id, active, metadata, created_at, updated_at
		FROM crm_teams WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, teamID).Scan(
		&team.ID, &team.Name, &team.Description, &team.LeaderID, &team.Active,
		&metadataJSON, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	if err = json.Unmarshal(metadataJSON, &team.Metadata); err != nil {
		team.Metadata = make(map[string]interface{})
	}

	return team, nil
}

func (s *SQLStore) ListTeams(ctx context.Context, limit, offset int) ([]*Team, int, error) {
	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM crm_teams WHERE active = true"
	err := s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get team count: %w", err)
	}

	// Get teams
	query := `
		SELECT id, name, description, leader_id, active, metadata, created_at, updated_at
		FROM crm_teams 
		WHERE active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	teams := []*Team{}
	for rows.Next() {
		team := &Team{}
		var metadataJSON []byte

		err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.LeaderID, &team.Active,
			&metadataJSON, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan team: %w", err)
		}

		if err = json.Unmarshal(metadataJSON, &team.Metadata); err != nil {
			team.Metadata = make(map[string]interface{})
		}

		teams = append(teams, team)
	}

	return teams, total, nil
}

func (s *SQLStore) UpdateTeam(ctx context.Context, teamID string, req *UpdateTeamRequest) (*Team, error) {
	// Get current team
	team, err := s.GetTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		team.Name = *req.Name
	}
	if req.Description != nil {
		team.Description = *req.Description
	}
	if req.LeaderID != nil {
		team.LeaderID = *req.LeaderID
	}
	if req.Active != nil {
		team.Active = *req.Active
	}
	if req.Metadata != nil {
		team.Metadata = req.Metadata
	}
	team.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(team.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE crm_teams 
		SET name = $2, description = $3, leader_id = $4, active = $5, metadata = $6, updated_at = $7
		WHERE id = $1
	`
	_, err = s.db.ExecContext(ctx, query,
		team.ID, team.Name, team.Description, team.LeaderID, team.Active,
		metadataJSON, team.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return team, nil
}

func (s *SQLStore) DeleteTeam(ctx context.Context, teamID string) error {
	// Soft delete by marking as inactive
	query := `UPDATE crm_teams SET active = false, updated_at = NOW() WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, teamID)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

func (s *SQLStore) ListUserTeams(ctx context.Context, userID string) ([]*Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.leader_id, t.active, t.metadata, t.created_at, t.updated_at
		FROM crm_teams t
		INNER JOIN crm_team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1 AND tm.active = true AND t.active = true
		ORDER BY t.name
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user teams: %w", err)
	}
	defer rows.Close()

	teams := []*Team{}
	for rows.Next() {
		team := &Team{}
		var metadataJSON []byte

		err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.LeaderID, &team.Active,
			&metadataJSON, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}

		if err = json.Unmarshal(metadataJSON, &team.Metadata); err != nil {
			team.Metadata = make(map[string]interface{})
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (s *SQLStore) GetTeamStats(ctx context.Context, teamID string) (*TeamStats, error) {
	stats := &TeamStats{TeamID: teamID}

	// Get member count
	memberCountQuery := `SELECT COUNT(*) FROM crm_team_members WHERE team_id = $1 AND active = true`
	err := s.db.QueryRowContext(ctx, memberCountQuery, teamID).Scan(&stats.MemberCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get member count: %w", err)
	}

	// Get active rosters count
	rosterCountQuery := `SELECT COUNT(*) FROM crm_rosters WHERE team_id = $1 AND status = 'published'`
	err = s.db.QueryRowContext(ctx, rosterCountQuery, teamID).Scan(&stats.ActiveRosters)
	if err != nil {
		return nil, fmt.Errorf("failed to get roster count: %w", err)
	}

	// Get pending reviews count
	pendingQuery := `SELECT COUNT(*) FROM crm_timesheets WHERE team_id = $1 AND status = 'submitted'`
	err = s.db.QueryRowContext(ctx, pendingQuery, teamID).Scan(&stats.PendingReviews)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reviews count: %w", err)
	}

	// Get member stats
	memberStatsQuery := `
		SELECT 
			tm.user_id,
			COALESCE(SUM(CASE WHEN ts.status = 'approved' THEN ts.hours ELSE 0 END), 0) as approved_hours,
			COALESCE(SUM(CASE WHEN ts.status = 'submitted' THEN ts.hours ELSE 0 END), 0) as pending_hours,
			COALESCE(SUM(CASE WHEN ts.status = 'rejected' THEN ts.hours ELSE 0 END), 0) as rejected_hours,
			COUNT(ts.id) as timesheet_count,
			MAX(ts.submitted_at) as last_submission
		FROM crm_team_members tm
		LEFT JOIN crm_timesheets ts ON tm.user_id = ts.user_id AND tm.team_id = ts.team_id
		WHERE tm.team_id = $1 AND tm.active = true
		GROUP BY tm.user_id
	`
	rows, err := s.db.QueryContext(ctx, memberStatsQuery, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member stats: %w", err)
	}
	defer rows.Close()

	stats.MemberStats = []TeamMemberStats{}
	for rows.Next() {
		memberStat := TeamMemberStats{}
		var lastSubmission *time.Time

		err := rows.Scan(
			&memberStat.UserID,
			&memberStat.ApprovedHours,
			&memberStat.PendingHours,
			&memberStat.RejectedHours,
			&memberStat.TimesheetCount,
			&lastSubmission)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member stats: %w", err)
		}

		memberStat.TotalHours = memberStat.ApprovedHours + memberStat.PendingHours + memberStat.RejectedHours
		memberStat.LastSubmission = lastSubmission

		stats.MemberStats = append(stats.MemberStats, memberStat)
	}

	return stats, nil
}

// Helper method for adding team member within a transaction
func (s *SQLStore) addTeamMemberTx(ctx context.Context, tx *sql.Tx, teamID string, req *AddTeamMemberRequest, addedBy string) (*TeamMember, error) {
	member := &TeamMember{
		ID:       uuid.New().String(),
		TeamID:   teamID,
		UserID:   req.UserID,
		Role:     req.Role,
		JoinedAt: time.Now(),
		Active:   true,
	}

	query := `
		INSERT INTO crm_team_members (id, team_id, user_id, role, joined_at, active)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (team_id, user_id) DO UPDATE SET 
		role = EXCLUDED.role, active = true
		RETURNING id, joined_at
	`
	err := tx.QueryRowContext(ctx, query,
		member.ID, member.TeamID, member.UserID, member.Role, member.JoinedAt, member.Active).
		Scan(&member.ID, &member.JoinedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	return member, nil
}

func (s *SQLStore) Close() error {
	return s.db.Close()
}