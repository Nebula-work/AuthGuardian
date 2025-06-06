package rbac

import (
	"context"
	"errors"
	"rbac-system/models"
	"rbac-system/pkg/common/repository"
	"time"
)

// AccessRequest represents a request to check access
type AccessRequest struct {
	UserID    string
	Resource  string
	Action    string
	OrgID     string
	Timestamp time.Time
	Context   map[string]interface{}
}

// AccessResponse contains detailed information about an access check
type AccessResponse struct {
	Allowed     bool
	Explanation string
	MatchedRule string
	Timestamp   time.Time
	UserRoles   []string
}

// Common errors
var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
)

// AccessControlServiceImpl implements AccessControlService
type AccessControlServiceImpl struct {
	roleRepo       RoleRepository
	permissionRepo PermissionRepository
	userRepo       repository.UserRepository
}

// NewAccessControlService creates a new access control service
func NewAccessControlService(roleRepo RoleRepository, permissionRepo PermissionRepository, userRepo repository.UserRepository) AccessControlService {
	return &AccessControlServiceImpl{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
	}
}

// CheckAccess checks if a user has access to a resource
func (s *AccessControlServiceImpl) CheckAccess(ctx context.Context, req AccessRequest) (bool, error) {
	// Validate request
	if req.UserID == "" || req.Resource == "" || req.Action == "" {
		return false, ErrInvalidInput
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return false, err
	}

	// Get user roles
	roles := make([]Role, 0, len(user.RoleIDs))
	for _, roleID := range user.RoleIDs {
		role, err := s.roleRepo.FindByID(ctx, roleID)
		if err != nil {
			continue
		}

		// Skip roles that don't match the organization if provided
		if req.OrgID != "" && role.OrganizationID != "" && role.OrganizationID != req.OrgID {
			continue
		}

		roles = append(roles, role)
	}

	// Check if any role has the required permission
	for _, role := range roles {
		permissions, err := s.getPermissionsForRole(ctx, role.ID)
		if err != nil {
			continue
		}

		for _, permission := range permissions {
			if permission.Resource == req.Resource && permission.Action == req.Action {
				// Permission matches exactly
				return true, nil
			}
			if permission.Resource == req.Resource && permission.Action == "*" {
				// Wildcard action for this resource
				return true, nil
			}
			if permission.Resource == "*" && permission.Action == req.Action {
				// Wildcard resource for this action
				return true, nil
			}
			if permission.Resource == "*" && permission.Action == "*" {
				// Full wildcard permission
				return true, nil
			}
		}
	}

	return false, nil
}

// CheckAccessDetailed checks access with detailed response
func (s *AccessControlServiceImpl) CheckAccessDetailed(ctx context.Context, req AccessRequest) (AccessResponse, error) {
	// Initialize response
	response := AccessResponse{
		Allowed:   false,
		Timestamp: time.Now(),
	}

	// Validate request
	if req.UserID == "" || req.Resource == "" || req.Action == "" {
		return response, ErrInvalidInput
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return response, err
	}

	// Get user roles
	roles := make([]Role, 0, len(user.RoleIDs))
	for _, roleID := range user.RoleIDs {
		role, err := s.roleRepo.FindByID(ctx, roleID)
		if err != nil {
			continue
		}

		// Skip roles that don't match the organization if provided
		if req.OrgID != "" && role.OrganizationID != "" && role.OrganizationID != req.OrgID {
			continue
		}

		roles = append(roles, role)
	}

	// Set user roles in response
	response.UserRoles = make([]string, len(roles))
	for i, role := range roles {
		response.UserRoles[i] = role.Name
	}

	// Check if any role has the required permission
	for _, role := range roles {
		permissions, err := s.getPermissionsForRole(ctx, role.ID)
		if err != nil {
			continue
		}

		for _, permission := range permissions {
			// Check for exact match
			if permission.Resource == req.Resource && permission.Action == req.Action {
				response.Allowed = true
				response.Explanation = "User has exact permission through role " + role.Name
				response.MatchedRule = permission.Name
				return response, nil
			}

			// Check for wildcard action
			if permission.Resource == req.Resource && permission.Action == "*" {
				response.Allowed = true
				response.Explanation = "User has wildcard action permission through role " + role.Name
				response.MatchedRule = permission.Name
				return response, nil
			}

			// Check for wildcard resource
			if permission.Resource == "*" && permission.Action == req.Action {
				response.Allowed = true
				response.Explanation = "User has wildcard resource permission through role " + role.Name
				response.MatchedRule = permission.Name
				return response, nil
			}

			// Check for full wildcard
			if permission.Resource == "*" && permission.Action == "*" {
				response.Allowed = true
				response.Explanation = "User has full wildcard permission through role " + role.Name
				response.MatchedRule = permission.Name
				return response, nil
			}
		}
	}

	response.Explanation = "User does not have required permission"
	return response, nil
}

// GetUserPermissions gets all permissions for a user
func (s *AccessControlServiceImpl) GetUserPermissions(ctx context.Context, userID string) ([]models.Permission, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all permissions for all roles
	allPermissions := make(map[string]models.Permission)
	for _, roleID := range user.RoleIDs {
		permissions, err := s.getPermissionsForRole(ctx, roleID)
		if err != nil {
			continue
		}

		for _, permission := range permissions {
			allPermissions[permission.ID] = permission
		}
	}

	// Convert map to slice
	permissions := make([]models.Permission, 0, len(allPermissions))
	for _, permission := range allPermissions {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetUserResources gets all resources a user can access
func (s *AccessControlServiceImpl) GetUserResources(ctx context.Context, userID string) ([]string, error) {
	// Get user permissions
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Extract unique resources
	resources := make(map[string]bool)
	for _, permission := range permissions {
		if permission.Resource == "*" {
			// Special case: if user has wildcard resource, get all resources
			allResources, err := s.getAllResources(ctx)
			if err != nil {
				// Continue with just the wildcard
				resources["*"] = true
			} else {
				for _, resource := range allResources {
					resources[resource] = true
				}
			}
		} else {
			resources[permission.Resource] = true
		}
	}

	// Convert map to slice
	resourceList := make([]string, 0, len(resources))
	for resource := range resources {
		resourceList = append(resourceList, resource)
	}

	return resourceList, nil
}

// GetUserActions gets all actions a user can perform on a resource
func (s *AccessControlServiceImpl) GetUserActions(ctx context.Context, userID, resource string) ([]string, error) {
	// Get user permissions
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Extract unique actions for the specified resource
	actions := make(map[string]bool)
	for _, permission := range permissions {
		if (permission.Resource == resource || permission.Resource == "*") && permission.Action == "*" {
			// Special case: if user has wildcard action, get all actions
			allActions, err := s.getAllActionsForResource(ctx, resource)
			if err != nil {
				// Continue with just the wildcard
				actions["*"] = true
			} else {
				for _, action := range allActions {
					actions[action] = true
				}
			}
		} else if permission.Resource == resource || permission.Resource == "*" {
			actions[permission.Action] = true
		}
	}

	// Convert map to slice
	actionList := make([]string, 0, len(actions))
	for action := range actions {
		actionList = append(actionList, action)
	}

	return actionList, nil
}

// HasPermission checks if a user has a specific permission
func (s *AccessControlServiceImpl) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Use CheckAccess
	return s.CheckAccess(ctx, AccessRequest{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	})
}

// Helper functions

// getPermissionsForRole gets all permissions for a role
func (s *AccessControlServiceImpl) getPermissionsForRole(ctx context.Context, roleID string) ([]models.Permission, error) {
	// Get role
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Get permissions
	permissions := make([]models.Permission, 0, len(role.PermissionIDs))
	for _, permissionID := range role.PermissionIDs {
		permission, err := s.permissionRepo.FindByID(ctx, permissionID)
		if err != nil {
			continue
		}

		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// getAllResources gets all unique resource names in the system
func (s *AccessControlServiceImpl) getAllResources(ctx context.Context) ([]string, error) {
	// Normally, we would query the database for all unique resource values
	// For simplicity, we'll return a hardcoded list
	return []string{
		"users",
		"roles",
		"permissions",
		"organizations",
	}, nil
}

// getAllActionsForResource gets all unique actions for a resource
func (s *AccessControlServiceImpl) getAllActionsForResource(ctx context.Context, resource string) ([]string, error) {
	// Normally, we would query the database for all unique action values for a resource
	// For simplicity, we'll return a hardcoded list
	return []string{
		"create",
		"read",
		"update",
		"delete",
		"list",
	}, nil
}
