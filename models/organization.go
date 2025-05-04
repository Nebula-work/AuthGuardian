package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Organization represents an organization/tenant in the system
type Organization struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	Domain      string               `bson:"domain,omitempty" json:"domain,omitempty"`
	Active      bool                 `bson:"active" json:"active"`
	AdminIDs    []primitive.ObjectID `bson:"adminIds" json:"adminIds"`
	CreatedAt   time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time            `bson:"updatedAt" json:"updatedAt"`
}

// OrganizationCreateInput represents the input to create a new organization
type OrganizationCreateInput struct {
	Name        string               `json:"name" validate:"required"`
	Description string               `json:"description"`
	Domain      string               `json:"domain,omitempty"`
	AdminIDs    []primitive.ObjectID `json:"adminIds,omitempty"`
}

// OrganizationUpdateInput represents the input to update an organization
type OrganizationUpdateInput struct {
	Name        string               `json:"name,omitempty"`
	Description string               `json:"description,omitempty"`
	Domain      string               `json:"domain,omitempty"`
	Active      *bool                `json:"active,omitempty"`
	AdminIDs    []primitive.ObjectID `json:"adminIds,omitempty"`
}

// OrganizationResponse represents the response for organization endpoints
type OrganizationResponse struct {
	ID          primitive.ObjectID   `json:"id,omitempty"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Domain      string               `json:"domain,omitempty"`
	Active      bool                 `json:"active"`
	AdminIDs    []primitive.ObjectID `json:"adminIds"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

// ToOrganizationResponse converts an Organization to an OrganizationResponse
func (o *Organization) ToOrganizationResponse() OrganizationResponse {
	return OrganizationResponse{
		ID:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		Domain:      o.Domain,
		Active:      o.Active,
		AdminIDs:    o.AdminIDs,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

// OrganizationAddUserInput represents the input to add a user to an organization
type OrganizationAddUserInput struct {
	UserID      primitive.ObjectID `json:"userId" validate:"required"`
	RoleIDs     []primitive.ObjectID `json:"roleIds,omitempty"`
}

// OrganizationRemoveUserInput represents the input to remove a user from an organization
type OrganizationRemoveUserInput struct {
	UserID primitive.ObjectID `json:"userId" validate:"required"`
}
