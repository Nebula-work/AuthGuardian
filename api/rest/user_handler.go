package rest

import (
	"rbac-system/core/identity"
	"rbac-system/pkg/common/repository"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// UserHandler manages user-related REST endpoints
type UserHandler struct {
	userService identity.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService identity.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterRoutes registers the user routes
func (h *UserHandler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	users := app.Group("/api/users", authMiddleware)

	users.Get("/", h.GetUsers)
	users.Post("/", h.CreateUser)
	users.Get("/me", h.GetCurrentUser)
	users.Get("/:id", h.GetUser)
	users.Put("/:id", h.UpdateUser)
	users.Delete("/:id", h.DeleteUser)

	// Password management
	users.Post("/:id/password", h.SetPassword)

	// Role management
	users.Post("/:id/roles", h.AddRole)
	users.Delete("/:id/roles/:roleId", h.RemoveRole)
	users.Get("/:id/roles", h.GetUserRoles)

	// Organization management
	users.Post("/:id/organizations", h.AddOrganization)
	users.Delete("/:id/organizations/:orgId", h.RemoveOrganization)
	users.Get("/:id/organizations", h.GetUserOrganizations)
}

// GetUsers returns all users
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// Parse query parameters
	orgID := c.Query("organizationId")
	skipStr := c.Query("skip", "0")
	limitStr := c.Query("limit", "100")

	// Parse pagination parameters
	skip, err := strconv.ParseInt(skipStr, 10, 64)
	if err != nil {
		skip = 0
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit = 100
	}

	// Get users
	users, count, err := h.userService.GetUsers(c.Context(), orgID, skip, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "fetch_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"users": users,
			"total": count,
			"skip":  skip,
			"limit": limit,
		},
	})
}

// GetUser returns a user by ID
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Get user
	user, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "user_not_found",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	// Get user from context
	user, ok := c.Locals("user").(repository.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "unauthorized",
			"message": "Not authenticated",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Username        string   `json:"username"`
		Email           string   `json:"email"`
		Password        string   `json:"password"`
		FirstName       string   `json:"firstName"`
		LastName        string   `json:"lastName"`
		RoleIDs         []string `json:"roleIds"`
		OrganizationIDs []string `json:"organizationIds"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_required_fields",
			"message": "Username, email, and password are required",
		})
	}

	// Create user object
	user := repository.User{
		Username:        input.Username,
		Email:           input.Email,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Active:          true,
		EmailVerified:   false,
		RoleIDs:         input.RoleIDs,
		OrganizationIDs: input.OrganizationIDs,
		AuthProvider:    "local",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create user
	id, err := h.userService.CreateUser(c.Context(), user)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "creation_failed"

		// Handle common errors
		if err == identity.ErrEmailAlreadyExists || err == identity.ErrUsernameAlreadyExists {
			statusCode = fiber.StatusConflict
			errorType = "user_already_exists"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Set password
	err = h.userService.SetPassword(c.Context(), id, input.Password)
	if err != nil {
		// User created but couldn't set password
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "password_error",
			"message": "User created but failed to set password: " + err.Error(),
		})
	}

	// Get created user
	createdUser, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		// User created but couldn't fetch it
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "User created successfully",
			"data": fiber.Map{
				"id": id,
			},
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data":    createdUser,
	})
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Parse input
	var input struct {
		Username        string   `json:"username"`
		Email           string   `json:"email"`
		FirstName       string   `json:"firstName"`
		LastName        string   `json:"lastName"`
		Active          *bool    `json:"active"`
		EmailVerified   *bool    `json:"emailVerified"`
		RoleIDs         []string `json:"roleIds"`
		OrganizationIDs []string `json:"organizationIds"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Get existing user
	existingUser, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "user_not_found",
			"message": err.Error(),
		})
	}

	// Update user object
	user := existingUser
	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Active != nil {
		user.Active = *input.Active
	}
	if input.EmailVerified != nil {
		user.EmailVerified = *input.EmailVerified
	}
	if input.RoleIDs != nil {
		user.RoleIDs = input.RoleIDs
	}
	if input.OrganizationIDs != nil {
		user.OrganizationIDs = input.OrganizationIDs
	}
	user.UpdatedAt = time.Now()

	// Update user
	err = h.userService.UpdateUser(c.Context(), id, user)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "update_failed"

		// Handle common errors
		if err == identity.ErrEmailAlreadyExists || err == identity.ErrUsernameAlreadyExists {
			statusCode = fiber.StatusConflict
			errorType = "user_already_exists"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Get updated user
	updatedUser, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		// User updated but couldn't fetch it
		return c.JSON(fiber.Map{
			"success": true,
			"message": "User updated successfully",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
		"data":    updatedUser,
	})
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Delete user
	err := h.userService.DeleteUser(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "deletion_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}

// SetPassword sets a user's password
func (h *UserHandler) SetPassword(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Parse input
	var input struct {
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_password",
			"message": "Password is required",
		})
	}

	// Set password
	err := h.userService.SetPassword(c.Context(), id, input.Password)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "password_update_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password updated successfully",
	})
}

// AddRole adds a role to a user
func (h *UserHandler) AddRole(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Parse input
	var input struct {
		RoleID string `json:"roleId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.RoleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_role_id",
			"message": "Role ID is required",
		})
	}

	// Add role
	err := h.userService.AddRoleToUser(c.Context(), id, input.RoleID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "add_role_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role added successfully",
	})
}

// RemoveRole removes a role from a user
func (h *UserHandler) RemoveRole(c *fiber.Ctx) error {
	// Get user ID and role ID from URL parameters
	id := c.Params("id")
	roleID := c.Params("roleId")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	if roleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_role_id",
			"message": "Role ID is required",
		})
	}

	// Remove role
	err := h.userService.RemoveRoleFromUser(c.Context(), id, roleID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "remove_role_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role removed successfully",
	})
}

// GetUserRoles returns all roles for a user
func (h *UserHandler) GetUserRoles(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Get roles
	roleIDs, err := h.userService.GetUserRoles(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "fetch_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"roleIds": roleIDs,
			"total":   len(roleIDs),
		},
	})
}

// AddOrganization adds an organization to a user
func (h *UserHandler) AddOrganization(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Parse input
	var input struct {
		OrganizationID string `json:"organizationId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.OrganizationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_organization_id",
			"message": "Organization ID is required",
		})
	}

	// Add organization
	err := h.userService.AddOrganizationToUser(c.Context(), id, input.OrganizationID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "add_organization_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Organization added successfully",
	})
}

// RemoveOrganization removes an organization from a user
func (h *UserHandler) RemoveOrganization(c *fiber.Ctx) error {
	// Get user ID and organization ID from URL parameters
	id := c.Params("id")
	orgID := c.Params("orgId")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_organization_id",
			"message": "Organization ID is required",
		})
	}

	// Remove organization
	err := h.userService.RemoveOrganizationFromUser(c.Context(), id, orgID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "remove_organization_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Organization removed successfully",
	})
}

// GetUserOrganizations returns all organizations for a user
func (h *UserHandler) GetUserOrganizations(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "User ID is required",
		})
	}

	// Get organizations
	orgIDs, err := h.userService.GetUserOrganizations(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "fetch_failed"

		// Handle common errors
		if err == identity.ErrUserNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "user_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"organizationIds": orgIDs,
			"total":           len(orgIDs),
		},
	})
}
