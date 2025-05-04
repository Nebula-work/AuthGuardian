package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Permission represents a permission in the RBAC system
type Permission struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name           string             `bson:"name" json:"name"`
	Description    string             `bson:"description" json:"description"`
	Resource       string             `bson:"resource" json:"resource"`
	Action         string             `bson:"action" json:"action"`
	OrganizationID primitive.ObjectID `bson:"organizationId,omitempty" json:"organizationId,omitempty"`
	IsSystemDefault bool              `bson:"isSystemDefault" json:"isSystemDefault"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Common permission actions
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionAll    = "*"
)

// Common resources
const (
	ResourceUser         = "user"
	ResourceRole         = "role"
	ResourcePermission   = "permission"
	ResourceOrganization = "organization"
	ResourceAll          = "*"
)

// PermissionCreateInput represents the input to create a new permission
type PermissionCreateInput struct {
	Name           string             `json:"name" validate:"required"`
	Description    string             `json:"description"`
	Resource       string             `json:"resource" validate:"required"`
	Action         string             `json:"action" validate:"required"`
	OrganizationID primitive.ObjectID `json:"organizationId,omitempty"`
}

// PermissionUpdateInput represents the input to update a permission
type PermissionUpdateInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Action      string `json:"action,omitempty"`
}

// PermissionResponse represents the response for permission endpoints
type PermissionResponse struct {
	ID             primitive.ObjectID `json:"id,omitempty"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Resource       string             `json:"resource"`
	Action         string             `json:"action"`
	OrganizationID primitive.ObjectID `json:"organizationId,omitempty"`
	IsSystemDefault bool              `json:"isSystemDefault"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// ToPermissionResponse converts a Permission to a PermissionResponse
func (p *Permission) ToPermissionResponse() PermissionResponse {
	return PermissionResponse{
		ID:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		Resource:       p.Resource,
		Action:         p.Action,
		OrganizationID: p.OrganizationID,
		IsSystemDefault: p.IsSystemDefault,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
