package identity

import (
	"context"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"time"
)

// Internal models with database specific fields
// These extend the base models for database operations

// User extends models.User with database specific fields
type User struct {
	models.User
	OAuthAccounts []OAuthAccount `json:"-" bson:"oauthAccounts"`
}

// OAuthAccount represents an OAuth account linked to a user
type OAuthAccount struct {
	Provider       string    `json:"provider" bson:"provider"`
	ProviderUserID string    `json:"providerUserId" bson:"providerUserId"`
	RefreshToken   string    `json:"-" bson:"refreshToken"`
	LinkedAt       time.Time `json:"linkedAt" bson:"linkedAt"`
}

// Organization extends models.Organization with database specific fields
type Organization struct {
	models.Organization
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// IsConnected checks if the repository is connected to the database
	IsConnected(ctx context.Context) bool

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id string) (User, error)

	// FindOne finds a single user matching the filter
	FindOne(ctx context.Context, filter repository.Filter) (User, error)

	// FindMany finds multiple users matching the filter
	FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]User, error)

	// Count counts users matching the filter
	Count(ctx context.Context, filter repository.Filter) (int64, error)

	// Create creates a new user
	Create(ctx context.Context, user User) (string, error)

	// Update updates an existing user
	Update(ctx context.Context, id string, user User) error

	// Delete deletes a user
	Delete(ctx context.Context, id string) error

	// FindByUsername finds a user by username
	FindByUsername(ctx context.Context, username string) (User, error)

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (User, error)

	// FindByOAuthID finds a user by OAuth provider and ID
	FindByOAuthID(ctx context.Context, provider, providerUserID string) (User, error)

	// UpdateLastLogin updates a user's last login time
	UpdateLastLogin(ctx context.Context, id string) error

	// AddRoleToUser adds a role to a user
	AddRoleToUser(ctx context.Context, userID, roleID string) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error

	// AddOrganizationToUser adds an organization to a user
	AddOrganizationToUser(ctx context.Context, userID, organizationID string) error

	// RemoveOrganizationFromUser removes an organization from a user
	RemoveOrganizationFromUser(ctx context.Context, userID, organizationID string) error

	// FindByRoleID finds users with a specific role
	FindByRoleID(ctx context.Context, roleID string) ([]User, error)

	// FindByOrganizationID finds users in a specific organization
	FindByOrganizationID(ctx context.Context, organizationID string) ([]User, error)

	// LinkOAuthAccount links an OAuth account to a user
	LinkOAuthAccount(ctx context.Context, userID, provider, providerUserID, refreshToken string) error

	// UnlinkOAuthAccount unlinks an OAuth account from a user
	UnlinkOAuthAccount(ctx context.Context, userID, provider string) error
}

// OrganizationRepository defines the interface for organization persistence
type OrganizationRepository interface {
	// IsConnected checks if the repository is connected to the database
	IsConnected(ctx context.Context) bool

	// FindByID finds an organization by ID
	FindByID(ctx context.Context, id string) (Organization, error)

	// FindOne finds a single organization matching the filter
	FindOne(ctx context.Context, filter repository.Filter) (Organization, error)

	// FindMany finds multiple organizations matching the filter
	FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]Organization, error)

	// Count counts organizations matching the filter
	Count(ctx context.Context, filter repository.Filter) (int64, error)

	// Create creates a new organization
	Create(ctx context.Context, org Organization) (string, error)

	// Update updates an existing organization
	Update(ctx context.Context, id string, org Organization) error

	// Delete deletes an organization
	Delete(ctx context.Context, id string) error

	// FindByName finds an organization by name
	FindByName(ctx context.Context, name string) (Organization, error)

	// FindByDomain finds an organization by domain
	FindByDomain(ctx context.Context, domain string) (Organization, error)

	// AddAdminToOrganization adds an admin to an organization
	AddAdminToOrganization(ctx context.Context, orgID, userID string) error

	// RemoveAdminFromOrganization removes an admin from an organization
	RemoveAdminFromOrganization(ctx context.Context, orgID, userID string) error

	// IsUserAdmin checks if a user is an admin of an organization
	IsUserAdmin(ctx context.Context, orgID, userID string) (bool, error)
}
