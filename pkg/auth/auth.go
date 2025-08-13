package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/artha-au/webserver/pkg/rbac"
)

type AuthService struct {
	store      *Store
	rbacStore  rbac.Store
	jwtSecret  []byte
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

type AuthConfig struct {
	JWTSecret      string
	TokenTTL       time.Duration
	RefreshTokenTTL time.Duration
}

func NewAuthService(db *sql.DB, config *AuthConfig) *AuthService {
	return &AuthService{
		store:      NewStore(db),
		rbacStore:  rbac.NewSQLStore(db),
		jwtSecret:  []byte(config.JWTSecret),
		tokenTTL:   config.TokenTTL,
		refreshTTL: config.RefreshTokenTTL,
	}
}

// Password hashing using Argon2
type passwordParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultPasswordParams = &passwordParams{
	memory:      64 * 1024,
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

func (a *AuthService) HashPassword(password string) (string, string, error) {
	salt := make([]byte, defaultPasswordParams.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", "", err
	}

	hash := argon2.IDKey([]byte(password), salt, defaultPasswordParams.iterations,
		defaultPasswordParams.memory, defaultPasswordParams.parallelism, defaultPasswordParams.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return b64Hash, b64Salt, nil
}

func (a *AuthService) VerifyPassword(password, hash, salt string) (bool, error) {
	saltBytes, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return false, err
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), saltBytes, defaultPasswordParams.iterations,
		defaultPasswordParams.memory, defaultPasswordParams.parallelism, defaultPasswordParams.keyLength)

	return subtle.ConstantTimeCompare(hashBytes, computedHash) == 1, nil
}

// JWT token operations
func (a *AuthService) GenerateToken(user *rbac.User, namespaceID *string) (string, string, error) {
	ctx := context.Background()
	// Get user roles for the token
	roles, err := a.rbacStore.GetUserRoles(ctx, user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.RoleID // We'll need to look up role names separately
	}

	// Create JWT claims
	now := time.Now()
	claims := TokenClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Name:        user.Name,
		NamespaceID: namespaceID,
		Roles:       roleNames,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.tokenTTL)),
			Issuer:    "artha-auth",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, err := a.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Store session
	session := &UserSession{
		ID:           generateID(),
		UserID:       user.ID,
		Token:        tokenString,
		RefreshToken: &refreshToken,
		ExpiresAt:    now.Add(a.refreshTTL),
		CreatedAt:    now,
		LastUsedAt:   now,
	}

	if err := a.store.CreateUserSession(session); err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}

	return tokenString, refreshToken, nil
}

func (a *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("token expired")
		}

		// Update session last used time
		if err := a.store.UpdateUserSessionLastUsed(tokenString, time.Now()); err != nil {
			// Log error but don't fail validation
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (a *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	session, err := a.store.GetUserSessionByToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if session.ExpiresAt.Before(time.Now()) {
		return "", "", fmt.Errorf("refresh token expired")
	}

	// Get user
	ctx := context.Background()
	user, err := a.rbacStore.GetUser(ctx, session.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	if !user.Active {
		return "", "", fmt.Errorf("user account is disabled")
	}

	// Revoke old session
	if err := a.store.RevokeUserSession(refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to revoke old session: %w", err)
	}

	// Generate new tokens
	return a.GenerateToken(user, nil)
}

func (a *AuthService) RevokeToken(token string) error {
	return a.store.RevokeUserSession(token)
}

func (a *AuthService) RevokeAllUserTokens(userID string) error {
	return a.store.RevokeAllUserSessions(userID)
}

// Authentication methods
func (a *AuthService) AuthenticateLocal(email, password string) (*rbac.User, error) {
	user, err := a.store.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	if !user.Active {
		return nil, fmt.Errorf("account is disabled")
	}

	// For local authentication, we need password fields in the user table
	// This would require extending the rbac.User struct or creating a separate query
	// For now, return an error indicating local auth setup is needed
	return nil, fmt.Errorf("local authentication not yet implemented - requires password fields")
}

func (a *AuthService) AuthenticateExternal(authProviderID, externalUserID string, attributes map[string]interface{}) (*rbac.User, error) {
	// Try to find existing user
	user, err := a.store.GetUserByExternalID(authProviderID, externalUserID)
	if err == sql.ErrNoRows {
		// Create new user from external provider attributes
		return a.createUserFromExternalProvider(authProviderID, externalUserID, attributes)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}

	if !user.Active {
		return nil, fmt.Errorf("account is disabled")
	}

	// Update user auth provider attributes
	uap, err := a.store.GetUserAuthProvider(user.ID, authProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user auth provider: %w", err)
	}

	uap.Attributes = attributes
	uap.UpdatedAt = time.Now()
	
	if err := a.store.UpdateUserAuthProvider(uap); err != nil {
		// Log error but don't fail authentication
	}

	return user, nil
}

func (a *AuthService) createUserFromExternalProvider(authProviderID, externalUserID string, attributes map[string]interface{}) (*rbac.User, error) {
	// Extract user information from attributes
	email, _ := attributes["email"].(string)
	name, _ := attributes["name"].(string)
	username, _ := attributes["username"].(string)
	
	if email == "" {
		return nil, fmt.Errorf("email is required for user creation")
	}

	if name == "" {
		name = email
	}

	if username == "" {
		username = email
	}

	// Create user
	now := time.Now()
	user := &rbac.User{
		ID:        generateID(),
		Email:     email,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := a.store.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Link to auth provider
	uap := &UserAuthProvider{
		ID:               generateID(),
		UserID:           user.ID,
		AuthProviderID:   authProviderID,
		ExternalUserID:   externalUserID,
		ExternalUsername: username,
		Attributes:       attributes,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := a.store.CreateUserAuthProvider(uap); err != nil {
		// Try to clean up the user
		a.store.DeleteUser(user.ID)
		return nil, fmt.Errorf("failed to link auth provider: %w", err)
	}

	return user, nil
}

// Utility functions
func (a *AuthService) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// Provider management
func (a *AuthService) CreateAuthProvider(provider *AuthProvider) error {
	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now
	return a.store.CreateAuthProvider(provider)
}

func (a *AuthService) GetAuthProvider(id string) (*AuthProvider, error) {
	return a.store.GetAuthProvider(id)
}

func (a *AuthService) ListAuthProviders(namespaceID *string) ([]*AuthProvider, error) {
	return a.store.ListAuthProviders(namespaceID)
}

func (a *AuthService) UpdateAuthProvider(provider *AuthProvider) error {
	provider.UpdatedAt = time.Now()
	return a.store.UpdateAuthProvider(provider)
}

func (a *AuthService) DeleteAuthProvider(id string) error {
	return a.store.DeleteAuthProvider(id)
}

// Session management
func (a *AuthService) CleanupExpiredSessions() error {
	return a.store.CleanupExpiredSessions()
}