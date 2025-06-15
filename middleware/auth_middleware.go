package middleware

import (
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware is a middleware that verifies JWT tokens
type AuthMiddleware struct {
	tokenService repository.TokenService
	userRepo     repository.UserRepository
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService repository.TokenService, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		userRepo:     userRepo,
	}
}

// Middleware handles authentication for Fiber
func (m *AuthMiddleware) Middleware(c *fiber.Ctx) error {
	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "missing_token",
			"message": "Authentication token is required",
		})
	}

	// Remove "Bearer " prefix if present
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Validate token
	claims, err := m.tokenService.ValidateToken(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_token",
			"message": err.Error(),
		})
	}

	// Get user
	user, err := m.userRepo.FindByID(c.Context(), claims.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "user_not_found",
			"message": "User associated with token not found",
		})
	}

	// Check if user is active
	if !user.Active {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "user_inactive",
			"message": "User is inactive",
		})
	}

	// Set user and claims in locals for use in handlers
	c.Locals("user", user)
	c.Locals("claims", claims)
	c.Locals("userId", claims.UserID)

	return c.Next()
}

// OptionalMiddleware handles optional authentication for Fiber
func (m *AuthMiddleware) OptionalMiddleware(c *fiber.Ctx) error {
	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Next()
	}

	// Remove "Bearer " prefix if present
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Validate token
	claims, err := m.tokenService.ValidateToken(c.Context(), token)
	if err != nil {
		return c.Next()
	}

	// Get user
	user, err := m.userRepo.FindByID(c.Context(), claims.UserID)
	if err != nil {
		return c.Next()
	}

	// Check if user is active
	if !user.Active {
		return c.Next()
	}

	// Set user and claims in locals for use in handlers
	c.Locals("user", user)
	c.Locals("claims", claims)
	c.Locals("userId", claims.UserID)

	return c.Next()
}

// AdminMiddleware handles admin authentication for Fiber
func (m *AuthMiddleware) AdminMiddleware(c *fiber.Ctx) error {
	// First apply standard auth middleware
	err := m.Middleware(c)
	if err != nil {
		return err
	}

	// Get claims from locals
	claims, ok := c.Locals("claims").(models.TokenClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "internal_error",
			"message": "Unable to retrieve user claims",
		})
	}

	// Check admin status (in a real implementation, this would check admin role)
	isAdmin := false
	for _, roleID := range claims.RoleIDs {
		// Check if this is an admin role
		// This is just a placeholder - in a real implementation,
		// you would query your database to check if this role has admin privileges
		if roleID == "admin" || roleID == "system_admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "admin_required",
			"message": "Admin privileges required",
		})
	}

	return c.Next()
}

// RequirePermissionMiddleware creates a middleware that requires a specific permission
func (m *AuthMiddleware) RequirePermissionMiddleware(resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First apply standard auth middleware
		err := m.Middleware(c)
		if err != nil {
			return err
		}

		// Get user ID from locals
		userID, ok := c.Locals("userId").(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "internal_error",
				"message": "Unable to retrieve user ID",
			})
		}

		// Check permission (in a real implementation, this would check against your RBAC system)
		// This is just a placeholder - in a real implementation, you would implement this check
		// using your RBAC service
		hasPermission := false

		// Simplified permission check for now
		user, err := m.userRepo.FindByID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "internal_error",
				"message": "Unable to retrieve user details",
			})
		}

		// For now, assume admin roles have all permissions
		for _, roleID := range user.RoleIDs {
			if roleID == "admin" || roleID == "system_admin" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "permission_denied",
				"message": "You don't have permission to access this resource",
			})
		}

		return c.Next()
	}
}
