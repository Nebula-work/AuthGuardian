package rest

import (
	"errors"
	"rbac-system/core/identity"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// OrganizationHandler manages organization-related REST endpoints
type OrganizationHandler struct {
	orgService identity.OrganizationService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgService identity.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
	}
}

// RegisterRoutes registers the organization routes
func (h *OrganizationHandler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	orgs := app.Group("/api/organizations", authMiddleware)

	orgs.Get("/", h.GetOrganizations)
	orgs.Post("/", h.CreateOrganization)
	orgs.Get("/:id", h.GetOrganization)
	orgs.Put("/:id", h.UpdateOrganization)
	orgs.Delete("/:id", h.DeleteOrganization)

	// User management
	orgs.Post("/:id/users", h.AddUser)
	orgs.Delete("/:id/users/:userId", h.RemoveUser)
	orgs.Get("/:id/users", h.GetOrganizationUsers)

	// Admin management
	orgs.Post("/:id/admins", h.AddAdmin)
	orgs.Delete("/:id/admins/:userId", h.RemoveAdmin)
	orgs.Get("/:id/admins", h.GetOrganizationAdmins)
	orgs.Get("/:id/admins/check/:userId", h.CheckIsAdmin)
}

// GetOrganizations returns all organizations
func (h *OrganizationHandler) GetOrganizations(c *fiber.Ctx) error {
	// Parse query parameters
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

	// Get organizations
	orgs, count, err := h.orgService.GetOrganizations(c.Context(), skip, limit)
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
			"organizations": orgs,
			"total":         count,
			"skip":          skip,
			"limit":         limit,
		},
	})
}

// GetOrganization returns an organization by ID
func (h *OrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Get organization
	org, err := h.orgService.GetOrganization(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "organization_not_found",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data":    org,
	})
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Domain      string   `json:"domain"`
		AdminIDs    []string `json:"adminIds"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_name",
			"message": "Organization name is required",
		})
	}

	// Create organization object
	org := repository.Organization{
		Organization: models.Organization{
			Name:        input.Name,
			Description: input.Description,
			Domain:      input.Domain,
			Active:      true,
			AdminIDs:    input.AdminIDs,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	// Create organization
	id, err := h.orgService.CreateOrganization(c.Context(), org)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "creation_failed"

		// Handle common errors
		if errors.Is(err, identity.ErrNameAlreadyExists) || errors.Is(err, identity.ErrDomainAlreadyExists) {
			statusCode = fiber.StatusConflict
			errorType = "organization_already_exists"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Get created organization
	createdOrg, err := h.orgService.GetOrganization(c.Context(), id)
	if err != nil {
		// Organization created but couldn't fetch it
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Organization created successfully",
			"data": fiber.Map{
				"id": id,
			},
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Organization created successfully",
		"data":    createdOrg,
	})
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Parse input
	var input struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Domain      string   `json:"domain"`
		Active      *bool    `json:"active"`
		AdminIDs    []string `json:"adminIds"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Get existing organization
	existingOrg, err := h.orgService.GetOrganization(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "organization_not_found",
			"message": err.Error(),
		})
	}

	// Update organization object
	org := existingOrg
	if input.Name != "" {
		org.Organization.Name = input.Name
	}
	if input.Description != "" {
		org.Organization.Description = input.Description
	}
	if input.Domain != "" {
		org.Organization.Domain = input.Domain
	}
	if input.Active != nil {
		org.Organization.Active = *input.Active
	}
	if input.AdminIDs != nil {
		org.Organization.AdminIDs = input.AdminIDs
	}
	org.Organization.UpdatedAt = time.Now().Format(time.RFC3339)

	// Update organization
	err = h.orgService.UpdateOrganization(c.Context(), id, org)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "update_failed"

		// Handle common errors
		if err == identity.ErrNameAlreadyExists || err == identity.ErrDomainAlreadyExists {
			statusCode = fiber.StatusConflict
			errorType = "organization_already_exists"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Get updated organization
	updatedOrg, err := h.orgService.GetOrganization(c.Context(), id)
	if err != nil {
		// Organization updated but couldn't fetch it
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Organization updated successfully",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Organization updated successfully",
		"data":    updatedOrg,
	})
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Delete organization
	err := h.orgService.DeleteOrganization(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "deletion_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
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
		"message": "Organization deleted successfully",
	})
}

// AddUser adds a user to an organization
func (h *OrganizationHandler) AddUser(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Parse input
	var input struct {
		UserID string `json:"userId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_user_id",
			"message": "User ID is required",
		})
	}

	// Add user to organization
	err := h.orgService.AddUserToOrganization(c.Context(), id, input.UserID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "add_user_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
		} else if err == identity.ErrUserNotFound {
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
		"message": "User added to organization successfully",
	})
}

// RemoveUser removes a user from an organization
func (h *OrganizationHandler) RemoveUser(c *fiber.Ctx) error {
	// Get organization ID and user ID from URL parameters
	id := c.Params("id")
	userID := c.Params("userId")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_user_id",
			"message": "User ID is required",
		})
	}

	// Remove user from organization
	err := h.orgService.RemoveUserFromOrganization(c.Context(), id, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "remove_user_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
		} else if err == identity.ErrUserNotFound {
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
		"message": "User removed from organization successfully",
	})
}

// GetOrganizationUsers returns all users in an organization
func (h *OrganizationHandler) GetOrganizationUsers(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Parse query parameters
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
	users, count, err := h.orgService.GetOrganizationUsers(c.Context(), id, skip, limit)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "fetch_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
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
			"users": users,
			"total": count,
			"skip":  skip,
			"limit": limit,
		},
	})
}

// AddAdmin adds an admin to an organization
func (h *OrganizationHandler) AddAdmin(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Parse input
	var input struct {
		UserID string `json:"userId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_user_id",
			"message": "User ID is required",
		})
	}

	// Add admin to organization
	err := h.orgService.AddAdminToOrganization(c.Context(), id, input.UserID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "add_admin_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
		} else if err == identity.ErrUserNotFound {
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
		"message": "Admin added to organization successfully",
	})
}

// RemoveAdmin removes an admin from an organization
func (h *OrganizationHandler) RemoveAdmin(c *fiber.Ctx) error {
	// Get organization ID and user ID from URL parameters
	id := c.Params("id")
	userID := c.Params("userId")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_user_id",
			"message": "User ID is required",
		})
	}

	// Remove admin from organization
	err := h.orgService.RemoveAdminFromOrganization(c.Context(), id, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "remove_admin_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
		} else if err == identity.ErrUserNotFound {
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
		"message": "Admin removed from organization successfully",
	})
}

// GetOrganizationAdmins returns all admins of an organization
func (h *OrganizationHandler) GetOrganizationAdmins(c *fiber.Ctx) error {
	// Get organization ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	// Get admins
	admins, err := h.orgService.GetOrganizationAdmins(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "fetch_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
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
			"admins": admins,
			"total":  len(admins),
		},
	})
}

// CheckIsAdmin checks if a user is an admin of an organization
func (h *OrganizationHandler) CheckIsAdmin(c *fiber.Ctx) error {
	// Get organization ID and user ID from URL parameters
	id := c.Params("id")
	userID := c.Params("userId")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Organization ID is required",
		})
	}

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_user_id",
			"message": "User ID is required",
		})
	}

	// Check if user is an admin
	isAdmin, err := h.orgService.IsUserAdmin(c.Context(), id, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "check_failed"

		// Handle common errors
		if err == identity.ErrOrganizationNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "organization_not_found"
		} else if err == identity.ErrUserNotFound {
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
			"isAdmin": isAdmin,
		},
	})
}
