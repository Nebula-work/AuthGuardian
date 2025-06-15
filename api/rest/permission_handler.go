package rest

import (
	"rbac-system/core/models"
	"rbac-system/core/rbac"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PermissionHandler manages permission-related REST endpoints
type PermissionHandler struct {
	permissionService rbac.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService rbac.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// RegisterRoutes registers the permission routes
func (h *PermissionHandler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	permissions := app.Group("/api/permissions", authMiddleware)

	permissions.Get("/", h.GetPermissions)
	permissions.Post("/", h.CreatePermission)
	permissions.Get("/:id", h.GetPermission)
	permissions.Put("/:id", h.UpdatePermission)
	permissions.Delete("/:id", h.DeletePermission)

	// Resource-specific endpoints
	permissions.Get("/resources/:resource", h.GetPermissionsByResource)
	permissions.Get("/resources/:resource/actions/:action", h.GetPermissionByResourceAndAction)
}

// GetPermissions returns all permissions
func (h *PermissionHandler) GetPermissions(c *fiber.Ctx) error {
	// Parse query parameters
	resource := c.Query("resource")
	action := c.Query("action")
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

	// Get permissions
	permissions, count, err := h.permissionService.GetPermissions(c.Context(), resource, action, orgID, skip, limit)
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
			"permissions": permissions,
			"total":       count,
			"skip":        skip,
			"limit":       limit,
		},
	})
}

// GetPermission returns a permission by ID
func (h *PermissionHandler) GetPermission(c *fiber.Ctx) error {
	// Get permission ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Permission ID is required",
		})
	}

	// Get permission
	permission, err := h.permissionService.GetPermission(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "fetch_failed"

		// Handle common errors
		if err == rbac.ErrNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "permission_not_found"
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
		"data":    permission,
	})
}

// CreatePermission creates a new permission
func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Resource        string `json:"resource"`
		Action          string `json:"action"`
		OrganizationID  string `json:"organizationId"`
		IsSystemDefault bool   `json:"isSystemDefault"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.Name == "" || input.Resource == "" || input.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_required_fields",
			"message": "Name, resource, and action are required",
		})
	}

	// Create permission object
	permission := models.Permission{
		Name:            input.Name,
		Description:     input.Description,
		Resource:        input.Resource,
		Action:          input.Action,
		OrganizationID:  input.OrganizationID,
		IsSystemDefault: input.IsSystemDefault,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create permission
	id, err := h.permissionService.CreatePermission(c.Context(), permission)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "creation_failed"

		// Handle common errors
		if err == rbac.ErrInvalidInput {
			statusCode = fiber.StatusBadRequest
			errorType = "invalid_input"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Get created permission
	createdPermission, err := h.permissionService.GetPermission(c.Context(), id)
	if err != nil {
		// Permission created but couldn't fetch it
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Permission created successfully",
			"data": fiber.Map{
				"id": id,
			},
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Permission created successfully",
		"data":    createdPermission,
	})
}

// UpdatePermission updates a permission
func (h *PermissionHandler) UpdatePermission(c *fiber.Ctx) error {
	// Get permission ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Permission ID is required",
		})
	}

	// Parse input
	var input struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Resource       string `json:"resource"`
		Action         string `json:"action"`
		OrganizationID string `json:"organizationId"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Get existing permission
	existingPermission, err := h.permissionService.GetPermission(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "fetch_failed"

		// Handle common errors
		if err == rbac.ErrNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "permission_not_found"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Check if permission is system default
	if existingPermission.IsSystemDefault {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "system_default",
			"message": "Cannot update system default permission",
		})
	}

	// Update permission object
	permission := existingPermission
	if input.Name != "" {
		permission.Name = input.Name
	}
	if input.Description != "" {
		permission.Description = input.Description
	}
	if input.Resource != "" {
		permission.Resource = input.Resource
	}
	if input.Action != "" {
		permission.Action = input.Action
	}
	if input.OrganizationID != "" {
		permission.OrganizationID = input.OrganizationID
	}
	permission.UpdatedAt = time.Now()

	// Update permission
	err = h.permissionService.UpdatePermission(c.Context(), id, permission)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "update_failed"

		// Handle common errors
		if err == rbac.ErrNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "permission_not_found"
		} else if err == rbac.ErrInvalidInput {
			statusCode = fiber.StatusBadRequest
			errorType = "invalid_input"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Get updated permission
	updatedPermission, err := h.permissionService.GetPermission(c.Context(), id)
	if err != nil {
		// Permission updated but couldn't fetch it
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Permission updated successfully",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Permission updated successfully",
		"data":    updatedPermission,
	})
}

// DeletePermission deletes a permission
func (h *PermissionHandler) DeletePermission(c *fiber.Ctx) error {
	// Get permission ID from URL parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_id",
			"message": "Permission ID is required",
		})
	}

	// Delete permission
	err := h.permissionService.DeletePermission(c.Context(), id)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "deletion_failed"

		// Handle common errors
		if err == rbac.ErrNotFound {
			statusCode = fiber.StatusNotFound
			errorType = "permission_not_found"
		} else if err == rbac.ErrInvalidInput {
			statusCode = fiber.StatusBadRequest
			errorType = "invalid_input"
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
		"message": "Permission deleted successfully",
	})
}

// GetPermissionsByResource returns permissions for a resource
func (h *PermissionHandler) GetPermissionsByResource(c *fiber.Ctx) error {
	// Get resource from URL parameter
	resource := c.Params("resource")
	if resource == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_resource",
			"message": "Resource is required",
		})
	}

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

	// Get permissions
	permissions, count, err := h.permissionService.GetPermissions(c.Context(), resource, "", orgID, skip, limit)
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
			"permissions": permissions,
			"total":       count,
			"skip":        skip,
			"limit":       limit,
		},
	})
}

// GetPermissionByResourceAndAction returns a permission by resource and action
func (h *PermissionHandler) GetPermissionByResourceAndAction(c *fiber.Ctx) error {
	// Get resource and action from URL parameters
	resource := c.Params("resource")
	action := c.Params("action")

	if resource == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_resource",
			"message": "Resource is required",
		})
	}

	if action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_action",
			"message": "Action is required",
		})
	}

	// Get organization ID from query parameter
	orgID := c.Query("organizationId")

	// Get permissions
	permissions, _, err := h.permissionService.GetPermissions(c.Context(), resource, action, orgID, 0, 1)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "fetch_failed",
			"message": err.Error(),
		})
	}

	if len(permissions) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "permission_not_found",
			"message": "Permission not found for this resource and action",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data":    permissions[0],
	})
}
