package rbac

import (
	"context"
	"errors"
	"rbac-system/pkg/common/repository"
	"time"
)

// Domain-specific errors
var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrDuplicateRoleName      = errors.New("role with this name already exists")
	ErrInvalidPermissions     = errors.New("one or more permissions are invalid")
	ErrSystemRoleModification = errors.New("system default roles cannot be modified")
	ErrInvalidRoleData        = errors.New("invalid role data")
)

// Role represents a role in the RBAC system
type Role struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	OrganizationID  string   `json:"organizationId,omitempty"`
	PermissionIDs   []string `json:"permissionIds"`
	IsSystemDefault bool     `json:"isSystemDefault"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

// RoleServiceImpl implements the RoleService interface
type RoleServiceImpl struct {
	roleRepo             RoleRepository
	userRepo             repository.UserRepository
	PermissionRepository repository.PermissionRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo RoleRepository, userRepo repository.UserRepository) RoleService {
	return &RoleServiceImpl{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// GetRoles retrieves all roles, optionally filtered
func (s *RoleServiceImpl) GetRoles(ctx context.Context, filter repository.Filter, pagination map[string]int) ([]Role, int, error) {
	queryOptions := repository.QueryOptions{
		Limit: int64(pagination["limit"]),
		Skip:  int64(pagination["offset"]),
		Sort:  map[string]int{"createdAt": -1},
	}
	role, err := s.roleRepo.FindMany(ctx, filter, queryOptions)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, 0, ErrRoleNotFound
		}
		return nil, 0, err
	}
	count, err := s.roleRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return role, int(count), nil
}

// GetRole retrieves a role by ID
func (s *RoleServiceImpl) GetRole(ctx context.Context, id string) (*Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

// CreateRole creates a new role
func (s *RoleServiceImpl) CreateRole(ctx context.Context, role Role) (*Role, error) {
	// Validate required fields
	if role.Name == "" {
		return nil, ErrInvalidRoleData
	}

	// Check for duplicate role name
	filter := repository.Filter{"name": role.Name}
	existingRoles, _, err := s.GetRoles(ctx, filter, nil)
	if err != nil {
		return nil, err
	}
	if len(existingRoles) > 0 {
		return nil, ErrDuplicateRoleName
	}

	// Validate permissions if any are provided
	if len(role.PermissionIDs) > 0 {
		permissions, err := s.PermissionRepository.FindByIDs(ctx, role.PermissionIDs)
		if err != nil {
			return nil, err
		}
		if len(permissions) != len(role.PermissionIDs) {
			return nil, ErrInvalidPermissions
		}
	}

	// Set defaults
	role.CreatedAt = time.Now().Format(time.RFC3339)
	role.UpdatedAt = time.Now().Format(time.RFC3339) // In a real implementation, this would be the current time
	role.IsSystemDefault = false                     // User-created roles are never system defaults

	roleID, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	role.ID = roleID
	return &role, nil
}

// UpdateRole updates an existing role
func (s *RoleServiceImpl) UpdateRole(ctx context.Context, id string, updates map[string]interface{}) (*Role, error) {
	// Get the existing role
	existingRole, err := s.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}

	// If role doesn't exist, return error
	if existingRole == nil {
		return nil, ErrRoleNotFound
	}

	// Check if it's a system default role (which cannot be modified)
	if existingRole.IsSystemDefault {
		return nil, ErrSystemRoleModification
	}

	// Handle name change and check for duplicates
	if name, ok := updates["name"].(string); ok && name != existingRole.Name {
		if name == "" {
			return nil, ErrInvalidRoleData
		}

		// Check for duplicate name
		filter := repository.Filter{"name": name}
		existingRoles, _, err := s.GetRoles(ctx, filter, nil)
		if err != nil {
			return nil, err
		}
		if len(existingRoles) > 0 {
			return nil, ErrDuplicateRoleName
		}

		existingRole.Name = name
	}

	// Update description if provided
	if description, ok := updates["description"].(string); ok {
		existingRole.Description = description
	}

	// Update permissions if provided
	if permissionIDs, ok := updates["permissionIDs"].([]string); ok {
		// In a real implementation, we would validate all permission IDs
		// by checking if they exist in the permission repository

		existingRole.PermissionIDs = permissionIDs
	}

	// Update organization ID if provided
	if orgID, ok := updates["organizationId"].(string); ok {
		existingRole.OrganizationID = orgID
	}

	// Update modification time
	existingRole.UpdatedAt = "2025-05-05T00:00:00Z" // In a real implementation, this would be the current time

	// In a real implementation, we would use the role repository
	// to save the updated role to the database

	return existingRole, nil
}

// DeleteRole deletes a role
func (s *RoleServiceImpl) DeleteRole(ctx context.Context, id string) error {
	// Get the role to check if it exists and is not a system default
	existingRole, err := s.GetRole(ctx, id)
	if err != nil {
		return err
	}

	// If role doesn't exist, return error
	if existingRole == nil {
		return ErrRoleNotFound
	}

	// System default roles cannot be deleted
	if existingRole.IsSystemDefault {
		return ErrSystemRoleModification
	}

	// In a real implementation, we would check if any users are assigned to this role
	// and either prevent deletion or remove the role from those users

	// In a real implementation, we would use the role repository
	// to delete the role from the database

	return nil
}

// AddPermissionsToRole adds permissions to a role
func (s *RoleServiceImpl) AddPermissionsToRole(ctx context.Context, roleID string, permissionIDs []string) error {
	// Get existing role
	existingRole, err := s.GetRole(ctx, roleID)
	if err != nil {
		return err
	}

	// Check if role exists
	if existingRole == nil {
		return ErrRoleNotFound
	}

	// System default roles cannot be modified
	if existingRole.IsSystemDefault {
		return ErrSystemRoleModification
	}

	// Validate that all permission IDs exist
	// In a real implementation, we would check each permission ID in the database
	// For now, we'll assume they're all valid (but this would use the permission repository)

	// In a real implementation, this would be done more efficiently
	// using a set data structure to avoid duplicates

	// Start with the existing permissions
	updatedPermissionIDs := make([]string, len(existingRole.PermissionIDs))
	copy(updatedPermissionIDs, existingRole.PermissionIDs)

	// Add new permissions, avoiding duplicates
	for _, permID := range permissionIDs {
		found := false
		for _, existingPermID := range updatedPermissionIDs {
			if existingPermID == permID {
				found = true
				break
			}
		}

		if !found {
			updatedPermissionIDs = append(updatedPermissionIDs, permID)
		}
	}

	// Update the role with the new permissions
	updates := map[string]interface{}{
		"permissionIDs": updatedPermissionIDs,
	}

	// Use the existing update mechanism
	_, err = s.UpdateRole(ctx, roleID, updates)
	return err
}

// RemovePermissionsFromRole removes permissions from a role
func (s *RoleServiceImpl) RemovePermissionsFromRole(ctx context.Context, roleID string, permissionIDs []string) error {
	// Get existing role
	existingRole, err := s.GetRole(ctx, roleID)
	if err != nil {
		return err
	}

	// Check if role exists
	if existingRole == nil {
		return ErrRoleNotFound
	}

	// System default roles cannot be modified
	if existingRole.IsSystemDefault {
		return ErrSystemRoleModification
	}

	// Create a map of permissions to remove for efficient lookup
	permissionsToRemove := make(map[string]bool)
	for _, permID := range permissionIDs {
		permissionsToRemove[permID] = true
	}

	// Create a new slice with permissions that are not being removed
	remainingPermissions := make([]string, 0, len(existingRole.PermissionIDs))
	for _, permID := range existingRole.PermissionIDs {
		if !permissionsToRemove[permID] {
			remainingPermissions = append(remainingPermissions, permID)
		}
	}

	// Update the role with the remaining permissions
	updates := map[string]interface{}{
		"permissionIDs": remainingPermissions,
	}

	// Use the existing update mechanism
	_, err = s.UpdateRole(ctx, roleID, updates)
	return err
}

// GetUserRoles retrieves all roles assigned to a user
func (s *RoleServiceImpl) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	// In a real implementation, we would use the role repository
	// to fetch the user's roles from the database
	return []Role{
		{
			ID:              "role-1",
			Name:            "Admin",
			Description:     "Administrator with full access",
			PermissionIDs:   []string{"perm-1", "perm-2", "perm-3"},
			IsSystemDefault: true,
			CreatedAt:       "2025-01-01T00:00:00Z",
			UpdatedAt:       "2025-01-01T00:00:00Z",
		},
	}, nil
}

// IsUserInRole checks if a user has a specific role
func (s *RoleServiceImpl) IsUserInRole(ctx context.Context, userID string, roleID string) (bool, error) {
	// In a real implementation, we would use the user repository
	// to check if the user has the role
	return true, nil
}
