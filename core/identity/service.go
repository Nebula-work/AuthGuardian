package identity

import (
	"context"
)

// UserService defines the interface for user operations
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user User) (string, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id string, user User) error

	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id string) error

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id string) (User, error)

	// GetUsers retrieves all users with optional filtering
	GetUsers(ctx context.Context, orgID string, skip, limit int64) ([]User, int64, error)

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (User, error)

	// GetUserByOAuthID retrieves a user by OAuth provider and ID
	GetUserByOAuthID(ctx context.Context, provider, providerUserID string) (User, error)

	// UpdateLastLogin updates a user's last login time
	UpdateLastLogin(ctx context.Context, id string) error

	// SetPassword sets a user's password
	SetPassword(ctx context.Context, id, password string) error

	// VerifyPassword verifies a user's password
	VerifyPassword(ctx context.Context, id, password string) (bool, error)

	// VerifyEmail marks a user's email as verified
	VerifyEmail(ctx context.Context, id string) error

	// AddRoleToUser adds a role to a user
	AddRoleToUser(ctx context.Context, userID, roleID string) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error

	// AddOrganizationToUser adds an organization to a user
	AddOrganizationToUser(ctx context.Context, userID, organizationID string) error

	// RemoveOrganizationFromUser removes an organization from a user
	RemoveOrganizationFromUser(ctx context.Context, userID, organizationID string) error

	// GetUserRoles retrieves roles for a user
	GetUserRoles(ctx context.Context, userID string) ([]string, error)

	// GetUserOrganizations retrieves organizations for a user
	GetUserOrganizations(ctx context.Context, userID string) ([]string, error)

	// LinkOAuthAccount links an OAuth account to a user
	LinkOAuthAccount(ctx context.Context, userID, provider, providerUserID, refreshToken string) error

	// UnlinkOAuthAccount unlinks an OAuth account from a user
	UnlinkOAuthAccount(ctx context.Context, userID, provider string) error
}

// OrganizationService defines the interface for organization operations
type OrganizationService interface {
	// CreateOrganization creates a new organization
	CreateOrganization(ctx context.Context, organization Organization) (string, error)

	// UpdateOrganization updates an existing organization
	UpdateOrganization(ctx context.Context, id string, organization Organization) error

	// DeleteOrganization deletes an organization
	DeleteOrganization(ctx context.Context, id string) error

	// GetOrganization retrieves an organization by ID
	GetOrganization(ctx context.Context, id string) (Organization, error)

	// GetOrganizations retrieves all organizations with optional filtering
	GetOrganizations(ctx context.Context, skip, limit int64) ([]Organization, int64, error)

	// GetOrganizationByName retrieves an organization by name
	GetOrganizationByName(ctx context.Context, name string) (Organization, error)

	// GetOrganizationByDomain retrieves an organization by domain
	GetOrganizationByDomain(ctx context.Context, domain string) (Organization, error)

	// GetUserOrganizations retrieves organizations for a user
	GetUserOrganizations(ctx context.Context, userID string) ([]Organization, error)

	// AddUserToOrganization adds a user to an organization
	AddUserToOrganization(ctx context.Context, orgID, userID string) error

	// RemoveUserFromOrganization removes a user from an organization
	RemoveUserFromOrganization(ctx context.Context, orgID, userID string) error

	// AddAdminToOrganization adds an admin to an organization
	AddAdminToOrganization(ctx context.Context, orgID, userID string) error

	// RemoveAdminFromOrganization removes an admin from an organization
	RemoveAdminFromOrganization(ctx context.Context, orgID, userID string) error

	// IsUserAdmin checks if a user is an admin of an organization
	IsUserAdmin(ctx context.Context, orgID, userID string) (bool, error)

	// GetOrganizationUsers retrieves users in an organization
	GetOrganizationUsers(ctx context.Context, orgID string, skip, limit int64) ([]User, int64, error)

	// GetOrganizationAdmins retrieves admins of an organization
	GetOrganizationAdmins(ctx context.Context, orgID string) ([]User, error)
}
