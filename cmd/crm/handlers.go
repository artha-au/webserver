package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/artha-au/webserver/pkg/auth"
	"github.com/artha-au/webserver/pkg/crm"
)

// Helper functions for parsing query parameters

func getLimit(r *http.Request) int {
	limit := 50 // default
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	return limit
}

func getOffset(r *http.Request) int {
	offset := 0 // default
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return offset
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Admin Team Handlers

func (h *APIHandlers) adminListTeams(w http.ResponseWriter, r *http.Request) {
	limit := getLimit(r)
	offset := getOffset(r)

	teams, total, err := h.crm.Store().ListTeams(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list teams: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"teams":  teams,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) adminCreateTeam(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req crm.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	team, err := h.crm.CreateTeamWithValidation(r.Context(), &req, user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to create team: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, team)
}

func (h *APIHandlers) adminGetTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	team, err := h.crm.Store().GetTeam(r.Context(), teamID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Team not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get team: "+err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, team)
}

func (h *APIHandlers) adminUpdateTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	var req crm.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	team, err := h.crm.Store().UpdateTeam(r.Context(), teamID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Team not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to update team: "+err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, team)
}

func (h *APIHandlers) adminDeleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	err := h.crm.Store().DeleteTeam(r.Context(), teamID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Team not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete team: "+err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Admin Team Member Handlers

func (h *APIHandlers) adminListTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	members, err := h.crm.Store().ListTeamMembers(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list team members: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (h *APIHandlers) adminAddTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req crm.AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	member, err := h.crm.Store().AddTeamMember(r.Context(), teamID, &req, user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to add team member: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, member)
}

func (h *APIHandlers) adminGetTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	memberID := chi.URLParam(r, "memberID")
	if teamID == "" || memberID == "" {
		writeError(w, http.StatusBadRequest, "Team ID and Member ID are required")
		return
	}

	member, err := h.crm.Store().GetTeamMember(r.Context(), teamID, memberID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Team member not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get team member: "+err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, member)
}

func (h *APIHandlers) adminUpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	memberID := chi.URLParam(r, "memberID")
	if teamID == "" || memberID == "" {
		writeError(w, http.StatusBadRequest, "Team ID and Member ID are required")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	member, err := h.crm.Store().UpdateTeamMember(r.Context(), teamID, memberID, req.Role)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Team member not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to update team member: "+err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, member)
}

func (h *APIHandlers) adminRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")
	memberID := chi.URLParam(r, "memberID")
	if teamID == "" || memberID == "" {
		writeError(w, http.StatusBadRequest, "Team ID and Member ID are required")
		return
	}

	err := h.crm.Store().RemoveTeamMember(r.Context(), teamID, memberID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Team member not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to remove team member: "+err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Regular Team Handlers

func (h *APIHandlers) listUserTeams(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	teams, err := h.crm.Store().ListUserTeams(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list user teams: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, teams)
}

func (h *APIHandlers) listTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	members, err := h.crm.Store().ListTeamMembers(r.Context(), teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list team members: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, members)
}

// Timesheet Handlers

func (h *APIHandlers) listMemberTimesheets(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	memberID := chi.URLParam(r, "memberID")
	if teamID == "" || memberID == "" {
		writeError(w, http.StatusBadRequest, "Team ID and Member ID are required")
		return
	}

	limit := getLimit(r)
	offset := getOffset(r)

	timesheets, total, err := h.crm.Store().ListMemberTimesheets(r.Context(), memberID, teamID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list member timesheets: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"timesheets": timesheets,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) approveTimesheet(w http.ResponseWriter, r *http.Request) {
	timesheetID := chi.URLParam(r, "timesheetID")
	if timesheetID == "" {
		writeError(w, http.StatusBadRequest, "Timesheet ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	req := &crm.ReviewTimesheetRequest{
		Status:      crm.TimesheetStatusApproved,
		ReviewNotes: "Approved",
	}

	err := h.crm.ReviewTimesheetWithValidation(r.Context(), user.ID, timesheetID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Timesheet not found")
		} else if strings.Contains(err.Error(), "not a leader") {
			writeError(w, http.StatusForbidden, "Only team leaders can approve timesheets")
		} else {
			writeError(w, http.StatusBadRequest, "Failed to approve timesheet: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"id":      timesheetID,
		"status":  "approved",
		"message": "Timesheet approved successfully",
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) rejectTimesheet(w http.ResponseWriter, r *http.Request) {
	timesheetID := chi.URLParam(r, "timesheetID")
	if timesheetID == "" {
		writeError(w, http.StatusBadRequest, "Timesheet ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	req := &crm.ReviewTimesheetRequest{
		Status:      crm.TimesheetStatusRejected,
		ReviewNotes: request.Reason,
	}

	err := h.crm.ReviewTimesheetWithValidation(r.Context(), user.ID, timesheetID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Timesheet not found")
		} else if strings.Contains(err.Error(), "not a leader") {
			writeError(w, http.StatusForbidden, "Only team leaders can reject timesheets")
		} else {
			writeError(w, http.StatusBadRequest, "Failed to reject timesheet: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"id":      timesheetID,
		"status":  "rejected",
		"reason":  request.Reason,
		"message": "Timesheet rejected",
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) listMyTimesheets(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	teamID := r.Context().Value(contextKey("teamID")).(string)
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	limit := getLimit(r)
	offset := getOffset(r)

	timesheets, total, err := h.crm.Store().ListUserTimesheets(r.Context(), user.ID, teamID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list timesheets: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"timesheets": timesheets,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) createTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	teamID := r.Context().Value(contextKey("teamID")).(string)
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}
	
	var req crm.CreateTimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	timesheet, err := h.crm.CreateTimesheetWithValidation(r.Context(), user.ID, teamID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to create timesheet: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, timesheet)
}

func (h *APIHandlers) getTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	timesheetID := chi.URLParam(r, "timesheetID")
	if timesheetID == "" {
		writeError(w, http.StatusBadRequest, "Timesheet ID is required")
		return
	}

	timesheet, err := h.crm.ValidateTimesheetAccess(r.Context(), user.ID, timesheetID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Timesheet not found")
		} else if strings.Contains(err.Error(), "cannot access") {
			writeError(w, http.StatusForbidden, "Access denied")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get timesheet: "+err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, timesheet)
}

func (h *APIHandlers) updateTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	timesheetID := chi.URLParam(r, "timesheetID")
	if timesheetID == "" {
		writeError(w, http.StatusBadRequest, "Timesheet ID is required")
		return
	}
	
	var req crm.UpdateTimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate access
	if _, err := h.crm.ValidateTimesheetAccess(r.Context(), user.ID, timesheetID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Timesheet not found")
		} else if strings.Contains(err.Error(), "cannot access") {
			writeError(w, http.StatusForbidden, "Access denied")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to access timesheet: "+err.Error())
		}
		return
	}

	timesheet, err := h.crm.Store().UpdateTimesheet(r.Context(), timesheetID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to update timesheet: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, timesheet)
}

func (h *APIHandlers) deleteTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	timesheetID := chi.URLParam(r, "timesheetID")
	if timesheetID == "" {
		writeError(w, http.StatusBadRequest, "Timesheet ID is required")
		return
	}

	// Validate access
	if _, err := h.crm.ValidateTimesheetAccess(r.Context(), user.ID, timesheetID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Timesheet not found")
		} else if strings.Contains(err.Error(), "cannot access") {
			writeError(w, http.StatusForbidden, "Access denied")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to access timesheet: "+err.Error())
		}
		return
	}

	err := h.crm.Store().DeleteTimesheet(r.Context(), timesheetID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to delete timesheet: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Submit timesheet for review
func (h *APIHandlers) submitTimesheet(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	timesheetID := chi.URLParam(r, "timesheetID")
	if timesheetID == "" {
		writeError(w, http.StatusBadRequest, "Timesheet ID is required")
		return
	}

	err := h.crm.SubmitTimesheetWithValidation(r.Context(), user.ID, timesheetID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Timesheet not found")
		} else if strings.Contains(err.Error(), "cannot submit") {
			writeError(w, http.StatusForbidden, "Cannot submit this timesheet")
		} else {
			writeError(w, http.StatusBadRequest, "Failed to submit timesheet: "+err.Error())
		}
		return
	}

	response := map[string]interface{}{
		"id":      timesheetID,
		"status":  "submitted",
		"message": "Timesheet submitted for review",
	}

	writeJSON(w, http.StatusOK, response)
}

// Roster Handlers

func (h *APIHandlers) listTeamRosters(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	limit := getLimit(r)
	offset := getOffset(r)

	rosters, total, err := h.crm.Store().ListTeamRosters(r.Context(), teamID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list rosters: "+err.Error())
		return
	}

	// Note: Shifts are loaded separately when getting individual rosters
	// to avoid N+1 query problems in list view

	response := map[string]interface{}{
		"rosters": rosters,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) getRoster(w http.ResponseWriter, r *http.Request) {
	rosterID := chi.URLParam(r, "rosterID")
	if rosterID == "" {
		writeError(w, http.StatusBadRequest, "Roster ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Validate access
	roster, err := h.crm.ValidateRosterAccess(r.Context(), user.ID, rosterID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Roster not found")
		} else if strings.Contains(err.Error(), "cannot access") {
			writeError(w, http.StatusForbidden, "Access denied")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get roster: "+err.Error())
		}
		return
	}

	// Get shifts for this roster
	shifts, err := h.crm.Store().ListRosterShifts(r.Context(), rosterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get roster shifts: "+err.Error())
		return
	}

	// Create response with shifts included
	response := map[string]interface{}{
		"roster": roster,
		"shifts": shifts,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) createRoster(w http.ResponseWriter, r *http.Request) {
	teamID := r.Context().Value(contextKey("teamID")).(string)
	if teamID == "" {
		writeError(w, http.StatusBadRequest, "Team ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	var req crm.CreateRosterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	roster, err := h.crm.CreateRosterWithValidation(r.Context(), teamID, &req, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not a leader") {
			writeError(w, http.StatusForbidden, "Only team leaders can create rosters")
		} else {
			writeError(w, http.StatusBadRequest, "Failed to create roster: "+err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, roster)
}

func (h *APIHandlers) updateRoster(w http.ResponseWriter, r *http.Request) {
	rosterID := chi.URLParam(r, "rosterID")
	if rosterID == "" {
		writeError(w, http.StatusBadRequest, "Roster ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	// Validate access first
	if _, err := h.crm.ValidateRosterAccess(r.Context(), user.ID, rosterID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Roster not found")
		} else if strings.Contains(err.Error(), "cannot access") {
			writeError(w, http.StatusForbidden, "Access denied")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to access roster: "+err.Error())
		}
		return
	}
	
	var req crm.UpdateRosterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	roster, err := h.crm.Store().UpdateRoster(r.Context(), rosterID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to update roster: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, roster)
}

func (h *APIHandlers) deleteRoster(w http.ResponseWriter, r *http.Request) {
	rosterID := chi.URLParam(r, "rosterID")
	if rosterID == "" {
		writeError(w, http.StatusBadRequest, "Roster ID is required")
		return
	}

	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Validate access first
	if _, err := h.crm.ValidateRosterAccess(r.Context(), user.ID, rosterID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "Roster not found")
		} else if strings.Contains(err.Error(), "cannot access") {
			writeError(w, http.StatusForbidden, "Access denied")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to access roster: "+err.Error())
		}
		return
	}

	err := h.crm.Store().DeleteRoster(r.Context(), rosterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete roster: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}