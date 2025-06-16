package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Nebula-work/AuthGuardian/config"
	"github.com/Nebula-work/AuthGuardian/database"
	"github.com/Nebula-work/AuthGuardian/models"
)

// RoleHandler handles role-related requests
type RoleHandler struct {
	config *config.Config
	db     *database.MongoClient
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(cfg *config.Config, db *database.MongoClient) *RoleHandler {
	return &RoleHandler{
		config: cfg,
		db:     db,
	}
}

// GetRoles godoc
// @Summary      Retrieve all roles
// @Description  Fetch a list of roles with optional filters for pagination
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        limit          query     int     false  "Number of roles to retrieve (default: 100)"
// @Param        skip           query     int     false  "Number of roles to skip (default: 0)"
// @Param        organizationId query     string  false  "Filter roles by organization ID"
// @Success      200              "List of roles retrieved successfully"
// @Failure      400            "Invalid query parameters"
// @Failure      500            "Internal server error"
// @Router       /api/roles [get]
func (h *RoleHandler) GetRoles(c *fiber.Ctx) error {
	// Get query parameters
	limit := c.QueryInt("limit", 100)
	skip := c.QueryInt("skip", 0)
	orgID := c.Query("organizationId")

	// Prepare query
	query := bson.M{}

	// Add organization filter if provided
	if orgID != "" {
		objectID, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid organization ID",
			})
		}
		query["organizationId"] = objectID
	}

	// Get roles collection
	collection := h.db.GetCollection(database.RolesCollection)

	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"name": 1})

	// Execute query
	cursor, err := collection.Find(context.Background(), query, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve roles",
		})
	}
	defer cursor.Close(context.Background())

	// Decode roles
	var roles []models.Role
	if err := cursor.All(context.Background(), &roles); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode roles",
		})
	}

	// Convert roles to response objects
	var responses []models.RoleResponse
	for _, role := range roles {
		responses = append(responses, role.ToRoleResponse())
	}

	// Get total count
	count, err := collection.CountDocuments(context.Background(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to count roles",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"roles": responses,
			"total": count,
			"limit": limit,
			"skip":  skip,
		},
	})
}

// GetRole godoc
// @Summary      Retrieve a role by ID
// @Description  Fetch a single role by its unique identifier
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  models.RoleResponse  "Role retrieved successfully"
// @Failure      400           "Invalid role ID"
// @Failure      404           "Role not found"
// @Failure      500           "Internal server error"
// @Router       /api/roles/{id} [get]
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	// Get role ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Role ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid role ID",
		})
	}

	// Get role from database
	collection := h.db.GetCollection(database.RolesCollection)
	role := models.Role{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&role)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Role not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve role",
		})
	}

	// Return role
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    role.ToRoleResponse(),
	})
}

// CreateRole godoc
// @Summary      Create a new role
// @Description  Creates a new role with the provided details
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        role  body      models.RoleCreateInput  true  "Role creation details"
// @Success      201   {object}  models.RoleResponse     "Role created successfully"
// @Failure      400            "Invalid request body or missing fields"
// @Failure      409            "Role name already exists"
// @Failure      500            "Internal server error"
// @Router       /api/roles [post]
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	// Parse request body
	var input models.RoleCreateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Role name is required",
		})
	}

	// Check if role name already exists in the same organization
	collection := h.db.GetCollection(database.RolesCollection)

	// Query to check for role with same name in same organization or system default
	query := bson.M{"name": input.Name}
	if !input.OrganizationID.IsZero() {
		query["$or"] = []bson.M{
			{"organizationId": input.OrganizationID},
			{"isSystemDefault": true},
		}
	} else {
		query["isSystemDefault"] = true
	}

	existingRole := models.Role{}
	err := collection.FindOne(context.Background(), query).Decode(&existingRole)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   "Role name already exists in this organization or as a system default",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check for existing role",
		})
	}

	// Create role
	role := models.Role{
		Name:            input.Name,
		Description:     input.Description,
		OrganizationID:  input.OrganizationID,
		PermissionIDs:   input.PermissionIDs,
		IsSystemDefault: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Insert role
	result, err := collection.InsertOne(context.Background(), role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create role",
		})
	}

	// Get the inserted ID
	role.ID = result.InsertedID.(primitive.ObjectID)

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    role.ToRoleResponse(),
	})
}

// UpdateRole godoc
// @Summary      Update a role
// @Description  Updates an existing role with the provided details
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Role ID"
// @Param        role  body      models.RoleUpdateInput  true  "Role update details"
// @Success      200   {object}  models.RoleResponse     "Role updated successfully"
// @Failure      400            "Invalid role ID or request body"
// @Failure      404            "Role not found"
// @Failure      409            "Role name already exists"
// @Failure      500            "Internal server error"
// @Router       /api/roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	// Get role ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Role ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid role ID",
		})
	}

	// Parse request body
	var input models.RoleUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Check if role exists
	collection := h.db.GetCollection(database.RolesCollection)
	existingRole := models.Role{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingRole)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Role not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve role",
		})
	}

	// Prevent updating system default roles
	if existingRole.IsSystemDefault {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "System default roles cannot be modified",
		})
	}

	// Prepare update fields
	update := bson.M{"updatedAt": time.Now()}

	if input.Name != "" {
		// Check if name is already taken by another role in the same organization
		if input.Name != existingRole.Name {
			query := bson.M{
				"name": input.Name,
				"_id":  bson.M{"$ne": objectID},
			}

			if !existingRole.OrganizationID.IsZero() {
				query["organizationId"] = existingRole.OrganizationID
			}

			count, err := collection.CountDocuments(context.Background(), query)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Failed to check name availability",
				})
			}
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"error":   "Role name already taken in this organization",
				})
			}
		}
		update["name"] = input.Name
	}

	if input.Description != "" {
		update["description"] = input.Description
	}

	if len(input.PermissionIDs) > 0 {
		update["permissionIds"] = input.PermissionIDs
	}

	// Update role
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update role",
		})
	}

	// Get updated role
	updatedRole := models.Role{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedRole)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve updated role",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updatedRole.ToRoleResponse(),
	})
}

// DeleteRole godoc
// @Summary      Delete a role
// @Description  Deletes a role by its unique identifier
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  models.APIResponse  "Role deleted successfully"
// @Failure      400           "Invalid role ID"
// @Failure      404           "Role not found"
// @Failure      500           "Internal server error"
// @Router       /api/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	// Get role ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Role ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid role ID",
		})
	}

	// Check if role exists and is not a system default
	collection := h.db.GetCollection(database.RolesCollection)
	existingRole := models.Role{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingRole)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Role not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve role",
		})
	}

	// Prevent deleting system default roles
	if existingRole.IsSystemDefault {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "System default roles cannot be deleted",
		})
	}

	// Delete role
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete role",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Role not found",
		})
	}

	// Update users by removing this role from their role list
	usersCollection := h.db.GetCollection(database.UsersCollection)
	_, err = usersCollection.UpdateMany(
		context.Background(),
		bson.M{"roleIds": objectID},
		bson.M{"$pull": bson.M{"roleIds": objectID}},
	)

	if err != nil {
		// Log error but don't fail the request
		// In a production app, this should be handled more robustly
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Role deleted successfully",
	})
}

// AddPermissionsToRole godoc
// @Summary      Add permissions to a role
// @Description  Adds one or more permissions to a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id    path      string                              true  "Role ID"
// @Param        input body      models.RoleAddPermissionInput       true  "Permission IDs to add"
// @Success      200   {object}  models.RoleResponse                 "Permissions added to role successfully"
// @Failure      400            "Invalid role ID or request body"
// @Failure      404            "Role not found"
// @Failure      500            "Internal server error"
// @Router       /api/roles/{id}/permissions [post]
func (h *RoleHandler) AddPermissionsToRole(c *fiber.Ctx) error {
	// Get role ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Role ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid role ID",
		})
	}

	// Parse request body
	var input models.RoleAddPermissionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	if len(input.PermissionIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "At least one permission ID is required",
		})
	}

	// Check if role exists
	collection := h.db.GetCollection(database.RolesCollection)
	role := models.Role{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&role)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Role not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve role",
		})
	}

	// Check if permissions exist
	permissionsCollection := h.db.GetCollection(database.PermissionsCollection)
	var permissionCount int64
	permissionCount, err = permissionsCollection.CountDocuments(
		context.Background(),
		bson.M{"_id": bson.M{"$in": input.PermissionIDs}},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check permissions",
		})
	}

	if int(permissionCount) != len(input.PermissionIDs) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "One or more permissions do not exist",
		})
	}

	// Add permissions to role
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$addToSet": bson.M{
				"permissionIds": bson.M{"$each": input.PermissionIDs},
			},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to add permissions to role",
		})
	}

	// Get updated role
	updatedRole := models.Role{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedRole)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve updated role",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updatedRole.ToRoleResponse(),
	})
}

// RemovePermissionsFromRole godoc
// @Summary      Remove permissions from a role
// @Description  Removes one or more permissions from a role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Param        id    path      string                              true  "Role ID"
// @Param        input body      models.RoleRemovePermissionInput    true  "Permission IDs to remove"
// @Success      200   {object}  models.RoleResponse                 "Permissions removed from role successfully"
// @Failure      400            "Invalid role ID or request body"
// @Failure      404            "Role not found"
// @Failure      500            "Internal server error"
// @Router       /api/roles/{id}/permissions [delete]
func (h *RoleHandler) RemovePermissionsFromRole(c *fiber.Ctx) error {
	// Get role ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Role ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid role ID",
		})
	}

	// Parse request body
	var input models.RoleRemovePermissionInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	if len(input.PermissionIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "At least one permission ID is required",
		})
	}

	// Check if role exists
	collection := h.db.GetCollection(database.RolesCollection)
	role := models.Role{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&role)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Role not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve role",
		})
	}

	// Remove permissions from role
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$pull": bson.M{
				"permissionIds": bson.M{"$in": input.PermissionIDs},
			},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to remove permissions from role",
		})
	}

	// Get updated role
	updatedRole := models.Role{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedRole)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve updated role",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updatedRole.ToRoleResponse(),
	})
}
