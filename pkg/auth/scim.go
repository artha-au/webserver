package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/artha-au/webserver/pkg/rbac"
	"github.com/go-chi/chi/v5"
)

type SCIMHandler struct {
	authService *AuthService
	rbacStore   rbac.Store
}

func NewSCIMHandler(authService *AuthService, rbacStore rbac.Store) *SCIMHandler {
	return &SCIMHandler{
		authService: authService,
		rbacStore:   rbacStore,
	}
}

// RegisterSCIMRoutes registers SCIM v2.0 endpoints with the router
func (h *SCIMHandler) RegisterSCIMRoutes(r chi.Router) {
	r.Route("/scim/v2", func(r chi.Router) {
		// Users endpoint
		r.Route("/Users", func(r chi.Router) {
			r.Get("/", h.listUsers)
			r.Post("/", h.createUser)
			r.Get("/{id}", h.getUser)
			r.Put("/{id}", h.updateUser)
			r.Patch("/{id}", h.patchUser)
			r.Delete("/{id}", h.deleteUser)
		})

		// Groups endpoint
		r.Route("/Groups", func(r chi.Router) {
			r.Get("/", h.listGroups)
			r.Post("/", h.createGroup)
			r.Get("/{id}", h.getGroup)
			r.Put("/{id}", h.updateGroup)
			r.Patch("/{id}", h.patchGroup)
			r.Delete("/{id}", h.deleteGroup)
		})

		// ServiceProviderConfig endpoint
		r.Get("/ServiceProviderConfig", h.getServiceProviderConfig)

		// ResourceTypes endpoint
		r.Get("/ResourceTypes", h.getResourceTypes)

		// Schemas endpoint
		r.Get("/Schemas", h.getSchemas)
	})
}

// User Operations

func (h *SCIMHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startIndex := 1
	count := 20
	filter := r.URL.Query().Get("filter")

	if start := r.URL.Query().Get("startIndex"); start != "" {
		if s, err := strconv.Atoi(start); err == nil && s > 0 {
			startIndex = s
		}
	}

	if c := r.URL.Query().Get("count"); c != "" {
		if cnt, err := strconv.Atoi(c); err == nil && cnt > 0 {
			count = cnt
		}
	}

	// For now, we'll get all users and filter/paginate in memory
	// In production, this should be done at the database level
	users, err := h.getAllSCIMUsers()
	if err != nil {
		h.sendSCIMError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Apply filter if provided
	if filter != "" {
		users = h.filterUsers(users, filter)
	}

	// Calculate pagination
	totalResults := len(users)
	end := startIndex + count - 1
	if end > totalResults {
		end = totalResults
	}

	if startIndex > totalResults {
		users = []SCIMUser{}
	} else {
		users = users[startIndex-1 : end]
	}

	response := SCIMListResponse{
		Schemas:      []string{SCIMSchemaList},
		TotalResults: totalResults,
		StartIndex:   startIndex,
		ItemsPerPage: len(users),
		Resources:    users,
	}

	w.Header().Set("Content-Type", "application/scim+json")
	json.NewEncoder(w).Encode(response)
}

func (h *SCIMHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var scimUser SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&scimUser); err != nil {
		h.sendSCIMError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if scimUser.UserName == "" {
		h.sendSCIMError(w, http.StatusBadRequest, "userName is required")
		return
	}

	// Extract email from emails array or use userName as email
	email := scimUser.UserName
	if len(scimUser.Emails) > 0 {
		email = scimUser.Emails[0].Value
	}

	// Create user
	now := time.Now()
	user := &rbac.User{
		ID:        generateID(),
		Email:     email,
		Name:      scimUser.DisplayName,
		Active:    scimUser.Active,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if user.Name == "" {
		user.Name = scimUser.UserName
	}

	// Create user with SCIM attributes
	if err := h.authService.store.CreateUserWithSCIMAttributes(user, &scimUser); err != nil {
		h.sendSCIMError(w, http.StatusConflict, "User already exists or creation failed")
		return
	}

	// Return created user
	createdUser, err := h.authService.store.GetUserWithSCIMAttributes(user.ID)
	if err != nil {
		h.sendSCIMError(w, http.StatusInternalServerError, "Failed to retrieve created user")
		return
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (h *SCIMHandler) getUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendSCIMError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	user, err := h.authService.store.GetUserWithSCIMAttributes(userID)
	if err != nil {
		h.sendSCIMError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/scim+json")
	json.NewEncoder(w).Encode(user)
}

func (h *SCIMHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendSCIMError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var scimUser SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&scimUser); err != nil {
		h.sendSCIMError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// TODO: Implement user update
	h.sendSCIMError(w, http.StatusNotImplemented, "User update not implemented")
}

func (h *SCIMHandler) patchUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendSCIMError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// TODO: Implement PATCH operations
	h.sendSCIMError(w, http.StatusNotImplemented, "User patch not implemented")
}

func (h *SCIMHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.sendSCIMError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.authService.store.DeleteUser(userID); err != nil {
		h.sendSCIMError(w, http.StatusNotFound, "User not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Group Operations (placeholders)

func (h *SCIMHandler) listGroups(w http.ResponseWriter, r *http.Request) {
	response := SCIMListResponse{
		Schemas:      []string{SCIMSchemaList},
		TotalResults: 0,
		StartIndex:   1,
		ItemsPerPage: 0,
		Resources:    []SCIMGroup{},
	}

	w.Header().Set("Content-Type", "application/scim+json")
	json.NewEncoder(w).Encode(response)
}

func (h *SCIMHandler) createGroup(w http.ResponseWriter, r *http.Request) {
	h.sendSCIMError(w, http.StatusNotImplemented, "Group creation not implemented")
}

func (h *SCIMHandler) getGroup(w http.ResponseWriter, r *http.Request) {
	h.sendSCIMError(w, http.StatusNotFound, "Group not found")
}

func (h *SCIMHandler) updateGroup(w http.ResponseWriter, r *http.Request) {
	h.sendSCIMError(w, http.StatusNotImplemented, "Group update not implemented")
}

func (h *SCIMHandler) patchGroup(w http.ResponseWriter, r *http.Request) {
	h.sendSCIMError(w, http.StatusNotImplemented, "Group patch not implemented")
}

func (h *SCIMHandler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	h.sendSCIMError(w, http.StatusNotFound, "Group not found")
}

// Configuration Endpoints

func (h *SCIMHandler) getServiceProviderConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		"patch": map[string]bool{
			"supported": false,
		},
		"bulk": map[string]interface{}{
			"supported":      false,
			"maxPayloadSize": 0,
			"maxOperations":  0,
		},
		"filter": map[string]interface{}{
			"supported":  true,
			"maxResults": 200,
		},
		"changePassword": map[string]bool{
			"supported": false,
		},
		"sort": map[string]bool{
			"supported": false,
		},
		"etag": map[string]bool{
			"supported": false,
		},
		"authenticationSchemes": []map[string]interface{}{
			{
				"type":             "httpbasic",
				"name":             "HTTP Basic",
				"description":      "Authentication scheme using HTTP Basic",
				"specUri":          "http://www.rfc-editor.org/info/rfc2617",
				"documentationUri": "http://example.com/help/httpBasic.html",
			},
		},
	}

	w.Header().Set("Content-Type", "application/scim+json")
	json.NewEncoder(w).Encode(config)
}

func (h *SCIMHandler) getResourceTypes(w http.ResponseWriter, r *http.Request) {
	resourceTypes := []map[string]interface{}{
		{
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"},
			"id":          "User",
			"name":        "User",
			"endpoint":    "/Users",
			"description": "User Account",
			"schema":      "urn:ietf:params:scim:schemas:core:2.0:User",
		},
		{
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"},
			"id":          "Group",
			"name":        "Group",
			"endpoint":    "/Groups",
			"description": "Group",
			"schema":      "urn:ietf:params:scim:schemas:core:2.0:Group",
		},
	}

	response := SCIMListResponse{
		Schemas:      []string{SCIMSchemaList},
		TotalResults: len(resourceTypes),
		StartIndex:   1,
		ItemsPerPage: len(resourceTypes),
		Resources:    resourceTypes,
	}

	w.Header().Set("Content-Type", "application/scim+json")
	json.NewEncoder(w).Encode(response)
}

func (h *SCIMHandler) getSchemas(w http.ResponseWriter, r *http.Request) {
	schemas := []map[string]interface{}{
		{
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"},
			"id":          SCIMSchemaUser,
			"name":        "User",
			"description": "User Account",
		},
		{
			"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"},
			"id":          SCIMSchemaGroup,
			"name":        "Group",
			"description": "Group",
		},
	}

	response := SCIMListResponse{
		Schemas:      []string{SCIMSchemaList},
		TotalResults: len(schemas),
		StartIndex:   1,
		ItemsPerPage: len(schemas),
		Resources:    schemas,
	}

	w.Header().Set("Content-Type", "application/scim+json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (h *SCIMHandler) sendSCIMError(w http.ResponseWriter, status int, detail string) {
	error := SCIMError{
		Schemas: []string{SCIMSchemaError},
		Status:  strconv.Itoa(status),
		Detail:  detail,
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)
}

func (h *SCIMHandler) getAllSCIMUsers() ([]SCIMUser, error) {
	// This is a placeholder - in production, you'd query the database with proper pagination
	// For now, we'll return an empty slice as the actual implementation would depend on
	// your specific database queries and user storage
	return []SCIMUser{}, nil
}

func (h *SCIMHandler) filterUsers(users []SCIMUser, filter string) []SCIMUser {
	// Simple filter implementation - in production, you'd want more sophisticated filtering
	filter = strings.ToLower(filter)
	var filtered []SCIMUser

	for _, user := range users {
		if strings.Contains(strings.ToLower(user.UserName), filter) ||
			strings.Contains(strings.ToLower(user.DisplayName), filter) {
			filtered = append(filtered, user)
		}
	}

	return filtered
}
