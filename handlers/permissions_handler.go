package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Nebula-work/AuthGuardian/config"
	"github.com/Nebula-work/AuthGuardian/database"
	"github.com/Nebula-work/AuthGuardian/models"
)

// PermissionHandler handles permission-related requests
type PermissionHandler struct {
	config *config.Config
	db     *database.MongoClient
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(cfg *config.Config, db *database.MongoClient) *PermissionHandler {
	return &PermissionHandler{
		config: cfg,
		db:     db,
	}
}

// GetPermissions retrieves all permissions
func (h *PermissionHandler) GetPermissions(c *fiber.Ctx) error {
	fmt.Println("GetPermissions")
	// Get query parameters
	limit := c.QueryInt("limit", 100)
	skip := c.QueryInt("skip", 0)
	resource := c.Query("resource")
	action := c.Query("action")
	orgID := c.Query("organizationId")

	// Prepare query
	query := bson.M{}

	// Add resource filter if provided
	if resource != "" {
		query["resource"] = resource
	}

	// Add action filter if provided
	if action != "" {
		query["action"] = action
	}

	// Add organization filter if provided - should show both organization-specific and system default permissions
	if orgID != "" {
		// Either match the organizationId or have isSystemDefault=true
		query["$or"] = []bson.M{
			{"organizationId": orgID},
			{"isSystemDefault": true},
		}
	}

	// Get permissions collection
	collection := h.db.GetCollection(database.PermissionsCollection)

	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"name": 1})

	// Execute query
	cursor, err := collection.Find(context.Background(), query, findOptions)
	fmt.Println(err)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve permissions",
		})
	}
	defer cursor.Close(context.Background())

	// Decode permissions
	var permissions []models.Permission
	if err := cursor.All(context.Background(), &permissions); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode permissions",
		})
	}

	// Convert permissions to response objects
	var responses []models.PermissionResponse
	for _, permission := range permissions {
		responses = append(responses, permission.ToPermissionResponse())
	}

	// Get total count
	count, err := collection.CountDocuments(context.Background(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to count permissions",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"permissions": responses,
			"total":       count,
			"limit":       limit,
			"skip":        skip,
		},
	})
}
