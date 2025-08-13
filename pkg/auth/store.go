package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/artha-au/webserver/pkg/rbac"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// AuthProvider operations
func (s *Store) CreateAuthProvider(provider *AuthProvider) error {
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO auth_providers (id, name, type, config, enabled, namespace_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.db.Exec(query, provider.ID, provider.Name, provider.Type, configJSON,
		provider.Enabled, provider.NamespaceID, provider.CreatedAt, provider.UpdatedAt)
	return err
}

func (s *Store) GetAuthProvider(id string) (*AuthProvider, error) {
	query := `
		SELECT id, name, type, config, enabled, namespace_id, created_at, updated_at
		FROM auth_providers WHERE id = $1
	`
	provider := &AuthProvider{}
	var configJSON []byte

	err := s.db.QueryRow(query, id).Scan(
		&provider.ID, &provider.Name, &provider.Type, &configJSON,
		&provider.Enabled, &provider.NamespaceID, &provider.CreatedAt, &provider.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return provider, nil
}

func (s *Store) ListAuthProviders(namespaceID *string) ([]*AuthProvider, error) {
	var query string
	var args []interface{}

	if namespaceID != nil {
		query = `
			SELECT id, name, type, config, enabled, namespace_id, created_at, updated_at
			FROM auth_providers WHERE namespace_id = $1 ORDER BY name
		`
		args = []interface{}{*namespaceID}
	} else {
		query = `
			SELECT id, name, type, config, enabled, namespace_id, created_at, updated_at
			FROM auth_providers WHERE namespace_id IS NULL ORDER BY name
		`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*AuthProvider
	for rows.Next() {
		provider := &AuthProvider{}
		var configJSON []byte

		err := rows.Scan(&provider.ID, &provider.Name, &provider.Type, &configJSON,
			&provider.Enabled, &provider.NamespaceID, &provider.CreatedAt, &provider.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &provider.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		providers = append(providers, provider)
	}

	return providers, rows.Err()
}

func (s *Store) UpdateAuthProvider(provider *AuthProvider) error {
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		UPDATE auth_providers 
		SET name = $2, type = $3, config = $4, enabled = $5, namespace_id = $6, updated_at = $7
		WHERE id = $1
	`
	_, err = s.db.Exec(query, provider.ID, provider.Name, provider.Type, configJSON,
		provider.Enabled, provider.NamespaceID, provider.UpdatedAt)
	return err
}

func (s *Store) DeleteAuthProvider(id string) error {
	query := `DELETE FROM auth_providers WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// UserSession operations
func (s *Store) CreateUserSession(session *UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, token, refresh_token, expires_at, created_at, last_used_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.Exec(query, session.ID, session.UserID, session.Token, session.RefreshToken,
		session.ExpiresAt, session.CreatedAt, session.LastUsedAt, session.IPAddress, session.UserAgent)
	return err
}

func (s *Store) GetUserSessionByToken(token string) (*UserSession, error) {
	query := `
		SELECT id, user_id, token, refresh_token, expires_at, created_at, last_used_at, ip_address, user_agent, revoked_at
		FROM user_sessions WHERE token = $1 AND revoked_at IS NULL
	`
	session := &UserSession{}

	err := s.db.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.ExpiresAt, &session.CreatedAt, &session.LastUsedAt,
		&session.IPAddress, &session.UserAgent, &session.RevokedAt)

	return session, err
}

func (s *Store) UpdateUserSessionLastUsed(token string, lastUsed time.Time) error {
	query := `UPDATE user_sessions SET last_used_at = $2 WHERE token = $1`
	_, err := s.db.Exec(query, token, lastUsed)
	return err
}

func (s *Store) RevokeUserSession(token string) error {
	query := `UPDATE user_sessions SET revoked_at = CURRENT_TIMESTAMP WHERE token = $1`
	_, err := s.db.Exec(query, token)
	return err
}

func (s *Store) RevokeAllUserSessions(userID string) error {
	query := `UPDATE user_sessions SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := s.db.Exec(query, userID)
	return err
}

func (s *Store) CleanupExpiredSessions() error {
	query := `DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP OR revoked_at IS NOT NULL`
	_, err := s.db.Exec(query)
	return err
}

// UserAuthProvider operations
func (s *Store) CreateUserAuthProvider(uap *UserAuthProvider) error {
	attributesJSON, err := json.Marshal(uap.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	query := `
		INSERT INTO user_auth_providers (id, user_id, auth_provider_id, external_user_id, external_username, attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.db.Exec(query, uap.ID, uap.UserID, uap.AuthProviderID, uap.ExternalUserID,
		uap.ExternalUsername, attributesJSON, uap.CreatedAt, uap.UpdatedAt)
	return err
}

func (s *Store) GetUserAuthProvider(userID, authProviderID string) (*UserAuthProvider, error) {
	query := `
		SELECT id, user_id, auth_provider_id, external_user_id, external_username, attributes, created_at, updated_at
		FROM user_auth_providers WHERE user_id = $1 AND auth_provider_id = $2
	`
	uap := &UserAuthProvider{}
	var attributesJSON []byte

	err := s.db.QueryRow(query, userID, authProviderID).Scan(
		&uap.ID, &uap.UserID, &uap.AuthProviderID, &uap.ExternalUserID,
		&uap.ExternalUsername, &attributesJSON, &uap.CreatedAt, &uap.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributesJSON, &uap.Attributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	return uap, nil
}

func (s *Store) GetUserByExternalID(authProviderID, externalUserID string) (*rbac.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.active, u.created_at, u.updated_at
		FROM users u
		JOIN user_auth_providers uap ON u.id = uap.user_id
		WHERE uap.auth_provider_id = $1 AND uap.external_user_id = $2
	`
	user := &rbac.User{}

	err := s.db.QueryRow(query, authProviderID, externalUserID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Active, &user.CreatedAt, &user.UpdatedAt)

	return user, err
}

func (s *Store) UpdateUserAuthProvider(uap *UserAuthProvider) error {
	attributesJSON, err := json.Marshal(uap.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	query := `
		UPDATE user_auth_providers 
		SET external_user_id = $3, external_username = $4, attributes = $5, updated_at = $6
		WHERE user_id = $1 AND auth_provider_id = $2
	`
	_, err = s.db.Exec(query, uap.UserID, uap.AuthProviderID, uap.ExternalUserID,
		uap.ExternalUsername, attributesJSON, uap.UpdatedAt)
	return err
}

func (s *Store) DeleteUserAuthProvider(userID, authProviderID string) error {
	query := `DELETE FROM user_auth_providers WHERE user_id = $1 AND auth_provider_id = $2`
	_, err := s.db.Exec(query, userID, authProviderID)
	return err
}

// Enhanced User operations for SCIM
func (s *Store) CreateUserWithSCIMAttributes(user *rbac.User, scimUser *SCIMUser) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create basic user
	userQuery := `
		INSERT INTO users (id, email, name, active, created_at, updated_at, external_id, user_name, display_name, nick_name, profile_url, title, user_type, preferred_language, locale, timezone, scim_attributes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	scimAttributesJSON, err := json.Marshal(map[string]interface{}{
		"emails":       scimUser.Emails,
		"phoneNumbers": scimUser.PhoneNumbers,
		"addresses":    scimUser.Addresses,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal SCIM attributes: %w", err)
	}

	_, err = tx.Exec(userQuery, user.ID, user.Email, user.Name, user.Active, user.CreatedAt, user.UpdatedAt,
		scimUser.ExternalID, scimUser.UserName, scimUser.DisplayName, scimUser.NickName, scimUser.ProfileURL,
		scimUser.Title, scimUser.UserType, scimUser.PreferredLanguage, scimUser.Locale, scimUser.Timezone, scimAttributesJSON)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) GetUserWithSCIMAttributes(userID string) (*SCIMUser, error) {
	query := `
		SELECT id, email, name, active, created_at, updated_at, external_id, user_name, display_name, nick_name, profile_url, title, user_type, preferred_language, locale, timezone, scim_attributes
		FROM users WHERE id = $1
	`

	var id, email, name, externalID, userName, displayName, nickName, profileURL, title, userType, preferredLanguage, locale, timezone sql.NullString
	var active bool
	var createdAt, updatedAt time.Time
	var scimAttributesJSON []byte

	err := s.db.QueryRow(query, userID).Scan(
		&id, &email, &name, &active, &createdAt, &updatedAt,
		&externalID, &userName, &displayName, &nickName, &profileURL,
		&title, &userType, &preferredLanguage, &locale, &timezone, &scimAttributesJSON)
	if err != nil {
		return nil, err
	}

	scimUser := &SCIMUser{
		ID:                id.String,
		Active:            active,
		ExternalID:        externalID.String,
		UserName:          userName.String,
		DisplayName:       displayName.String,
		NickName:          nickName.String,
		ProfileURL:        profileURL.String,
		Title:             title.String,
		UserType:          userType.String,
		PreferredLanguage: preferredLanguage.String,
		Locale:            locale.String,
		Timezone:          timezone.String,
		Meta: SCIMMetadata{
			ResourceType: "User",
			Created:      createdAt,
			LastModified: updatedAt,
			Location:     fmt.Sprintf("/scim/v2/Users/%s", id.String),
			Version:      "1",
		},
		Schemas: []string{SCIMSchemaUser},
	}

	// Parse SCIM attributes
	var scimAttrs map[string]interface{}
	if err := json.Unmarshal(scimAttributesJSON, &scimAttrs); err == nil {
		if emails, ok := scimAttrs["emails"]; ok {
			emailsJSON, _ := json.Marshal(emails)
			json.Unmarshal(emailsJSON, &scimUser.Emails)
		}
		if phoneNumbers, ok := scimAttrs["phoneNumbers"]; ok {
			phoneJSON, _ := json.Marshal(phoneNumbers)
			json.Unmarshal(phoneJSON, &scimUser.PhoneNumbers)
		}
		if addresses, ok := scimAttrs["addresses"]; ok {
			addrJSON, _ := json.Marshal(addresses)
			json.Unmarshal(addrJSON, &scimUser.Addresses)
		}
	}

	// Add primary email if not in SCIM emails
	if email.String != "" && len(scimUser.Emails) == 0 {
		scimUser.Emails = []SCIMEmail{
			{Value: email.String, Type: "work", Primary: true},
		}
	}

	return scimUser, nil
}

// GetUserByEmail retrieves a user by email address
func (s *Store) GetUserByEmail(email string) (*rbac.User, error) {
	query := `
		SELECT id, email, name, active, created_at, updated_at
		FROM users WHERE email = $1
	`
	user := &rbac.User{}

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Active, &user.CreatedAt, &user.UpdatedAt)

	return user, err
}

// GetUserByEmailWithPassword retrieves a user by email address including password fields
func (s *Store) GetUserByEmailWithPassword(email string) (*rbac.User, string, string, error) {
	query := `
		SELECT id, email, name, active, created_at, updated_at, password_hash, password_salt
		FROM users WHERE email = $1
	`
	user := &rbac.User{}
	var passwordHash, passwordSalt sql.NullString

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Active, &user.CreatedAt, &user.UpdatedAt,
		&passwordHash, &passwordSalt)

	if err != nil {
		return nil, "", "", err
	}

	return user, passwordHash.String, passwordSalt.String, nil
}

// CreateUser creates a new user (wrapper around rbac operations)
func (s *Store) CreateUser(user *rbac.User) error {
	query := `
		INSERT INTO users (id, email, name, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.Exec(query, user.ID, user.Email, user.Name, user.Active, user.CreatedAt, user.UpdatedAt)
	return err
}

// DeleteUser deletes a user (wrapper around rbac operations)
func (s *Store) DeleteUser(userID string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.db.Exec(query, userID)
	return err
}

// SetUserPassword sets the password hash and salt for a user
func (s *Store) SetUserPassword(userID, passwordHash, passwordSalt string) error {
	query := `UPDATE users SET password_hash = $2, password_salt = $3 WHERE id = $1`
	_, err := s.db.Exec(query, userID, passwordHash, passwordSalt)
	return err
}
