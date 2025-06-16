package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rbac-system/config"
	"rbac-system/database"
	"rbac-system/models"
)

// OrganizationHandler handles organization-related requests
type OrganizationHandler struct {
	config *config.Config
	db     *database.MongoClient
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(cfg *config.Config, db *database.MongoClient) *OrganizationHandler {
	return &OrganizationHandler{
		config: cfg,
		db:     db,
	}
}

// GetOrganizations godoc
// @Summary      Retrieve all organizations
// @Description  Fetch a list of organizations with optional filters for pagination
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        limit  query     int     false  "Number of organizations to retrieve (default: 100)"
// @Param        skip   query     int     false  "Number of organizations to skip (default: 0)"
// @Success      200      "List of organizations retrieved successfully"
// @Failure      500                    "Internal server error"
// @Router       /api/organizations [get]
func (h *OrganizationHandler) GetOrganizations(c *fiber.Ctx) error {
	// Get query parameters
	limit := c.QueryInt("limit", 100)
	skip := c.QueryInt("skip", 0)

	// Get organizations collection
	collection := h.db.GetCollection(database.OrganizationsCollection)

	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"name": 1})

	// Execute query
	cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve organizations",
		})
	}
	defer cursor.Close(context.Background())

	// Decode organizations
	var organizations []models.Organization
	if err := cursor.All(context.Background(), &organizations); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode organizations",
		})
	}

	// Convert organizations to response objects
	var responses []models.OrganizationResponse
	for _, org := range organizations {
		responses = append(responses, org.ToOrganizationResponse())
	}

	// Get total count
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to count organizations",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"organizations": responses,
			"total":         count,
			"limit":         limit,
			"skip":          skip,
		},
	})
}

// GetOrganization godoc
// @Summary      Retrieve an organization by ID
// @Description  Fetch a single organization by its unique identifier
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  models.OrganizationResponse  "Organization retrieved successfully"
// @Failure      400           "Invalid organization ID"
// @Failure      404           "Organization not found"
// @Failure      500           "Internal server error"
// @Router       /api/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid organization ID",
		})
	}

	// Get organization from database
	collection := h.db.GetCollection(database.OrganizationsCollection)
	organization := models.Organization{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&organization)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Organization not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve organization",
		})
	}

	// Return organization
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    organization.ToOrganizationResponse(),
	})
}

// CreateOrganization godoc
// @Summary      Create a new organization
// @Description  Creates a new organization with the provided details
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        organization  body      models.OrganizationCreateInput  true  "Organization creation details"
// @Success      201           {object}  models.OrganizationResponse     "Organization created successfully"
// @Failure      400           "Invalid request body or missing fields"
// @Failure      409           "Organization name or domain already exists"
// @Failure      500           "Internal server error"
// @Router       /api/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	// Parse request body
	var input models.OrganizationCreateInput
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
			"error":   "Organization name is required",
		})
	}

	// Check if organization name already exists
	collection := h.db.GetCollection(database.OrganizationsCollection)
	existingOrg := models.Organization{}

	err := collection.FindOne(
		context.Background(),
		bson.M{"name": input.Name},
	).Decode(&existingOrg)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   "Organization name already exists",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check for existing organization",
		})
	}

	// If domain is provided, check for uniqueness
	if input.Domain != "" {
		err = collection.FindOne(
			context.Background(),
			bson.M{"domain": input.Domain},
		).Decode(&existingOrg)

		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error":   "Domain already registered for another organization",
			})
		} else if err != mongo.ErrNoDocuments {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to check for existing domain",
			})
		}
	}

	// Get current user as admin if no admins provided
	var adminIDs []primitive.ObjectID
	if len(input.AdminIDs) > 0 {
		adminIDs = input.AdminIDs
	} else {
		// Get user ID from context
		userID, ok := c.Locals("userId").(primitive.ObjectID)
		if ok {
			adminIDs = append(adminIDs, userID)
		}
	}

	// Create organization
	organization := models.Organization{
		Name:        input.Name,
		Description: input.Description,
		Domain:      input.Domain,
		Active:      true,
		AdminIDs:    adminIDs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Insert organization
	result, err := collection.InsertOne(context.Background(), organization)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create organization",
		})
	}

	// Get the inserted ID
	organization.ID = result.InsertedID.(primitive.ObjectID)

	// Update admin users with this organization
	if len(adminIDs) > 0 {
		usersCollection := h.db.GetCollection(database.UsersCollection)
		_, err = usersCollection.UpdateMany(
			context.Background(),
			bson.M{"_id": bson.M{"$in": adminIDs}},
			bson.M{"$addToSet": bson.M{"organizationIds": organization.ID}},
		)

		if err != nil {
			// Log error but don't fail the request
			// In a production app, this should be handled more robustly
			// For simplicity, we just continue
		}
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    organization.ToOrganizationResponse(),
	})
}

// UpdateOrganization godoc
// @Summary      Update an organization
// @Description  Updates an existing organization with the provided details
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id            path      string                          true  "Organization ID"
// @Param        organization  body      models.OrganizationUpdateInput  true  "Organization update details"
// @Success      200           {object}  models.OrganizationResponse     "Organization updated successfully"
// @Failure      400           "Invalid organization ID or request body"
// @Failure      404           "Organization not found"
// @Failure      409           "Organization name or domain already exists"
// @Failure      500           "Internal server error"
// @Router       /api/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid organization ID",
		})
	}

	// Parse request body
	var input models.OrganizationUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Check if organization exists
	collection := h.db.GetCollection(database.OrganizationsCollection)
	existingOrg := models.Organization{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingOrg)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Organization not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve organization",
		})
	}

	// Prepare update fields
	update := bson.M{"updatedAt": time.Now()}

	if input.Name != "" {
		// Check if name is already taken by another organization
		if input.Name != existingOrg.Name {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"name": input.Name, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Failed to check name availability",
				})
			}
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"error":   "Organization name already taken",
				})
			}
		}
		update["name"] = input.Name
	}

	if input.Description != "" {
		update["description"] = input.Description
	}

	if input.Domain != "" {
		// Check if domain is already taken by another organization
		if input.Domain != existingOrg.Domain {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"domain": input.Domain, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Failed to check domain availability",
				})
			}
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"error":   "Domain already registered for another organization",
				})
			}
		}
		update["domain"] = input.Domain
	}

	if input.Active != nil {
		update["active"] = *input.Active
	}

	if len(input.AdminIDs) > 0 {
		update["adminIds"] = input.AdminIDs
	}

	// Update organization
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update organization",
		})
	}

	// Get updated organization
	updatedOrg := models.Organization{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedOrg)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve updated organization",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updatedOrg.ToOrganizationResponse(),
	})
}

// DeleteOrganization godoc
// @Summary      Delete an organization
// @Description  Deletes an organization by its unique identifier
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  models.APIResponse  "Organization deleted successfully"
// @Failure      400           "Invalid organization ID"
// @Failure      404           "Organization not found"
// @Failure      500           "Internal server error"
// @Router       /api/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid organization ID",
		})
	}

	// Delete organization
	collection := h.db.GetCollection(database.OrganizationsCollection)
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete organization",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Organization not found",
		})
	}

	// Update users by removing this organization
	usersCollection := h.db.GetCollection(database.UsersCollection)
	_, err = usersCollection.UpdateMany(
		context.Background(),
		bson.M{"organizationIds": objectID},
		bson.M{"$pull": bson.M{"organizationIds": objectID}},
	)

	if err != nil {
		// Log error but don't fail the request
		// In a production app, this should be handled more robustly
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Organization deleted successfully",
	})
}

// AddUserToOrganization godoc
// @Summary      Add a user to an organization
// @Description  Adds a user to an organization with optional role assignments
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id    path      string                              true  "Organization ID"
// @Param        input body      models.OrganizationAddUserInput     true  "User and role details"
// @Success      200   {object}  models.APIResponse                  "User added to organization successfully"
// @Failure      400            "Invalid organization ID or request body"
// @Failure      404            "Organization or user not found"
// @Failure      500            "Internal server error"
// @Router       /api/organizations/{id}/users [post]
func (h *OrganizationHandler) AddUserToOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID is required",
		})
	}

	// Convert string ID to ObjectID
	orgID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid organization ID",
		})
	}

	// Parse request body
	var input models.OrganizationAddUserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Check if organization exists
	orgCollection := h.db.GetCollection(database.OrganizationsCollection)
	organization := models.Organization{}

	err = orgCollection.FindOne(
		context.Background(),
		bson.M{"_id": orgID},
	).Decode(&organization)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Organization not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve organization",
		})
	}

	// Check if user exists
	userCollection := h.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err = userCollection.FindOne(
		context.Background(),
		bson.M{"_id": input.UserID},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve user",
		})
	}

	// Update user's organization list
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": input.UserID},
		bson.M{"$addToSet": bson.M{"organizationIds": orgID}},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to add user to organization",
		})
	}

	// If role IDs are provided, update user's roles
	if len(input.RoleIDs) > 0 {
		// TODO: Validate that the roles exist and belong to the organization
		_, err = userCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": input.UserID},
			bson.M{"$addToSet": bson.M{"roleIds": bson.M{"$each": input.RoleIDs}}},
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to update user roles",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User added to organization successfully",
	})
}

// RemoveUserFromOrganization godoc
// @Summary      Remove a user from an organization
// @Description  Removes a user from an organization by their unique identifier
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Organization ID"
// @Param        userId  path      string  true  "User ID"
// @Success      200     {object}  models.APIResponse  "User removed from organization successfully"
// @Failure      400                "Invalid organization ID or user ID"
// @Failure      404                "Organization or user not found"
// @Failure      500                "Internal server error"
// @Router       /api/organizations/{id}/users/{userId} [delete]
func (h *OrganizationHandler) RemoveUserFromOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL
	orgID := c.Params("id")
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID is required",
		})
	}

	// Get user ID from URL
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "User ID is required",
		})
	}

	// Convert string IDs to ObjectIDs
	orgObjectID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid organization ID",
		})
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	// Check if organization exists
	orgCollection := h.db.GetCollection(database.OrganizationsCollection)
	organization := models.Organization{}

	err = orgCollection.FindOne(
		context.Background(),
		bson.M{"_id": orgObjectID},
	).Decode(&organization)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Organization not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve organization",
		})
	}

	// Check if user exists and belongs to the organization
	userCollection := h.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err = userCollection.FindOne(
		context.Background(),
		bson.M{"_id": userObjectID, "organizationIds": orgObjectID},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found or not a member of the organization",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve user",
		})
	}

	// Update user's organization list
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$pull": bson.M{"organizationIds": orgObjectID}},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to remove user from organization",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User removed from organization successfully",
	})
}

// GetOrganizationUsers godoc
// @Summary      Get all users in an organization
// @Description  Retrieves a list of users belonging to a specific organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        id     path      string  true   "Organization ID"
// @Param        limit  query     int     false  "Number of users to retrieve (default: 100)"
// @Param        skip   query     int     false  "Number of users to skip (default: 0)"
// @Success      200      "List of users retrieved successfully"
// @Failure      400             "Invalid organization ID"
// @Failure      404             "Organization not found"
// @Failure      500             "Internal server error"
// @Router       /api/organizations/{id}/users [get]
func (h *OrganizationHandler) GetOrganizationUsers(c *fiber.Ctx) error {
	// Get organization ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Organization ID is required",
		})
	}

	// Convert string ID to ObjectID
	orgID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid organization ID",
		})
	}

	// Get query parameters
	limit := c.QueryInt("limit", 100)
	skip := c.QueryInt("skip", 0)

	// Check if organization exists
	orgCollection := h.db.GetCollection(database.OrganizationsCollection)
	organization := models.Organization{}

	err = orgCollection.FindOne(
		context.Background(),
		bson.M{"_id": orgID},
	).Decode(&organization)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Organization not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve organization",
		})
	}

	// Get users in organization
	userCollection := h.db.GetCollection(database.UsersCollection)

	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"username": 1})

	// Execute query
	cursor, err := userCollection.Find(
		context.Background(),
		bson.M{"organizationIds": orgID},
		findOptions,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve users",
		})
	}
	defer cursor.Close(context.Background())

	// Decode users
	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to decode users",
		})
	}

	// Convert users to response objects
	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToUserResponse())
	}

	// Get total count
	count, err := userCollection.CountDocuments(
		context.Background(),
		bson.M{"organizationIds": orgID},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to count users",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"users":        responses,
			"total":        count,
			"limit":        limit,
			"skip":         skip,
			"organization": organization.ToOrganizationResponse(),
		},
	})
}
