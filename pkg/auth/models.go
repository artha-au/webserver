package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthProvider represents an external authentication provider (e.g., SAML, OIDC)
type AuthProvider struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        string                 `json:"type" db:"type"` // "saml", "oidc", "ldap"
	Config      map[string]interface{} `json:"config" db:"config"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	NamespaceID *string                `json:"namespace_id" db:"namespace_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// UserSession represents an active user session
type UserSession struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	Token        string     `json:"token" db:"token"`
	RefreshToken *string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt   time.Time  `json:"last_used_at" db:"last_used_at"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	RevokedAt    *time.Time `json:"revoked_at" db:"revoked_at"`
}

// UserAuthProvider links users to their external authentication providers
type UserAuthProvider struct {
	ID               string                 `json:"id" db:"id"`
	UserID           string                 `json:"user_id" db:"user_id"`
	AuthProviderID   string                 `json:"auth_provider_id" db:"auth_provider_id"`
	ExternalUserID   string                 `json:"external_user_id" db:"external_user_id"`
	ExternalUsername string                 `json:"external_username" db:"external_username"`
	Attributes       map[string]interface{} `json:"attributes" db:"attributes"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// SCIMUser represents a user in SCIM format
type SCIMUser struct {
	ID                string            `json:"id"`
	ExternalID        string            `json:"externalId,omitempty"`
	UserName          string            `json:"userName"`
	Name              *SCIMName         `json:"name,omitempty"`
	DisplayName       string            `json:"displayName,omitempty"`
	NickName          string            `json:"nickName,omitempty"`
	ProfileURL        string            `json:"profileUrl,omitempty"`
	Title             string            `json:"title,omitempty"`
	UserType          string            `json:"userType,omitempty"`
	PreferredLanguage string            `json:"preferredLanguage,omitempty"`
	Locale            string            `json:"locale,omitempty"`
	Timezone          string            `json:"timezone,omitempty"`
	Active            bool              `json:"active"`
	Password          string            `json:"password,omitempty"`
	Emails            []SCIMEmail       `json:"emails,omitempty"`
	PhoneNumbers      []SCIMPhoneNumber `json:"phoneNumbers,omitempty"`
	Addresses         []SCIMAddress     `json:"addresses,omitempty"`
	Groups            []SCIMGroupRef    `json:"groups,omitempty"`
	Roles             []SCIMRole        `json:"roles,omitempty"`
	Meta              SCIMMetadata      `json:"meta"`
	Schemas           []string          `json:"schemas"`
}

type SCIMName struct {
	Formatted       string `json:"formatted,omitempty"`
	FamilyName      string `json:"familyName,omitempty"`
	GivenName       string `json:"givenName,omitempty"`
	MiddleName      string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type SCIMPhoneNumber struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type SCIMAddress struct {
	Type          string `json:"type,omitempty"`
	StreetAddress string `json:"streetAddress,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postalCode,omitempty"`
	Country       string `json:"country,omitempty"`
	Formatted     string `json:"formatted,omitempty"`
	Primary       bool   `json:"primary,omitempty"`
}

type SCIMGroupRef struct {
	Value   string `json:"value"`
	Ref     string `json:"$ref,omitempty"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
}

type SCIMRole struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type SCIMMetadata struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Location     string    `json:"location"`
	Version      string    `json:"version"`
}

// SCIMGroup represents a group in SCIM format
type SCIMGroup struct {
	ID          string         `json:"id"`
	ExternalID  string         `json:"externalId,omitempty"`
	DisplayName string         `json:"displayName"`
	Members     []SCIMGroupRef `json:"members,omitempty"`
	Meta        SCIMMetadata   `json:"meta"`
	Schemas     []string       `json:"schemas"`
}

// SCIMListResponse represents a SCIM list response
type SCIMListResponse struct {
	Schemas      []string    `json:"schemas"`
	TotalResults int         `json:"totalResults"`
	StartIndex   int         `json:"startIndex"`
	ItemsPerPage int         `json:"itemsPerPage"`
	Resources    interface{} `json:"Resources"`
}

// SCIMError represents a SCIM error response
type SCIMError struct {
	Schemas []string `json:"schemas"`
	Status  string   `json:"status"`
	Detail  string   `json:"detail"`
}

// Constants for SCIM schemas
const (
	SCIMSchemaUser  = "urn:ietf:params:scim:schemas:core:2.0:User"
	SCIMSchemaGroup = "urn:ietf:params:scim:schemas:core:2.0:Group"
	SCIMSchemaError = "urn:ietf:params:scim:api:messages:2.0:Error"
	SCIMSchemaList  = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
)

// AuthProvider types
const (
	AuthProviderSAML = "saml"
	AuthProviderOIDC = "oidc"
	AuthProviderLDAP = "ldap"
)

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	NamespaceID *string  `json:"namespace_id,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// Implement jwt.Claims interface methods
func (c TokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.ExpiresAt, nil
}

func (c TokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.IssuedAt, nil
}

func (c TokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.NotBefore, nil
}

func (c TokenClaims) GetIssuer() (string, error) {
	return c.RegisteredClaims.Issuer, nil
}

func (c TokenClaims) GetSubject() (string, error) {
	return c.RegisteredClaims.Subject, nil
}

func (c TokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.Audience, nil
}
