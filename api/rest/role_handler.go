package rest

import (
	"errors"
	"rbac-system/core/rbac"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// RoleHandler handles HTTP requests related to roles
type RoleHandler struct {
	roleService       rbac.RoleService
	permissionService rbac.PermissionService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService rbac.RoleService, permissionService rbac.PermissionService) *RoleHandler {
	return &RoleHandler{
		roleService:       roleService,
		permissionService: permissionService,
	}
}

// RegisterRoutes registers all role routes
func (h *RoleHandler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	// Create a roles group with auth middleware applied to all routes
	roles := app.Group("/api/roles", authMiddleware)

	// Role endpoints
	roles.Get("/", h.GetRoles)
	roles.Post("/", h.CreateRole)
	roles.Get("/:id", h.GetRole)
	roles.Put("/:id", h.UpdateRole)
	roles.Delete("/:id", h.DeleteRole)

	// Role permissions management
	roles.Post("/:id/permissions", h.AddPermissions)
	roles.Delete("/:id/permissions", h.RemovePermissions)
}

// GetRoles handles GET /api/roles request
func (h *RoleHandler) GetRoles(c *fiber.Ctx) error {
	// Parse query parameters for filtering and pagination
	filter := make(map[string]interface{})

	// Extract organization filter if provided
	if orgID := c.Query("organizationId"); orgID != "" {
		filter["organizationId"] = orgID
	}

	// Extract name filter if provided
	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	// Parse pagination parameters
	pagination := make(map[string]int)
	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 {
			pagination["limit"] = limitInt
		}
	}

	if skip := c.Query("skip"); skip != "" {
		if skipInt, err := strconv.Atoi(skip); err == nil && skipInt >= 0 {
			pagination["skip"] = skipInt
		}
	}

	// Use the service to fetch roles based on filters and pagination
	ctx := c.Context()
	roles, total, err := h.roleService.GetRoles(ctx, filter, pagination)

	// Handle potential errors
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed_to_fetch_roles",
			"message": err.Error(),
		})
	}

	// Return successful response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"roles": roles,
			"total": total,
		},
	})
}

// CreateRole handles POST /api/roles request
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	// Parse request body
	var input struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		PermissionIDs  []string `json:"permissionIds"`
		OrganizationID string   `json:"organizationId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate required fields
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "validation_error",
			"message": "Role name is required",
			"fields":  []string{"name"},
		})
	}

	// Create role domain object
	role := rbac.Role{
		Name:           input.Name,
		Description:    input.Description,
		PermissionIDs:  input.PermissionIDs,
		OrganizationID: input.OrganizationID,
	}

	// Call service to create role
	ctx := c.Context()
	createdRole, err := h.roleService.CreateRole(ctx, role)

	// Handle errors
	if err != nil {
		// Determine appropriate error type and status code
		statusCode := fiber.StatusInternalServerError
		errorCode := "role_creation_failed"

		// Map domain-specific errors to HTTP responses
		if errors.Is(err, rbac.ErrDuplicateRoleName) {
			statusCode = fiber.StatusConflict
			errorCode = "duplicate_role_name"
		} else if errors.Is(err, rbac.ErrInvalidPermissions) {
			statusCode = fiber.StatusBadRequest
			errorCode = "invalid_permissions"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorCode,
			"message": err.Error(),
		})
	}

	// Return successful response with created role
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"role": createdRole,
		},
	})
}

// GetRole handles GET /api/roles/:id request
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	// In a real implementation, we would use the role service
	// to fetch a role by ID from the database
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"role": fiber.Map{
				"id":          roleID,
				"name":        "Sample Role",
				"description": "This is a sample role",
				"permissionIds": []string{
					"perm-1", "perm-2",
				},
			},
		},
	})
}

// UpdateRole handles PUT /api/roles/:id request
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	// Get role ID from path parameter
	roleID := c.Params("id")

	// Parse request body
	var input struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		PermissionIDs  []string `json:"permissionIds"`
		OrganizationID string   `json:"organizationId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Create updates map from input
	updates := make(map[string]interface{})

	// Only include fields that were actually provided in the request
	if input.Name != "" {
		updates["name"] = input.Name
	}

	if input.Description != "" {
		updates["description"] = input.Description
	}

	if input.PermissionIDs != nil {
		updates["permissionIDs"] = input.PermissionIDs
	}

	if input.OrganizationID != "" {
		updates["organizationId"] = input.OrganizationID
	}

	// Call service to update role
	ctx := c.Context()
	updatedRole, err := h.roleService.UpdateRole(ctx, roleID, updates)

	// Handle errors
	if err != nil {
		// Map different error types to appropriate HTTP status codes
		statusCode := fiber.StatusInternalServerError
		errorCode := "role_update_failed"

		if errors.Is(err, rbac.ErrRoleNotFound) {
			statusCode = fiber.StatusNotFound
			errorCode = "role_not_found"
		} else if errors.Is(err, rbac.ErrDuplicateRoleName) {
			statusCode = fiber.StatusConflict
			errorCode = "duplicate_role_name"
		} else if errors.Is(err, rbac.ErrInvalidPermissions) {
			statusCode = fiber.StatusBadRequest
			errorCode = "invalid_permissions"
		} else if errors.Is(err, rbac.ErrSystemRoleModification) {
			statusCode = fiber.StatusForbidden
			errorCode = "system_role_modification_not_allowed"
		} else if errors.Is(err, rbac.ErrInvalidRoleData) {
			statusCode = fiber.StatusBadRequest
			errorCode = "invalid_role_data"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorCode,
			"message": err.Error(),
		})
	}

	// Return successful response with updated role
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"role": updatedRole,
		},
	})
}

// DeleteRole handles DELETE /api/roles/:id request
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	// Get role ID from path parameter
	roleID := c.Params("id")

	// Call service to delete the role
	ctx := c.Context()
	err := h.roleService.DeleteRole(ctx, roleID)

	// Handle errors
	if err != nil {
		// Map different error types to appropriate HTTP status codes
		statusCode := fiber.StatusInternalServerError
		errorCode := "role_deletion_failed"

		if errors.Is(err, rbac.ErrRoleNotFound) {
			statusCode = fiber.StatusNotFound
			errorCode = "role_not_found"
		} else if errors.Is(err, rbac.ErrSystemRoleModification) {
			statusCode = fiber.StatusForbidden
			errorCode = "system_role_deletion_not_allowed"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorCode,
			"message": err.Error(),
		})
	}

	// Return successful response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Role deleted successfully",
		"data": fiber.Map{
			"id": roleID,
		},
	})
}

// AddPermissions handles POST /api/roles/:id/permissions request
func (h *RoleHandler) AddPermissions(c *fiber.Ctx) error {
	roleID := c.Params("id")

	var input struct {
		PermissionIds []string `json:"permissionIds"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// In a real implementation, we would use the role service
	// to add permissions to the role
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permissions added to role",
		"data": fiber.Map{
			"roleId":        roleID,
			"permissionIds": input.PermissionIds,
		},
	})
}

// RemovePermissions handles DELETE /api/roles/:id/permissions request
func (h *RoleHandler) RemovePermissions(c *fiber.Ctx) error {
	roleID := c.Params("id")

	var input struct {
		PermissionIds []string `json:"permissionIds"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// In a real implementation, we would use the role service
	// to remove permissions from the role
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permissions removed from role",
		"data": fiber.Map{
			"roleId":        roleID,
			"permissionIds": input.PermissionIds,
		},
	})
}
