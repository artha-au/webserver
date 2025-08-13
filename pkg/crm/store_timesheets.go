package crm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Timesheet operations

func (s *SQLStore) CreateTimesheet(ctx context.Context, userID, teamID string, req *CreateTimesheetRequest) (*Timesheet, error) {
	timesheet := &Timesheet{
		ID:          uuid.New().String(),
		UserID:      userID,
		TeamID:      teamID,
		Date:        req.Date,
		Hours:       req.Hours,
		Description: req.Description,
		Status:      TimesheetStatusDraft,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	metadataJSON, err := json.Marshal(timesheet.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO crm_timesheets (id, user_id, team_id, date, hours, description, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = s.db.ExecContext(ctx, query,
		timesheet.ID, timesheet.UserID, timesheet.TeamID, timesheet.Date, timesheet.Hours,
		timesheet.Description, timesheet.Status, metadataJSON, timesheet.CreatedAt, timesheet.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create timesheet: %w", err)
	}

	return timesheet, nil
}

func (s *SQLStore) GetTimesheet(ctx context.Context, timesheetID string) (*Timesheet, error) {
	timesheet := &Timesheet{}
	var metadataJSON []byte

	query := `
		SELECT id, user_id, team_id, date, hours, description, status, 
		       submitted_at, reviewed_at, reviewed_by, review_notes, metadata, created_at, updated_at
		FROM crm_timesheets WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, timesheetID).Scan(
		&timesheet.ID, &timesheet.UserID, &timesheet.TeamID, &timesheet.Date, &timesheet.Hours,
		&timesheet.Description, &timesheet.Status, &timesheet.SubmittedAt, &timesheet.ReviewedAt,
		&timesheet.ReviewedBy, &timesheet.ReviewNotes, &metadataJSON, &timesheet.CreatedAt, &timesheet.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("timesheet not found")
		}
		return nil, fmt.Errorf("failed to get timesheet: %w", err)
	}

	if err = json.Unmarshal(metadataJSON, &timesheet.Metadata); err != nil {
		timesheet.Metadata = make(map[string]interface{})
	}

	return timesheet, nil
}

func (s *SQLStore) ListUserTimesheets(ctx context.Context, userID string, teamID string, limit, offset int) ([]*Timesheet, int, error) {
	// Build query conditions
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	
	if teamID != "" {
		whereClause += " AND team_id = $2"
		args = append(args, teamID)
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM crm_timesheets %s", whereClause)
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get timesheet count: %w", err)
	}

	// Get timesheets
	args = append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, user_id, team_id, date, hours, description, status, 
		       submitted_at, reviewed_at, reviewed_by, review_notes, metadata, created_at, updated_at
		FROM crm_timesheets 
		%s
		ORDER BY date DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)-1, len(args))
	
	rows, err := s.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list timesheets: %w", err)
	}
	defer rows.Close()

	timesheets := []*Timesheet{}
	for rows.Next() {
		timesheet := &Timesheet{}
		var metadataJSON []byte

		err := rows.Scan(
			&timesheet.ID, &timesheet.UserID, &timesheet.TeamID, &timesheet.Date, &timesheet.Hours,
			&timesheet.Description, &timesheet.Status, &timesheet.SubmittedAt, &timesheet.ReviewedAt,
			&timesheet.ReviewedBy, &timesheet.ReviewNotes, &metadataJSON, &timesheet.CreatedAt, &timesheet.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan timesheet: %w", err)
		}

		if err = json.Unmarshal(metadataJSON, &timesheet.Metadata); err != nil {
			timesheet.Metadata = make(map[string]interface{})
		}

		timesheets = append(timesheets, timesheet)
	}

	return timesheets, total, nil
}

func (s *SQLStore) ListMemberTimesheets(ctx context.Context, memberID, teamID string, limit, offset int) ([]*Timesheet, int, error) {
	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM crm_timesheets WHERE user_id = $1 AND team_id = $2"
	err := s.db.QueryRowContext(ctx, countQuery, memberID, teamID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get member timesheet count: %w", err)
	}

	// Get timesheets
	query := `
		SELECT id, user_id, team_id, date, hours, description, status, 
		       submitted_at, reviewed_at, reviewed_by, review_notes, metadata, created_at, updated_at
		FROM crm_timesheets 
		WHERE user_id = $1 AND team_id = $2
		ORDER BY date DESC, created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := s.db.QueryContext(ctx, query, memberID, teamID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list member timesheets: %w", err)
	}
	defer rows.Close()

	timesheets := []*Timesheet{}
	for rows.Next() {
		timesheet := &Timesheet{}
		var metadataJSON []byte

		err := rows.Scan(
			&timesheet.ID, &timesheet.UserID, &timesheet.TeamID, &timesheet.Date, &timesheet.Hours,
			&timesheet.Description, &timesheet.Status, &timesheet.SubmittedAt, &timesheet.ReviewedAt,
			&timesheet.ReviewedBy, &timesheet.ReviewNotes, &metadataJSON, &timesheet.CreatedAt, &timesheet.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan member timesheet: %w", err)
		}

		if err = json.Unmarshal(metadataJSON, &timesheet.Metadata); err != nil {
			timesheet.Metadata = make(map[string]interface{})
		}

		timesheets = append(timesheets, timesheet)
	}

	return timesheets, total, nil
}

func (s *SQLStore) UpdateTimesheet(ctx context.Context, timesheetID string, req *UpdateTimesheetRequest) (*Timesheet, error) {
	// Get current timesheet
	timesheet, err := s.GetTimesheet(ctx, timesheetID)
	if err != nil {
		return nil, err
	}

	// Can only update draft timesheets
	if timesheet.Status != TimesheetStatusDraft {
		return nil, fmt.Errorf("can only update draft timesheets")
	}

	// Update fields if provided
	if req.Date != nil {
		timesheet.Date = *req.Date
	}
	if req.Hours != nil {
		timesheet.Hours = *req.Hours
	}
	if req.Description != nil {
		timesheet.Description = *req.Description
	}
	if req.Metadata != nil {
		timesheet.Metadata = req.Metadata
	}
	timesheet.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(timesheet.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE crm_timesheets 
		SET date = $2, hours = $3, description = $4, metadata = $5, updated_at = $6
		WHERE id = $1 AND status = 'draft'
	`
	result, err := s.db.ExecContext(ctx, query,
		timesheet.ID, timesheet.Date, timesheet.Hours, timesheet.Description, metadataJSON, timesheet.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update timesheet: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return nil, fmt.Errorf("timesheet not found or not in draft status")
	}

	return timesheet, nil
}

func (s *SQLStore) DeleteTimesheet(ctx context.Context, timesheetID string) error {
	query := `DELETE FROM crm_timesheets WHERE id = $1 AND status = 'draft'`
	result, err := s.db.ExecContext(ctx, query, timesheetID)
	if err != nil {
		return fmt.Errorf("failed to delete timesheet: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("timesheet not found or not in draft status")
	}

	return nil
}

func (s *SQLStore) SubmitTimesheet(ctx context.Context, timesheetID string) error {
	now := time.Now()
	query := `
		UPDATE crm_timesheets 
		SET status = $1, submitted_at = $2, updated_at = $3
		WHERE id = $4 AND status = 'draft'
	`
	result, err := s.db.ExecContext(ctx, query, TimesheetStatusSubmitted, now, now, timesheetID)
	if err != nil {
		return fmt.Errorf("failed to submit timesheet: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("timesheet not found or not in draft status")
	}

	return nil
}

func (s *SQLStore) ReviewTimesheet(ctx context.Context, timesheetID string, req *ReviewTimesheetRequest, reviewedBy string) error {
	now := time.Now()
	query := `
		UPDATE crm_timesheets 
		SET status = $1, reviewed_at = $2, reviewed_by = $3, review_notes = $4, updated_at = $5
		WHERE id = $6 AND status = 'submitted'
	`
	result, err := s.db.ExecContext(ctx, query, req.Status, now, reviewedBy, req.ReviewNotes, now, timesheetID)
	if err != nil {
		return fmt.Errorf("failed to review timesheet: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("timesheet not found or not in submitted status")
	}

	return nil
}

func (s *SQLStore) GetMemberStats(ctx context.Context, userID, teamID string) (*TeamMemberStats, error) {
	stats := &TeamMemberStats{UserID: userID}

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'approved' THEN hours ELSE 0 END), 0) as approved_hours,
			COALESCE(SUM(CASE WHEN status = 'submitted' THEN hours ELSE 0 END), 0) as pending_hours,
			COALESCE(SUM(CASE WHEN status = 'rejected' THEN hours ELSE 0 END), 0) as rejected_hours,
			COUNT(*) as timesheet_count,
			MAX(submitted_at) as last_submission
		FROM crm_timesheets
		WHERE user_id = $1 AND team_id = $2
	`
	
	var lastSubmission *time.Time
	err := s.db.QueryRowContext(ctx, query, userID, teamID).Scan(
		&stats.ApprovedHours,
		&stats.PendingHours,
		&stats.RejectedHours,
		&stats.TimesheetCount,
		&lastSubmission)
	if err != nil {
		return nil, fmt.Errorf("failed to get member stats: %w", err)
	}

	stats.TotalHours = stats.ApprovedHours + stats.PendingHours + stats.RejectedHours
	stats.LastSubmission = lastSubmission

	return stats, nil
}