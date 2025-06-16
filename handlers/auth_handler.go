package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Nebula-work/AuthGuardian/config"
	"github.com/Nebula-work/AuthGuardian/database"
	"github.com/Nebula-work/AuthGuardian/models"
	"github.com/Nebula-work/AuthGuardian/utils"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	config *config.Config
	db     *database.MongoClient
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, db *database.MongoClient) *AuthHandler {
	// Initialize OAuth providers
	goth.UseProviders(
		google.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			"http://localhost:5000/api/auth/oauth/google/callback",
			"email", "profile",
		),
		github.New(
			cfg.GitHubClientID,
			cfg.GitHubClientSecret,
			"http://localhost:5000/api/auth/oauth/github/callback",
			"user:email",
		),
	)

	return &AuthHandler{
		config: cfg,
		db:     db,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the provided details
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserCreateInput  true  "User registration details"
// @Success      201   {object}  models.UserSwaggerResponse               "User created successfully"
// @Failure      400                  "Invalid request body or missing fields"
// @Failure      409                  "Username or email already exists"
// github.com/Nebula-work/AuthGuardian               "Internal server error"
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Parse request body
	var input models.UserCreateInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Username, email, and password are required",
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

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to hash password",
		})
	}

	// Get the default user role
	rolesCollection := h.db.GetCollection(database.RolesCollection)
	defaultRole := models.Role{}

	err = rolesCollection.FindOne(
		context.Background(),
		bson.M{"name": models.UserRole, "isSystemDefault": true},
	).Decode(&defaultRole)

	var roleIDs []primitive.ObjectID
	if err == nil {
		roleIDs = append(roleIDs, defaultRole.ID)
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve default role",
		})
	}

	// Create user
	user := models.User{
		Username:        input.Username,
		Email:           input.Email,
		HashedPassword:  hashedPassword,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Active:          true,
		EmailVerified:   false,
		RoleIDs:         roleIDs,
		OrganizationIDs: input.OrganizationIDs,
		AuthProvider:    models.LocalAuth,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		h.config.JWTSecret,
		h.config.JWTExpirationTime,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate token",
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"token":   token,
		"user":    user.ToUserResponse(),
	})
}

// Login godoc
// @Summary      Login a new user
// @Description  Logs in a user account with the provided details
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      models.UserLoginInput  true  "User Login details"
// @Success      201   {object}  models.UserSwaggerResponse               "User Logged successfully"
// @Failure      400                  "Invalid request body or missing fields"
// @Failure      409                  "Username or email already exists"
// @Failure      500                 "Internal server error"
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse request body
	var input models.UserLoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	fmt.Println(input)
	if input.Username == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Username and password are required",
		})
	}

	// Find user by username
	collection := h.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{"username": input.Username, "authProvider": models.LocalAuth},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve user",
		})
	}

	// Verify password
	valid, err := utils.VerifyPassword(input.Password, user.HashedPassword)
	if err != nil || !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	// Check if user is active
	if !user.Active {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "User account is disabled",
		})
	}

	// Update last login
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"lastLogin": time.Now()}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update last login",
		})
	}

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		h.config.JWTSecret,
		h.config.JWTExpirationTime,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate token",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"token":   token,
		"user":    user.ToUserResponse(),
	})
}

// GoogleOAuth initiates Google OAuth flow
func (h *AuthHandler) GoogleOAuth(c *fiber.Ctx) error {
	return adaptor.HTTPHandlerFunc(gothic.BeginAuthHandler)(c)
}

// GoogleOAuthCallback handles the Google OAuth callback
func (h *AuthHandler) GoogleOAuthCallback(c *fiber.Ctx) error {
	return adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to complete authentication: " + err.Error(),
			})
			return
		}
		// Handle user after successful OAuth
		_ = h.handleOAuthUser(c, user, models.GoogleAuth)
	})(c)
}

// GitHubOAuth initiates GitHub OAuth flow
func (h *AuthHandler) GitHubOAuth(c *fiber.Ctx) error {
	return adaptor.HTTPHandlerFunc(gothic.BeginAuthHandler)(c)
}

// GitHubOAuthCallback handles the GitHub OAuth callback
func (h *AuthHandler) GitHubOAuthCallback(c *fiber.Ctx) error {
	return adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to complete authentication: " + err.Error(),
			})
			return
		}
		// Handle user after successful OAuth
		_ = h.handleOAuthUser(c, user, models.GitHubAuth)
	})(c)
}

// handleOAuthUser processes the authenticated OAuth user
func (h *AuthHandler) handleOAuthUser(c *fiber.Ctx, gothUser goth.User, provider models.AuthProvider) error {
	// Extract user info
	oauthInfo := models.OAuthUserInfo{
		ID:        gothUser.UserID,
		Email:     gothUser.Email,
		FirstName: gothUser.FirstName,
		LastName:  gothUser.LastName,
		Username:  gothUser.NickName,
	}

	// If no username, use email as fallback
	if oauthInfo.Username == "" {
		oauthInfo.Username = oauthInfo.Email
	}

	// Check if user already exists
	collection := h.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{
			"$or": []bson.M{
				{"email": oauthInfo.Email, "authProvider": provider},
				{"providerUserId": oauthInfo.ID, "authProvider": provider},
			},
		},
	).Decode(&existingUser)

	if err == nil {
		// User exists, update last login
		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"_id": existingUser.ID},
			bson.M{"$set": bson.M{"lastLogin": time.Now()}},
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to update last login",
			})
		}

		// Generate JWT token
		token, err := utils.GenerateToken(
			existingUser.ID,
			existingUser.Username,
			existingUser.Email,
			existingUser.RoleIDs,
			existingUser.OrganizationIDs,
			h.config.JWTSecret,
			h.config.JWTExpirationTime,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to generate token",
			})
		}

		// Return success response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"token":   token,
			"user":    existingUser.ToUserResponse(),
		})
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to check for existing user",
		})
	}

	// User doesn't exist, create new user
	// Get the default user role
	rolesCollection := h.db.GetCollection(database.RolesCollection)
	defaultRole := models.Role{}

	err = rolesCollection.FindOne(
		context.Background(),
		bson.M{"name": models.UserRole, "isSystemDefault": true},
	).Decode(&defaultRole)

	var roleIDs []primitive.ObjectID
	if err == nil {
		roleIDs = append(roleIDs, defaultRole.ID)
	} else if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve default role",
		})
	}

	// Create user
	user := models.User{
		Username:       oauthInfo.Username,
		Email:          oauthInfo.Email,
		FirstName:      oauthInfo.FirstName,
		LastName:       oauthInfo.LastName,
		Active:         true,
		EmailVerified:  true, // Email is verified by OAuth provider
		RoleIDs:        roleIDs,
		AuthProvider:   provider,
		ProviderUserID: oauthInfo.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastLogin:      time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		h.config.JWTSecret,
		h.config.JWTExpirationTime,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate token",
		})
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"token":   token,
		"user":    user.ToUserResponse(),
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Refreshes an access token using the provided valid token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Success      200            {object}  models.UserSwaggerResponse  "Token refreshed successfully"
// @Failure      400               "Invalid request or missing token"
// @Failure      401               "Unauthorized or invalid token"
// @Failure      403               "User account is disabled"
// @Failure      500                "Internal server error"
// @Router       /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Missing authorization header",
		})
	}

	// Extract the token
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid authorization format",
		})
	}

	// Validate the token
	claims, err := utils.ValidateToken(tokenParts[1], h.config.JWTSecret)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid token: " + err.Error(),
		})
	}

	// Find user by ID
	collection := h.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": claims.UserID},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to retrieve user",
		})
	}

	// Check if user is active
	if !user.Active {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "User account is disabled",
		})
	}

	// Generate new JWT token
	newToken, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		h.config.JWTSecret,
		h.config.JWTExpirationTime,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate token",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"token":   newToken,
		"user":    user.ToUserResponse(),
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// JWT is stateless, so there's not much to do on the server side
	// In a real app, you might want to blacklist the token

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}
