package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/artha-au/webserver/pkg/auth"
)

// Data structures for API responses
type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Member struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	TeamID   string    `json:"team_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Timesheet struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TeamID      string    `json:"team_id"`
	Date        string    `json:"date"`
	Hours       float64   `json:"hours"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending", "approved", "rejected"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Roster struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Shifts    []Shift   `json:"shifts"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Shift struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Date     string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// Admin Team Handlers

func (h *APIHandlers) adminListTeams(w http.ResponseWriter, r *http.Request) {
	// Stub: Return sample teams
	teams := []Team{
		{
			ID:          "team-1",
			Name:        "Engineering Team",
			Description: "Product development team",
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "team-2",
			Name:        "Sales Team",
			Description: "Customer acquisition team",
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *APIHandlers) adminCreateTeam(w http.ResponseWriter, r *http.Request) {
	var team Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Stub: Generate ID and timestamps
	team.ID = "team-" + time.Now().Format("20060102150405")
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func (h *APIHandlers) adminGetTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	
	// Stub: Return sample team
	team := Team{
		ID:          teamID,
		Name:        "Sample Team",
		Description: "This is a sample team",
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *APIHandlers) adminUpdateTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	
	var team Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	team.ID = teamID
	team.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *APIHandlers) adminDeleteTeam(w http.ResponseWriter, r *http.Request) {
	// Stub: Just return success
	w.WriteHeader(http.StatusNoContent)
}

// Admin Team Member Handlers

func (h *APIHandlers) adminListTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	
	// Stub: Return sample members
	members := []Member{
		{
			ID:       "member-1",
			UserID:   "user-1",
			TeamID:   teamID,
			Role:     "team-leader",
			JoinedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:       "member-2",
			UserID:   "user-2",
			TeamID:   teamID,
			Role:     "team-member",
			JoinedAt: time.Now().Add(-20 * 24 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *APIHandlers) adminAddTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	
	var member Member
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member.ID = "member-" + time.Now().Format("20060102150405")
	member.TeamID = teamID
	member.JoinedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}

func (h *APIHandlers) adminGetTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	memberID := chi.URLParam(r, "memberID")
	
	// Stub: Return sample member
	member := Member{
		ID:       memberID,
		UserID:   "user-1",
		TeamID:   teamID,
		Role:     "team-member",
		JoinedAt: time.Now().Add(-30 * 24 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (h *APIHandlers) adminUpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	memberID := chi.URLParam(r, "memberID")
	
	var member Member
	if err := json.NewDecoder(r.Body).Decode(&member); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member.ID = memberID
	member.TeamID = teamID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

func (h *APIHandlers) adminRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	// Stub: Just return success
	w.WriteHeader(http.StatusNoContent)
}

// Regular Team Handlers

func (h *APIHandlers) listUserTeams(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Stub: Return teams based on user
	teams := []Team{
		{
			ID:          "team-1",
			Name:        "My Team",
			Description: "Team I belong to",
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *APIHandlers) listTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	
	// Stub: Return team members
	members := []Member{
		{
			ID:       "member-1",
			UserID:   "user-1",
			TeamID:   teamID,
			Role:     "team-leader",
			JoinedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:       "member-2",
			UserID:   "user-2",
			TeamID:   teamID,
			Role:     "team-member",
			JoinedAt: time.Now().Add(-20 * 24 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// Timesheet Handlers

func (h *APIHandlers) listMemberTimesheets(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	memberID := chi.URLParam(r, "memberID")
	
	// Stub: Return member's timesheets
	timesheets := []Timesheet{
		{
			ID:          "ts-1",
			UserID:      memberID,
			TeamID:      teamID,
			Date:        "2024-01-15",
			Hours:       8,
			Description: "Regular work",
			Status:      "pending",
			CreatedAt:   time.Now().Add(-2 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "ts-2",
			UserID:      memberID,
			TeamID:      teamID,
			Date:        "2024-01-14",
			Hours:       7.5,
			Description: "Project work",
			Status:      "approved",
			CreatedAt:   time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timesheets)
}

func (h *APIHandlers) approveTimesheet(w http.ResponseWriter, r *http.Request) {
	timesheetID := chi.URLParam(r, "timesheetID")
	
	// Stub: Return success message
	response := map[string]interface{}{
		"id":      timesheetID,
		"status":  "approved",
		"message": "Timesheet approved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandlers) rejectTimesheet(w http.ResponseWriter, r *http.Request) {
	timesheetID := chi.URLParam(r, "timesheetID")
	
	var request struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&request)
	
	// Stub: Return success message
	response := map[string]interface{}{
		"id":      timesheetID,
		"status":  "rejected",
		"reason":  request.Reason,
		"message": "Timesheet rejected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandlers) listMyTimesheets(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	teamID := r.Context().Value(contextKey("teamID")).(string)
	
	// Stub: Return user's own timesheets
	timesheets := []Timesheet{
		{
			ID:          "ts-user-1",
			UserID:      user.ID,
			TeamID:      teamID,
			Date:        "2024-01-15",
			Hours:       8,
			Description: "Daily standup, code review, feature development",
			Status:      "pending",
			CreatedAt:   time.Now().Add(-1 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "ts-user-2",
			UserID:      user.ID,
			TeamID:      teamID,
			Date:        "2024-01-14",
			Hours:       7,
			Description: "Bug fixes and testing",
			Status:      "approved",
			CreatedAt:   time.Now().Add(-2 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timesheets)
}

func (h *APIHandlers) createTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	teamID := r.Context().Value(contextKey("teamID")).(string)
	
	var timesheet Timesheet
	if err := json.NewDecoder(r.Body).Decode(&timesheet); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timesheet.ID = "ts-" + time.Now().Format("20060102150405")
	timesheet.UserID = user.ID
	timesheet.TeamID = teamID
	timesheet.Status = "pending"
	timesheet.CreatedAt = time.Now()
	timesheet.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(timesheet)
}

func (h *APIHandlers) getTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	teamID := r.Context().Value(contextKey("teamID")).(string)
	timesheetID := chi.URLParam(r, "timesheetID")
	
	// Stub: Return sample timesheet
	timesheet := Timesheet{
		ID:          timesheetID,
		UserID:      user.ID,
		TeamID:      teamID,
		Date:        "2024-01-15",
		Hours:       8,
		Description: "Daily work",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-1 * 24 * time.Hour),
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timesheet)
}

func (h *APIHandlers) updateTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	teamID := r.Context().Value(contextKey("teamID")).(string)
	timesheetID := chi.URLParam(r, "timesheetID")
	
	var timesheet Timesheet
	if err := json.NewDecoder(r.Body).Decode(&timesheet); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timesheet.ID = timesheetID
	timesheet.UserID = user.ID
	timesheet.TeamID = teamID
	timesheet.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timesheet)
}

func (h *APIHandlers) deleteTimesheet(w http.ResponseWriter, r *http.Request) {
	// Stub: Just return success
	w.WriteHeader(http.StatusNoContent)
}

// Roster Handlers

func (h *APIHandlers) listTeamRosters(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	
	// Stub: Return team rosters
	rosters := []Roster{
		{
			ID:        "roster-1",
			TeamID:    teamID,
			Name:      "Week 3 Roster",
			StartDate: "2024-01-15",
			EndDate:   "2024-01-21",
			Shifts: []Shift{
				{
					ID:        "shift-1",
					UserID:    "user-1",
					Date:      "2024-01-15",
					StartTime: "09:00",
					EndTime:   "17:00",
				},
			},
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "roster-2",
			TeamID:    teamID,
			Name:      "Week 4 Roster",
			StartDate: "2024-01-22",
			EndDate:   "2024-01-28",
			Shifts:    []Shift{},
			CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rosters)
}

func (h *APIHandlers) getRoster(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	rosterID := chi.URLParam(r, "rosterID")
	
	// Stub: Return sample roster
	roster := Roster{
		ID:        rosterID,
		TeamID:    teamID,
		Name:      "Sample Roster",
		StartDate: "2024-01-15",
		EndDate:   "2024-01-21",
		Shifts: []Shift{
			{
				ID:        "shift-1",
				UserID:    "user-1",
				Date:      "2024-01-15",
				StartTime: "09:00",
				EndTime:   "17:00",
			},
			{
				ID:        "shift-2",
				UserID:    "user-2",
				Date:      "2024-01-16",
				StartTime: "10:00",
				EndTime:   "18:00",
			},
		},
		CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roster)
}

func (h *APIHandlers) createRoster(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	
	var roster Roster
	if err := json.NewDecoder(r.Body).Decode(&roster); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	roster.ID = "roster-" + time.Now().Format("20060102150405")
	roster.TeamID = teamID
	roster.CreatedAt = time.Now()
	roster.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(roster)
}

func (h *APIHandlers) updateRoster(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	rosterID := chi.URLParam(r, "rosterID")
	
	var roster Roster
	if err := json.NewDecoder(r.Body).Decode(&roster); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	roster.ID = rosterID
	roster.TeamID = teamID
	roster.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roster)
}

func (h *APIHandlers) deleteRoster(w http.ResponseWriter, r *http.Request) {
	// Stub: Just return success
	w.WriteHeader(http.StatusNoContent)
}