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
	"github.com/Nebula-work/AuthGuardian/utils"
)

// UserHandler handles user-related requests
type UserHandler struct {
	config *config.Config
	db     *database.MongoClient
}

// NewUserHandler creates a new user handler
func NewUserHandler(cfg *config.Config, db *database.MongoClient) *UserHandler {
	return &UserHandler{
		config: cfg,
		db:     db,
	}
}

// GetUsers godoc
// @Summary      Retrieve all users
// @Description  Fetch a list of users with optional filters for pagination and organization
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        limit          query     int     false  "Number of users to retrieve (default: 100)"
// @Param        skip           query     int     false  "Number of users to skip (default: 0)"
// @Param        organizationId query     string  false  "Filter users by organization ID"
// @Success      200              "List of users retrieved successfully"
// @Failure      400                    "Invalid query parameters"
// @Failure      500                    "Internal server error"
// @Router       /api/users [get]
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// Get query parameters
	limit := c.QueryInt("limit", 100)
	skip := c.QueryInt("skip", 0)
	orgID := c.Query("organizationId")

	// Prepare a query
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
		query["organizationIds"] = objectID
	}

	// Get users collection
	collection := h.db.GetCollection(database.UsersCollection)

	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"username": 1})

	// Execute query
	cursor, err := collection.Find(context.Background(), query, findOptions)
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
	count, err := collection.CountDocuments(context.Background(), query)
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
			"users": responses,
			"total": count,
			"limit": limit,
			"skip":  skip,
		},
	})
}

// GetUser godoc
// @Summary      Retrieve a user by ID
// @Description  Fetch a single user by their unique identifier
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  models.UserResponse  "User retrieved successfully"
// @Failure      400           "Invalid user ID"
// @Failure      404           "User not found"
// @Failure      500           "Internal server error"
// @Router       /api/users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	// Get user ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "User ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	// Get user from database
	collection := h.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
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

	// Return user
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user.ToUserResponse(),
	})
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Creates a new user with the provided details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserCreateInput  true  "User creation details"
// @Success      201   {object}  models.UserResponse     "User created successfully"
// @Failure      400            "Invalid request body or missing fields"
// @Failure      409            "Username or email already exists"
// @Failure      500            "Internal server error"
// @Router       /api/users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// Parse request body
	var input models.UserCreateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	if input.Username == "" || input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Username and email are required",
		})
	}

	// For local auth, password is required
	if input.AuthProvider == models.LocalAuth && input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Password is required for local authentication",
		})
	}

	// Check if username or email already exists
	collection := h.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{
			"$or": []bson.M{
				{"username": input.Username},
				{"email": input.Email},
			},
		},
	).Decode(&existingUser)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   "Username or email already exists",
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check for existing user",
		})
	}

	// Create user
	user := models.User{
		Username:        input.Username,
		Email:           input.Email,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Active:          true,
		EmailVerified:   false,
		RoleIDs:         input.RoleIDs,
		OrganizationIDs: input.OrganizationIDs,
		AuthProvider:    input.AuthProvider,
		ProviderUserID:  input.ProviderUserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Hash password for local auth
	if input.AuthProvider == models.LocalAuth {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to hash password",
			})
		}
		user.HashedPassword = hashedPassword
	}

	// Insert user
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    user.ToUserResponse(),
	})
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Updates an existing user with the provided details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "User ID"
// @Param        user  body      models.UserUpdateInput  true  "User update details"
// @Success      200   {object}  models.UserResponse     "User updated successfully"
// @Failure      400            "Invalid user ID or request body"
// @Failure      404            "User not found"
// @Failure      409            "Username or email already exists"
// @Failure      500            "Internal server error"
// @Router       /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// Get user ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "User ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	// Parse request body
	var input models.UserUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Check if user exists
	collection := h.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingUser)

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

	// Prepare update fields
	update := bson.M{"updatedAt": time.Now()}

	if input.Username != "" {
		// Check if username is already taken by another user
		if input.Username != existingUser.Username {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"username": input.Username, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Failed to check username availability",
				})
			}
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"error":   "Username already taken",
				})
			}
		}
		update["username"] = input.Username
	}

	if input.Email != "" {
		// Check if email is already taken by another user
		if input.Email != existingUser.Email {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"email": input.Email, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Failed to check email availability",
				})
			}
			if count > 0 {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"error":   "Email already taken",
				})
			}
		}
		update["email"] = input.Email
	}

	if input.FirstName != "" {
		update["firstName"] = input.FirstName
	}

	if input.LastName != "" {
		update["lastName"] = input.LastName
	}

	if input.Active != nil {
		update["active"] = *input.Active
	}

	if len(input.RoleIDs) > 0 {
		update["roleIds"] = input.RoleIDs
	}

	if len(input.OrganizationIDs) > 0 {
		update["organizationIds"] = input.OrganizationIDs
	}

	// Update password if provided
	if input.Password != "" && existingUser.AuthProvider == models.LocalAuth {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to hash password",
			})
		}
		update["hashedPassword"] = hashedPassword
	}

	// Update user
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update user",
		})
	}

	// Get updated user
	updatedUser := models.User{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedUser)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve updated user",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updatedUser.ToUserResponse(),
	})
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Deletes a user by their unique identifier
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200    "User deleted successfully"
// @Failure      400           "Invalid user ID"
// @Failure      404           "User not found"
// @Failure      500           "Internal server error"
// @Router       /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// Get user ID from URL
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "User ID is required",
		})
	}

	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	// Delete user
	collection := h.db.GetCollection(database.UsersCollection)
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete user",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Retrieves the authenticated user's details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        userId    path      string                  true  "User ID"
// @Success      200  {object}  models.UserResponse  "Authenticated user retrieved successfully"
// @Failure      401           "User not authenticated"
// @Failure      404           "User not found"
// @Failure      500           "Internal server error"
// @Router       /api/users/me [get]
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	// Get user from database
	collection := h.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{"_id": userID},
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

	// Return user
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user.ToUserResponse(),
	})
}

// UpdateCurrentUser godoc
// @Summary      Update current user
// @Description  Updates the authenticated user's details
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        userId    path      string                  true  "User ID"
// @Success      200   {object}  models.UserResponse     "Authenticated user updated successfully"
// @Failure      400            "Invalid request body"
// @Failure      401            "User not authenticated"
// @Failure      500            "Internal server error"
// @Router       /api/users/me [put]
func (h *UserHandler) UpdateCurrentUser(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userId").(primitive.ObjectID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not authenticated",
		})
	}

	// Parse request body
	var input models.UserUpdateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Get user from database
	collection := h.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{"_id": userID},
	).Decode(&existingUser)

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

	// Prepare update fields (limited fields for self-update)
	update := bson.M{"updatedAt": time.Now()}

	if input.FirstName != "" {
		update["firstName"] = input.FirstName
	}

	if input.LastName != "" {
		update["lastName"] = input.LastName
	}

	// Update password if provided
	if input.Password != "" && existingUser.AuthProvider == models.LocalAuth {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to hash password",
			})
		}
		update["hashedPassword"] = hashedPassword
	}

	// Update user
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": update},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update user",
		})
	}

	// Get updated user
	updatedUser := models.User{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": userID},
	).Decode(&updatedUser)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve updated user",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    updatedUser.ToUserResponse(),
	})
}
