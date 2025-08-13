package crm

import (
	"context"
	"database/sql"
	"fmt"
)

// Team member operations

func (s *SQLStore) AddTeamMember(ctx context.Context, teamID string, req *AddTeamMemberRequest, addedBy string) (*TeamMember, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	member, err := s.addTeamMemberTx(ctx, tx, teamID, req, addedBy)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return member, nil
}

func (s *SQLStore) RemoveTeamMember(ctx context.Context, teamID, memberID string) error {
	query := `
		UPDATE crm_team_members 
		SET active = false 
		WHERE team_id = $1 AND (id = $2 OR user_id = $2)
	`
	result, err := s.db.ExecContext(ctx, query, teamID, memberID)
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team member not found")
	}

	return nil
}

func (s *SQLStore) ListTeamMembers(ctx context.Context, teamID string) ([]*TeamMember, error) {
	query := `
		SELECT id, team_id, user_id, role, joined_at, active
		FROM crm_team_members 
		WHERE team_id = $1 AND active = true
		ORDER BY role DESC, joined_at ASC
	`
	rows, err := s.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}
	defer rows.Close()

	members := []*TeamMember{}
	for rows.Next() {
		member := &TeamMember{}
		err := rows.Scan(
			&member.ID, &member.TeamID, &member.UserID, &member.Role, 
			&member.JoinedAt, &member.Active)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

func (s *SQLStore) GetTeamMember(ctx context.Context, teamID, memberID string) (*TeamMember, error) {
	member := &TeamMember{}
	query := `
		SELECT id, team_id, user_id, role, joined_at, active
		FROM crm_team_members 
		WHERE team_id = $1 AND (id = $2 OR user_id = $2) AND active = true
	`
	err := s.db.QueryRowContext(ctx, query, teamID, memberID).Scan(
		&member.ID, &member.TeamID, &member.UserID, &member.Role, 
		&member.JoinedAt, &member.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team member not found")
		}
		return nil, fmt.Errorf("failed to get team member: %w", err)
	}

	return member, nil
}

func (s *SQLStore) UpdateTeamMember(ctx context.Context, teamID, memberID string, role string) (*TeamMember, error) {
	query := `
		UPDATE crm_team_members 
		SET role = $3
		WHERE team_id = $1 AND (id = $2 OR user_id = $2) AND active = true
		RETURNING id, team_id, user_id, role, joined_at, active
	`
	member := &TeamMember{}
	err := s.db.QueryRowContext(ctx, query, teamID, memberID, role).Scan(
		&member.ID, &member.TeamID, &member.UserID, &member.Role, 
		&member.JoinedAt, &member.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team member not found")
		}
		return nil, fmt.Errorf("failed to update team member: %w", err)
	}

	return member, nil
}

func (s *SQLStore) IsTeamLeader(ctx context.Context, teamID, userID string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM crm_team_members 
		WHERE team_id = $1 AND user_id = $2 AND role = 'leader' AND active = true
	`
	err := s.db.QueryRowContext(ctx, query, teamID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check team leader: %w", err)
	}

	return count > 0, nil
}

func (s *SQLStore) IsTeamMember(ctx context.Context, teamID, userID string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM crm_team_members 
		WHERE team_id = $1 AND user_id = $2 AND active = true
	`
	err := s.db.QueryRowContext(ctx, query, teamID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check team membership: %w", err)
	}

	return count > 0, nil
}