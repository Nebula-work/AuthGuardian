package rest

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"rbac-system/core/password"
	"rbac-system/pkg/common/repository"
	"strings"
)

// AuthHandler manages authentication REST endpoints
type AuthHandler struct {
	authService     repository.AuthService
	oauthService    repository.OAuthService
	passwordManager auth.PasswordManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	authService repository.AuthService,
	oauthService repository.OAuthService,
	passwordManager auth.PasswordManager,
) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		oauthService:    oauthService,
		passwordManager: passwordManager,
	}
}

// RegisterRoutes registers the auth routes
func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")

	// Authentication routes
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)
	auth.Post("/logout", h.Logout)
	auth.Post("/refresh-token", h.RefreshToken)
	auth.Post("/reset-password", h.RequestPasswordReset)
	auth.Post("/reset-password/complete", h.CompletePasswordReset)

	// OAuth routes
	oauth := auth.Group("/oauth")
	oauth.Get("/:provider", h.InitiateOAuthLogin)
	oauth.Get("/:provider/callback", h.CompleteOAuthLogin)
	oauth.Post("/link/:provider", h.LinkOAuthAccount)
	oauth.Delete("/link/:provider", h.UnlinkOAuthAccount)
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Username string `json:"username"`
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
	if input.Username == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_credentials",
			"message": "Username and password are required",
		})
	}

	// Authenticate user
	authInfo, err := h.authService.Login(c.Context(), input.Username, input.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "authentication_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success":      true,
		"token":        authInfo.Token,
		"refreshToken": authInfo.RefreshToken,
		"user":         authInfo.User,
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
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

	// Register user
	authInfo, err := h.authService.Register(c.Context(), repository.RegisterRequest{
		Username:  input.Username,
		Email:     input.Email,
		Password:  input.Password,
		FirstName: input.FirstName,
		LastName:  input.LastName,
	})

	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errorType := "registration_failed"

		// Handle common errors
		if strings.Contains(err.Error(), "already exists") {
			statusCode = fiber.StatusConflict
			errorType = "user_already_exists"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"error":   errorType,
			"message": err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":      true,
		"token":        authInfo.Token,
		"refreshToken": authInfo.RefreshToken,
		"user":         authInfo.User,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get token from header
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_token",
			"message": "Authentication token is required",
		})
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// Logout user
	err := h.authService.Logout(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "logout_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Successfully logged out",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_refresh_token",
			"message": "Refresh token is required",
		})
	}

	// Refresh token
	authInfo, err := h.authService.RefreshToken(c.Context(), input.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "refresh_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success":      true,
		"token":        authInfo.Token,
		"refreshToken": authInfo.RefreshToken,
		"user":         authInfo.User,
	})
}

// RequestPasswordReset handles password reset requests
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_email",
			"message": "Email is required",
		})
	}

	// Request password reset
	resetToken, err := h.authService.RequestPasswordReset(c.Context(), input.Email)
	if err != nil {
		// Don't reveal whether the email exists or not for security reasons
		return c.JSON(fiber.Map{
			"success": true,
			"message": "If your email is registered, you will receive a password reset link",
		})
	}

	// In a real app, we would send an email with the reset link
	// For development, we can return the token directly
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password reset link sent",
		"token":   resetToken, // Only include in development
	})
}

// CompletePasswordReset handles password reset completion
func (h *AuthHandler) CompletePasswordReset(c *fiber.Ctx) error {
	// Parse input
	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Validate input
	if input.Token == "" || input.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_required_fields",
			"message": "Token and new password are required",
		})
	}

	// Complete password reset
	err := h.authService.CompletePasswordReset(c.Context(), input.Token, input.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "reset_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password reset successfully",
	})
}

// InitiateOAuthLogin initiates OAuth login
func (h *AuthHandler) InitiateOAuthLogin(c *fiber.Ctx) error {
	// Get provider from URL parameter
	provider := c.Params("provider")
	if provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_provider",
			"message": "OAuth provider is required",
		})
	}

	// Initiate OAuth login
	authURL, _, err := h.oauthService.InitiateOAuth(c.Context(), repository.ProviderType(provider))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "oauth_initiation_failed",
			"message": err.Error(),
		})
	}

	// Store state in session (in a real app)
	// ...

	// Redirect to OAuth provider
	return c.Redirect(authURL)
}

// CompleteOAuthLogin completes OAuth login
func (h *AuthHandler) CompleteOAuthLogin(c *fiber.Ctx) error {
	// Get provider from URL parameter
	provider := c.Params("provider")
	if provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_provider",
			"message": "OAuth provider is required",
		})
	}

	// Get code and state from query parameters
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_parameters",
			"message": "Code and state are required",
		})
	}

	// Verify state (in a real app)
	// ...

	// Complete OAuth login
	authInfo, err := h.oauthService.CompleteOAuth(c.Context(), repository.ProviderType(provider), code, state)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "oauth_completion_failed",
			"message": err.Error(),
		})
	}

	// Return response or redirect to frontend with token
	frontendURL := c.Query("redirect_uri", "/")
	redirectURL := fmt.Sprintf("%s?token=%s&refresh_token=%s",
		frontendURL,
		authInfo.Token,
		authInfo.RefreshToken,
	)

	return c.Redirect(redirectURL)
}

// LinkOAuthAccount links an OAuth account
func (h *AuthHandler) LinkOAuthAccount(c *fiber.Ctx) error {
	// Get user ID from token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "unauthorized",
			"message": "You must be logged in to link an account",
		})
	}

	// Get provider from URL parameter
	provider := c.Params("provider")
	if provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_provider",
			"message": "OAuth provider is required",
		})
	}

	// Parse input
	var input struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid_input",
			"message": "Invalid input format",
		})
	}

	// Link account
	err = h.oauthService.LinkUserAccount(c.Context(), userID, repository.ProviderType(provider), input.Code, input.State)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "link_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Account linked successfully",
	})
}

// UnlinkOAuthAccount unlinks an OAuth account
func (h *AuthHandler) UnlinkOAuthAccount(c *fiber.Ctx) error {
	// Get user ID from token
	userID, err := h.getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "unauthorized",
			"message": "You must be logged in to unlink an account",
		})
	}

	// Get provider from URL parameter
	provider := c.Params("provider")
	if provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "missing_provider",
			"message": "OAuth provider is required",
		})
	}

	// Unlink account
	err = h.oauthService.UnlinkUserAccount(c.Context(), userID, repository.ProviderType(provider))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "unlink_failed",
			"message": err.Error(),
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Account unlinked successfully",
	})
}

// Helper to get user ID from token
func (h *AuthHandler) getUserIDFromToken(c *fiber.Ctx) (string, error) {
	// Get token from header
	token := c.Get("Authorization")
	if token == "" {
		return "", errors.New("missing token")
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// Validate token
	userID, err := h.authService.ValidateToken(c.Context(), token)
	if err != nil {
		return "", err
	}

	return userID, nil
}
