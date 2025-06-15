package repository

import (
	"context"
	"rbac-system/core/models"
)

// AuthInfo contains authentication result information
type UserData struct {
	models.User
	OAuthAccounts []OAuthAccount `json:"-" bson:"oauthAccounts"`
}
type AuthInfo struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	User         UserData `json:"user"`
}

// RegisterRequest contains data needed for user registration
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// ProviderType represents an OAuth provider type
type ProviderType string

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Login authenticates a user with username and password
	Login(ctx context.Context, username, password string) (AuthInfo, error)

	// Register creates a new user account
	Register(ctx context.Context, request RegisterRequest) (AuthInfo, error)

	// ValidateToken validates a token and returns the user ID
	ValidateToken(ctx context.Context, token string) (string, error)

	// RefreshToken refreshes an access token
	RefreshToken(ctx context.Context, refreshToken string) (AuthInfo, error)

	// Logout logs out a user
	Logout(ctx context.Context, token string) error

	// RequestPasswordReset initiates a password reset
	RequestPasswordReset(ctx context.Context, email string) (string, error)

	// CompletePasswordReset completes a password reset
	CompletePasswordReset(ctx context.Context, token, newPassword string) error

	// VerifyEmail verifies a user's email address
	VerifyEmail(ctx context.Context, token string) error
}

// TokenService defines the interface for token operations
type TokenService interface {
	// GenerateToken generates a new token
	GenerateToken(ctx context.Context, claims models.TokenClaims) (string, error)

	// ValidateToken validates a token and returns the claims
	ValidateToken(ctx context.Context, token string) (models.TokenClaims, error)

	// GenerateRefreshToken generates a new refresh token
	GenerateRefreshToken(ctx context.Context, userID string) (string, error)

	// ValidateRefreshToken validates a refresh token
	ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error)

	// RevokeToken revokes a token
	RevokeToken(ctx context.Context, token string) error

	// RevokeAllUserTokens revokes all tokens for a user
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// GeneratePasswordResetToken generates a password reset token
	GeneratePasswordResetToken(ctx context.Context, userID string) (string, error)

	// ValidatePasswordResetToken validates a password reset token
	ValidatePasswordResetToken(ctx context.Context, token string) (string, error)

	// GenerateEmailVerificationToken generates an email verification token
	GenerateEmailVerificationToken(ctx context.Context, userID string) (string, error)

	// ValidateEmailVerificationToken validates an email verification token
	ValidateEmailVerificationToken(ctx context.Context, token string) (string, error)
}

// For storing tokens in db
type TokenRepository interface {
	//StoreToken stores a token in the database
	StoreToken(ctx context.Context, token models.Token) error

	// FindTokenByValue finds a token by its value
	FindTokenByValue(ctx context.Context, tokenType models.TokenType, tokenValue string) (models.Token, error)

	// FindTokensByUser finds all tokens for a user
	FindTokensByUser(ctx context.Context, tokenType models.TokenType, userID string) ([]models.Token, error)

	// DeleteToken deletes a token
	DeleteToken(ctx context.Context, tokenType models.TokenType, tokenValue string) error

	// DeleteTokensByUser deletes all tokens for a user
	DeleteTokensByUser(ctx context.Context, tokenType models.TokenType, userID string) error

	// DeleteExpiredTokens deletes all expired tokens
	DeleteExpiredTokens(ctx context.Context) error
}

// OAuthService defines the interface for OAuth operations
type OAuthService interface {
	// InitiateOAuth begins an OAuth flow
	InitiateOAuth(ctx context.Context, providerType ProviderType) (string, string, error)

	// CompleteOAuth completes an OAuth flow and returns auth info
	CompleteOAuth(ctx context.Context, providerType ProviderType, code, state string) (AuthInfo, error)

	// LinkUserAccount links a user account to an OAuth provider
	LinkUserAccount(ctx context.Context, userID string, providerType ProviderType, code, state string) error

	// UnlinkUserAccount unlinks a user account from an OAuth provider
	UnlinkUserAccount(ctx context.Context, userID string, providerType ProviderType) error

	// GetOAuthProviders returns available OAuth providers
	GetOAuthProviders(ctx context.Context) ([]ProviderType, error)

	// GetLinkedProviders returns OAuth providers linked to a user
	GetLinkedProviders(ctx context.Context, userID string) ([]ProviderType, error)
}
