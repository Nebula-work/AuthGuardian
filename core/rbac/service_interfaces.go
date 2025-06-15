package rbac

import (
	"context"
)

// AccessControlService defines the interface for access control operations
type AccessControlService interface {
	// CheckAccess checks if a user has access to a resource
	CheckAccess(ctx context.Context, req AccessRequest) (bool, error)

	// CheckAccessDetailed checks if a user has access to a resource and returns a detailed response
	CheckAccessDetailed(ctx context.Context, req AccessRequest) (AccessResponse, error)

	// GetUserPermissions gets permissions for a user
	GetUserPermissions(ctx context.Context, userID string) ([]Permission, error)

	// GetUserResources gets resources accessible to a user
	GetUserResources(ctx context.Context, userID string) ([]string, error)

	// GetUserActions gets actions a user can perform on a resource
	GetUserActions(ctx context.Context, userID, resource string) ([]string, error)

	// HasPermission checks if a user has a specific permission
	HasPermission(ctx context.Context, userID, resource, action string) (bool, error)
}
