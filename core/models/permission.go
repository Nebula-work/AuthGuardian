package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Permission struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string             `bson:"name" json:"name"`
	Description     string             `bson:"description" json:"description"`
	Resource        string             `bson:"resource" json:"resource"`
	Action          string             `bson:"action" json:"action"`
	OrganizationID  primitive.ObjectID `bson:"organizationId,omitempty" json:"organizationId,omitempty"`
	IsSystemDefault bool               `bson:"isSystemDefault" json:"isSystemDefault"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}
