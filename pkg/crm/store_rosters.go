package crm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Roster operations

func (s *SQLStore) CreateRoster(ctx context.Context, teamID string, req *CreateRosterRequest, createdBy string) (*Roster, error) {
	roster := &Roster{
		ID:          uuid.New().String(),
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      RosterStatusDraft,
		CreatedBy:   createdBy,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	metadataJSON, err := json.Marshal(roster.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO crm_rosters (id, team_id, name, description, start_date, end_date, status, created_by, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = s.db.ExecContext(ctx, query,
		roster.ID, roster.TeamID, roster.Name, roster.Description, roster.StartDate,
		roster.EndDate, roster.Status, roster.CreatedBy, metadataJSON, roster.CreatedAt, roster.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create roster: %w", err)
	}

	return roster, nil
}

func (s *SQLStore) GetRoster(ctx context.Context, rosterID string) (*Roster, error) {
	roster := &Roster{}
	var metadataJSON []byte

	query := `
		SELECT id, team_id, name, description, start_date, end_date, status, created_by, metadata, created_at, updated_at
		FROM crm_rosters WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, rosterID).Scan(
		&roster.ID, &roster.TeamID, &roster.Name, &roster.Description, &roster.StartDate,
		&roster.EndDate, &roster.Status, &roster.CreatedBy, &metadataJSON, &roster.CreatedAt, &roster.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("roster not found")
		}
		return nil, fmt.Errorf("failed to get roster: %w", err)
	}

	if err = json.Unmarshal(metadataJSON, &roster.Metadata); err != nil {
		roster.Metadata = make(map[string]interface{})
	}

	return roster, nil
}

func (s *SQLStore) ListTeamRosters(ctx context.Context, teamID string, limit, offset int) ([]*Roster, int, error) {
	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM crm_rosters WHERE team_id = $1"
	err := s.db.QueryRowContext(ctx, countQuery, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get roster count: %w", err)
	}

	// Get rosters
	query := `
		SELECT id, team_id, name, description, start_date, end_date, status, created_by, metadata, created_at, updated_at
		FROM crm_rosters 
		WHERE team_id = $1
		ORDER BY start_date DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list rosters: %w", err)
	}
	defer rows.Close()

	rosters := []*Roster{}
	for rows.Next() {
		roster := &Roster{}
		var metadataJSON []byte

		err := rows.Scan(
			&roster.ID, &roster.TeamID, &roster.Name, &roster.Description, &roster.StartDate,
			&roster.EndDate, &roster.Status, &roster.CreatedBy, &metadataJSON, &roster.CreatedAt, &roster.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan roster: %w", err)
		}

		if err = json.Unmarshal(metadataJSON, &roster.Metadata); err != nil {
			roster.Metadata = make(map[string]interface{})
		}

		rosters = append(rosters, roster)
	}

	return rosters, total, nil
}

func (s *SQLStore) UpdateRoster(ctx context.Context, rosterID string, req *UpdateRosterRequest) (*Roster, error) {
	// Get current roster
	roster, err := s.GetRoster(ctx, rosterID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		roster.Name = *req.Name
	}
	if req.Description != nil {
		roster.Description = *req.Description
	}
	if req.StartDate != nil {
		roster.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		roster.EndDate = *req.EndDate
	}
	if req.Status != nil {
		roster.Status = *req.Status
	}
	if req.Metadata != nil {
		roster.Metadata = req.Metadata
	}
	roster.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(roster.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE crm_rosters 
		SET name = $2, description = $3, start_date = $4, end_date = $5, status = $6, metadata = $7, updated_at = $8
		WHERE id = $1
	`
	_, err = s.db.ExecContext(ctx, query,
		roster.ID, roster.Name, roster.Description, roster.StartDate,
		roster.EndDate, roster.Status, metadataJSON, roster.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update roster: %w", err)
	}

	return roster, nil
}

func (s *SQLStore) DeleteRoster(ctx context.Context, rosterID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all roster shifts first
	_, err = tx.ExecContext(ctx, "DELETE FROM crm_roster_shifts WHERE roster_id = $1", rosterID)
	if err != nil {
		return fmt.Errorf("failed to delete roster shifts: %w", err)
	}

	// Delete the roster
	result, err := tx.ExecContext(ctx, "DELETE FROM crm_rosters WHERE id = $1", rosterID)
	if err != nil {
		return fmt.Errorf("failed to delete roster: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("roster not found")
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SQLStore) PublishRoster(ctx context.Context, rosterID string) error {
	query := `UPDATE crm_rosters SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, RosterStatusPublished, rosterID)
	if err != nil {
		return fmt.Errorf("failed to publish roster: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("roster not found")
	}

	return nil
}

// Roster shift operations

func (s *SQLStore) CreateRosterShift(ctx context.Context, rosterID string, req *CreateRosterShiftRequest) (*RosterShift, error) {
	shift := &RosterShift{
		ID:        uuid.New().String(),
		RosterID:  rosterID,
		UserID:    req.UserID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Role:      req.Role,
		Notes:     req.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO crm_roster_shifts (id, roster_id, user_id, start_time, end_time, role, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.ExecContext(ctx, query,
		shift.ID, shift.RosterID, shift.UserID, shift.StartTime, shift.EndTime,
		shift.Role, shift.Notes, shift.CreatedAt, shift.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create roster shift: %w", err)
	}

	return shift, nil
}

func (s *SQLStore) GetRosterShift(ctx context.Context, shiftID string) (*RosterShift, error) {
	shift := &RosterShift{}
	query := `
		SELECT id, roster_id, user_id, start_time, end_time, role, notes, created_at, updated_at
		FROM crm_roster_shifts WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, shiftID).Scan(
		&shift.ID, &shift.RosterID, &shift.UserID, &shift.StartTime, &shift.EndTime,
		&shift.Role, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("roster shift not found")
		}
		return nil, fmt.Errorf("failed to get roster shift: %w", err)
	}

	return shift, nil
}

func (s *SQLStore) ListRosterShifts(ctx context.Context, rosterID string) ([]*RosterShift, error) {
	query := `
		SELECT id, roster_id, user_id, start_time, end_time, role, notes, created_at, updated_at
		FROM crm_roster_shifts 
		WHERE roster_id = $1
		ORDER BY start_time ASC
	`
	rows, err := s.db.QueryContext(ctx, query, rosterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roster shifts: %w", err)
	}
	defer rows.Close()

	shifts := []*RosterShift{}
	for rows.Next() {
		shift := &RosterShift{}
		err := rows.Scan(
			&shift.ID, &shift.RosterID, &shift.UserID, &shift.StartTime, &shift.EndTime,
			&shift.Role, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan roster shift: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, nil
}

func (s *SQLStore) UpdateRosterShift(ctx context.Context, shiftID string, req *CreateRosterShiftRequest) (*RosterShift, error) {
	shift := &RosterShift{
		ID:        shiftID,
		UserID:    req.UserID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Role:      req.Role,
		Notes:     req.Notes,
		UpdatedAt: time.Now(),
	}

	query := `
		UPDATE crm_roster_shifts 
		SET user_id = $2, start_time = $3, end_time = $4, role = $5, notes = $6, updated_at = $7
		WHERE id = $1
		RETURNING roster_id, created_at
	`
	err := s.db.QueryRowContext(ctx, query,
		shift.ID, shift.UserID, shift.StartTime, shift.EndTime,
		shift.Role, shift.Notes, shift.UpdatedAt).Scan(&shift.RosterID, &shift.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("roster shift not found")
		}
		return nil, fmt.Errorf("failed to update roster shift: %w", err)
	}

	return shift, nil
}

func (s *SQLStore) DeleteRosterShift(ctx context.Context, shiftID string) error {
	query := `DELETE FROM crm_roster_shifts WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, shiftID)
	if err != nil {
		return fmt.Errorf("failed to delete roster shift: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("roster shift not found")
	}

	return nil
}