package rbac

import (
	"context"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
)

// Repository interfaces

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role Role) (string, error)

	// Update updates an existing role
	Update(ctx context.Context, id string, role Role) error

	// Delete deletes a role
	Delete(ctx context.Context, id string) error

	// FindByID retrieves a role by ID
	FindByID(ctx context.Context, id string) (Role, error)

	// FindByName retrieves a role by name
	FindByName(ctx context.Context, name string) (Role, error)

	// FindMany retrieves roles by filter
	FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]Role, error)

	// Count counts roles by filter
	Count(ctx context.Context, filter repository.Filter) (int64, error)

	// AddPermissionToRole adds a permission to a role
	AddPermissionToRole(ctx context.Context, roleID, permissionID string) error

	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
}

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	// Create creates a new permission
	Create(ctx context.Context, permission models.Permission) (string, error)

	// Update updates an existing permission
	Update(ctx context.Context, id string, permission models.Permission) error

	// Delete deletes a permission
	Delete(ctx context.Context, id string) error

	// FindByID retrieves a permission by ID
	FindByID(ctx context.Context, id string) (models.Permission, error)

	// FindByName retrieves a permission by name
	FindByName(ctx context.Context, name string) (models.Permission, error)

	// FindByResourceAction retrieves permissions by resource and action
	FindByResourceAction(ctx context.Context, resource, action string) ([]models.Permission, error)

	// FindMany retrieves permissions by filter
	FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]models.Permission, error)

	// Count counts permissions by filter
	Count(ctx context.Context, filter repository.Filter) (int64, error)
}

// Service interfaces

// RoleService defines the interface for role management operations
type RoleService interface {
	// GetRoles retrieves all roles, optionally filtered
	GetRoles(ctx context.Context, filter repository.Filter, pagination map[string]int) ([]Role, int, error)

	// GetRole retrieves a role by ID
	GetRole(ctx context.Context, id string) (*Role, error)

	// CreateRole creates a new role
	CreateRole(ctx context.Context, role Role) (*Role, error)

	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, id string, updates map[string]interface{}) (*Role, error)

	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, id string) error

	// AddPermissionsToRole adds permissions to a role
	AddPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string) error

	// RemovePermissionsFromRole removes permissions from a role
	RemovePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string) error

	// GetUserRoles retrieves all roles assigned to a user
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)

	// IsUserInRole checks if a user has a specific role
	IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error)
}

// PermissionService defines the interface for permission management operations
type PermissionService interface {
	// GetPermissions retrieves all permissions, optionally filtered
	GetPermissions(ctx context.Context, filter repository.Filter, pagination map[string]int) ([]models.Permission, int, error)

	// GetPermission retrieves a permission by ID
	GetPermission(ctx context.Context, id string) (*models.Permission, error)

	// CreatePermission creates a new permission
	CreatePermission(ctx context.Context, permission models.Permission) (*models.Permission, error)

	// UpdatePermission updates an existing permission
	UpdatePermission(ctx context.Context, id string, updates map[string]interface{}) (*models.Permission, error)

	// DeletePermission deletes a permission
	DeletePermission(ctx context.Context, id string) error

	// GetPermissionsByRole retrieves all permissions for a specific role
	GetPermissionsByRole(ctx context.Context, roleID string) ([]models.Permission, error)

	// GetPermissionsByUser retrieves all permissions for a specific user
	GetPermissionsByUser(ctx context.Context, userID string) ([]models.Permission, error)
}
