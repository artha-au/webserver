package auth

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type SSOHandler struct {
	authService *AuthService
}

func NewSSOHandler(authService *AuthService) *SSOHandler {
	return &SSOHandler{
		authService: authService,
	}
}

// RegisterSSORoutes registers SSO endpoints with the router
func (h *SSOHandler) RegisterSSORoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		// OAuth/OIDC endpoints
		r.Get("/login/{provider}", h.initiateLogin)
		r.Get("/callback/{provider}", h.handleCallback)
		r.Post("/token", h.tokenEndpoint)
		r.Post("/refresh", h.refreshToken)
		r.Post("/logout", h.logout)

		// SAML endpoints
		r.Get("/saml/{provider}/metadata", h.samlMetadata)
		r.Post("/saml/{provider}/acs", h.samlACS)
		r.Get("/saml/{provider}/sso", h.samlSSO)

		// User info endpoint
		r.Get("/userinfo", h.userInfo)

		// Provider management (admin endpoints)
		r.Route("/providers", func(r chi.Router) {
			r.Get("/", h.listProviders)
			r.Post("/", h.createProvider)
			r.Get("/{id}", h.getProvider)
			r.Put("/{id}", h.updateProvider)
			r.Delete("/{id}", h.deleteProvider)
		})
	})
}

// OAuth/OIDC Endpoints

func (h *SSOHandler) initiateLogin(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")

	provider, err := h.authService.GetAuthProvider(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	if !provider.Enabled {
		http.Error(w, "Provider not enabled", http.StatusForbidden)
		return
	}

	switch provider.Type {
	case AuthProviderOIDC:
		h.initiateOIDCLogin(w, r, provider, redirectURI, state)
	case AuthProviderSAML:
		h.initiateSAMLLogin(w, r, provider, redirectURI, state)
	default:
		http.Error(w, "Unsupported provider type", http.StatusBadRequest)
	}
}

func (h *SSOHandler) initiateOIDCLogin(w http.ResponseWriter, r *http.Request, provider *AuthProvider, redirectURI, state string) {
	// Extract OIDC configuration
	authURL, _ := provider.Config["auth_url"].(string)
	clientID, _ := provider.Config["client_id"].(string)
	scope, _ := provider.Config["scope"].(string)

	if authURL == "" || clientID == "" {
		http.Error(w, "Provider configuration incomplete", http.StatusInternalServerError)
		return
	}

	if scope == "" {
		scope = "openid profile email"
	}

	// Build authorization URL
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {clientID},
		"scope":         {scope},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}

	authorizationURL := authURL + "?" + params.Encode()
	http.Redirect(w, r, authorizationURL, http.StatusFound)
}

func (h *SSOHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	provider, err := h.authService.GetAuthProvider(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	switch provider.Type {
	case AuthProviderOIDC:
		h.handleOIDCCallback(w, r, provider, code, state)
	case AuthProviderSAML:
		http.Error(w, "SAML callback not supported via GET", http.StatusBadRequest)
	default:
		http.Error(w, "Unsupported provider type", http.StatusBadRequest)
	}
}

func (h *SSOHandler) handleOIDCCallback(w http.ResponseWriter, r *http.Request, provider *AuthProvider, code, state string) {
	// This is a simplified implementation
	// In production, you would:
	// 1. Exchange code for access token
	// 2. Fetch user info from the OIDC provider
	// 3. Create or update the user
	// 4. Generate your own JWT token

	// For now, we'll create a mock user
	userAttributes := map[string]interface{}{
		"email":    "user@example.com",
		"name":     "OIDC User",
		"username": "oidc_user",
	}

	user, err := h.authService.AuthenticateExternal(provider.ID, "external_user_id", userAttributes)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, refreshToken, err := h.authService.GenerateToken(user, nil)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Return tokens as JSON
	response := map[string]interface{}{
		"access_token":  token,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600, // 1 hour
		"state":         state,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SSOHandler) tokenEndpoint(w http.ResponseWriter, r *http.Request) {
	// OAuth2 token endpoint for client credentials or authorization code flow
	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r)
	case "client_credentials":
		h.handleClientCredentialsGrant(w, r)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, "Unsupported grant type", http.StatusBadRequest)
	}
}

func (h *SSOHandler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle OAuth2 authorization code flow
	http.Error(w, "Authorization code grant not implemented", http.StatusNotImplemented)
}

func (h *SSOHandler) handleClientCredentialsGrant(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle OAuth2 client credentials flow
	http.Error(w, "Client credentials grant not implemented", http.StatusNotImplemented)
}

func (h *SSOHandler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshTokenStr := r.FormValue("refresh_token")
	if refreshTokenStr == "" {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	newToken, newRefreshToken, err := h.authService.RefreshToken(refreshTokenStr)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"access_token":  newToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SSOHandler) refreshToken(w http.ResponseWriter, r *http.Request) {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	newToken, newRefreshToken, err := h.authService.RefreshToken(request.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{}{
		"access_token":  newToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SSOHandler) logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	// Extract token from "Bearer <token>"
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[len(bearerPrefix):]
	if err := h.authService.RevokeToken(token); err != nil {
		http.Error(w, "Token revocation failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SSOHandler) userInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := authHeader[len(bearerPrefix):]
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userInfo := map[string]interface{}{
		"sub":   claims.UserID,
		"email": claims.Email,
		"name":  claims.Name,
		"roles": claims.Roles,
	}

	if claims.NamespaceID != nil {
		userInfo["namespace_id"] = *claims.NamespaceID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// SAML Endpoints

func (h *SSOHandler) initiateSAMLLogin(w http.ResponseWriter, r *http.Request, provider *AuthProvider, redirectURI, state string) {
	// SAML SP-initiated SSO
	http.Error(w, "SAML SSO not implemented", http.StatusNotImplemented)
}

func (h *SSOHandler) samlMetadata(w http.ResponseWriter, r *http.Request) {
	// Return SAML SP metadata
	http.Error(w, "SAML metadata not implemented", http.StatusNotImplemented)
}

func (h *SSOHandler) samlACS(w http.ResponseWriter, r *http.Request) {
	// SAML Assertion Consumer Service
	http.Error(w, "SAML ACS not implemented", http.StatusNotImplemented)
}

func (h *SSOHandler) samlSSO(w http.ResponseWriter, r *http.Request) {
	// SAML SSO endpoint
	http.Error(w, "SAML SSO not implemented", http.StatusNotImplemented)
}

// Provider Management Endpoints

func (h *SSOHandler) listProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.authService.ListAuthProviders(nil)
	if err != nil {
		http.Error(w, "Failed to list providers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (h *SSOHandler) createProvider(w http.ResponseWriter, r *http.Request) {
	var provider AuthProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	provider.ID = generateID()
	if err := h.authService.CreateAuthProvider(&provider); err != nil {
		http.Error(w, "Failed to create provider", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(provider)
}

func (h *SSOHandler) getProvider(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	provider, err := h.authService.GetAuthProvider(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

func (h *SSOHandler) updateProvider(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	var provider AuthProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	provider.ID = providerID
	if err := h.authService.UpdateAuthProvider(&provider); err != nil {
		http.Error(w, "Failed to update provider", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

func (h *SSOHandler) deleteProvider(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	if err := h.authService.DeleteAuthProvider(providerID); err != nil {
		http.Error(w, "Failed to delete provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
