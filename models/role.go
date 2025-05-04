package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role represents a role in the RBAC system
type Role struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string               `bson:"name" json:"name"`
	Description     string               `bson:"description" json:"description"`
	OrganizationID  primitive.ObjectID   `bson:"organizationId,omitempty" json:"organizationId,omitempty"`
	PermissionIDs   []primitive.ObjectID `bson:"permissionIds" json:"permissionIds"`
	IsSystemDefault bool                 `bson:"isSystemDefault" json:"isSystemDefault"`
	CreatedAt       time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time            `bson:"updatedAt" json:"updatedAt"`
}

// System default roles
const (
	SystemAdminRole      = "system_admin"
	OrganizationAdminRole = "organization_admin"
	UserRole             = "user"
)

// RoleCreateInput represents the input to create a new role
type RoleCreateInput struct {
	Name           string               `json:"name" validate:"required"`
	Description    string               `json:"description"`
	OrganizationID primitive.ObjectID   `json:"organizationId,omitempty"`
	PermissionIDs  []primitive.ObjectID `json:"permissionIds,omitempty"`
}

// RoleUpdateInput represents the input to update a role
type RoleUpdateInput struct {
	Name          string               `json:"name,omitempty"`
	Description   string               `json:"description,omitempty"`
	PermissionIDs []primitive.ObjectID `json:"permissionIds,omitempty"`
}

// RoleResponse represents the response for role endpoints
type RoleResponse struct {
	ID              primitive.ObjectID   `json:"id,omitempty"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	OrganizationID  primitive.ObjectID   `json:"organizationId,omitempty"`
	PermissionIDs   []primitive.ObjectID `json:"permissionIds"`
	IsSystemDefault bool                 `json:"isSystemDefault"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

// ToRoleResponse converts a Role to a RoleResponse
func (r *Role) ToRoleResponse() RoleResponse {
	return RoleResponse{
		ID:              r.ID,
		Name:            r.Name,
		Description:     r.Description,
		OrganizationID:  r.OrganizationID,
		PermissionIDs:   r.PermissionIDs,
		IsSystemDefault: r.IsSystemDefault,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// RoleAddPermissionInput represents the input to add permissions to a role
type RoleAddPermissionInput struct {
	PermissionIDs []primitive.ObjectID `json:"permissionIds" validate:"required"`
}

// RoleRemovePermissionInput represents the input to remove permissions from a role
type RoleRemovePermissionInput struct {
	PermissionIDs []primitive.ObjectID `json:"permissionIds" validate:"required"`
}
