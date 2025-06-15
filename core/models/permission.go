package models

import (
	"time"
)

type Permission struct {
	ID              string    `json:"id"`
	Name            string    `bson:"name" json:"name"`
	Description     string    `bson:"description" json:"description"`
	Resource        string    `bson:"resource" json:"resource"`
	Action          string    `bson:"action" json:"action"`
	OrganizationID  string    `json:"organizationId,omitempty"`
	IsSystemDefault bool      `bson:"isSystemDefault" json:"isSystemDefault"`
	CreatedAt       time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time `bson:"updatedAt" json:"updatedAt"`
}
