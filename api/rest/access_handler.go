package rest

import (
	"rbac-system/core/rbac"

	"github.com/gofiber/fiber/v2"
)

// AccessControlHandler manages access control-related REST endpoints
type AccessControlHandler struct {
	accessService rbac.AccessControlService
}

// NewAccessControlHandler creates a new access control handler
func NewAccessControlHandler(accessService rbac.AccessControlService) *AccessControlHandler {
	return &AccessControlHandler{
		accessService: accessService,
	}
}

// RegisterRoutes registers the access control routes
func (h *AccessControlHandler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	access := app.Group("/api/access", authMiddleware)

	access.Post("/check", h.CheckAccess)
	access.Post("/check/detailed", h.CheckAccessDetailed)
	access.Get("/permissions", h.GetUserPermissions)
	access.Get("/resources", h.GetUserResources)
	access.Get("/actions/:resource", h.GetUserActions)
	access.Get("/has/:resource/:action", h.HasPermission)
}

// CheckAccess checks if the user has access to a resource
func (h *AccessControlHandler) CheckAccess(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		UserID   string                 `json:"userId"`
		Resource string                 `json:"resource"`
		Action   string                 `json:"action"`
		OrgID    string                 `json:"organizationId"`
		Context  map[string]interface{} `json:"context"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// If userId is not provided, use the authenticated user
	if input.UserID == "" {
		userID, ok := c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "missing_user_id",
				"message": "User ID is required",
			})
		}
		input.UserID = userID
	}

	// Validate required fields
	if input.Resource == "" || input.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_required_fields",
			"message": "Resource and action are required",
		})
	}

	// Check access
	allowed, err := h.accessService.CheckAccess(c.Context(), rbac.AccessRequest{
		UserID:   input.UserID,
		Resource: input.Resource,
		Action:   input.Action,
		OrgID:    input.OrgID,
		Context:  input.Context,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "check_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"allowed": allowed,
		},
	})
}

// CheckAccessDetailed checks if the user has access to a resource with detailed info
func (h *AccessControlHandler) CheckAccessDetailed(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		UserID   string                 `json:"userId"`
		Resource string                 `json:"resource"`
		Action   string                 `json:"action"`
		OrgID    string                 `json:"organizationId"`
		Context  map[string]interface{} `json:"context"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// If userId is not provided, use the authenticated user
	if input.UserID == "" {
		userID, ok := c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "missing_user_id",
				"message": "User ID is required",
			})
		}
		input.UserID = userID
	}

	// Validate required fields
	if input.Resource == "" || input.Action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_required_fields",
			"message": "Resource and action are required",
		})
	}

	// Check access with detailed response
	response, err := h.accessService.CheckAccessDetailed(c.Context(), rbac.AccessRequest{
		UserID:   input.UserID,
		Resource: input.Resource,
		Action:   input.Action,
		OrgID:    input.OrgID,
		Context:  input.Context,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "check_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetUserPermissions gets all permissions for a user
func (h *AccessControlHandler) GetUserPermissions(c *fiber.Ctx) error {
	// Get user ID from query parameter or use authenticated user
	userID := c.Query("userId")
	if userID == "" {
		var ok bool
		userID, ok = c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "missing_user_id",
				"message": "User ID is required",
			})
		}
	}

	// Get permissions
	permissions, err := h.accessService.GetUserPermissions(c.Context(), userID)
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
			"total":       len(permissions),
		},
	})
}

// GetUserResources gets all resources a user can access
func (h *AccessControlHandler) GetUserResources(c *fiber.Ctx) error {
	// Get user ID from query parameter or use authenticated user
	userID := c.Query("userId")
	if userID == "" {
		var ok bool
		userID, ok = c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "missing_user_id",
				"message": "User ID is required",
			})
		}
	}

	// Get resources
	resources, err := h.accessService.GetUserResources(c.Context(), userID)
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
			"resources": resources,
			"total":     len(resources),
		},
	})
}

// GetUserActions gets all actions a user can perform on a resource
func (h *AccessControlHandler) GetUserActions(c *fiber.Ctx) error {
	// Get user ID from query parameter or use authenticated user
	userID := c.Query("userId")
	if userID == "" {
		var ok bool
		userID, ok = c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "missing_user_id",
				"message": "User ID is required",
			})
		}
	}

	// Get resource from URL parameter
	resource := c.Params("resource")
	if resource == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_resource",
			"message": "Resource is required",
		})
	}

	// Get actions
	actions, err := h.accessService.GetUserActions(c.Context(), userID, resource)
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
			"actions": actions,
			"total":   len(actions),
		},
	})
}

// HasPermission checks if a user has a specific permission
func (h *AccessControlHandler) HasPermission(c *fiber.Ctx) error {
	// Get user ID from query parameter or use authenticated user
	userID := c.Query("userId")
	if userID == "" {
		var ok bool
		userID, ok = c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "missing_user_id",
				"message": "User ID is required",
			})
		}
	}

	// Get resource and action from URL parameters
	resource := c.Params("resource")
	action := c.Params("action")

	if resource == "" || action == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_parameters",
			"message": "Resource and action are required",
		})
	}

	// Check permission
	hasPermission, err := h.accessService.HasPermission(c.Context(), userID, resource, action)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "check_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"hasPermission": hasPermission,
		},
	})
}
