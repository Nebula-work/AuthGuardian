package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"rbac-system/core/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Common errors
var (
	// ErrInvalidToken is now defined in auth_service.go
	ErrExpiredToken     = errors.New("expired token")
	ErrTokenNotFound    = errors.New("token not found")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// TokenServiceImpl implements TokenService
type TokenServiceImpl struct {
	jwtSecret           string
	accessTokenDuration time.Duration
	tokenRepository     TokenRepository
}

// NewTokenService creates a new token service
func NewTokenService(jwtSecret string, accessTokenDuration time.Duration) TokenService {
	// Use in-memory repository as fallback
	return NewTokenServiceWithRepository(jwtSecret, accessTokenDuration, nil)
}

// NewTokenServiceWithRepository creates a new token service with a specific repository
func NewTokenServiceWithRepository(jwtSecret string, accessTokenDuration time.Duration, repository TokenRepository) TokenService {
	// Create in-memory repository if none provided
	if repository == nil {
		repository = &inMemoryTokenRepository{
			refreshTokens:      make(map[string]string),
			revokedTokens:      make(map[string]bool),
			resetTokens:        make(map[string]string),
			verificationTokens: make(map[string]string),
		}
	}

	return &TokenServiceImpl{
		jwtSecret:           jwtSecret,
		accessTokenDuration: accessTokenDuration,
		tokenRepository:     repository,
	}
}

// inMemoryTokenRepository provides legacy in-memory token storage
// This is maintained for backwards compatibility until the transition to the new TokenRepository is complete
type inMemoryTokenRepository struct {
	refreshTokens      map[string]string
	revokedTokens      map[string]bool
	resetTokens        map[string]string
	verificationTokens map[string]string
}

// GenerateToken generates a new JWT token
func (s *TokenServiceImpl) GenerateToken(ctx context.Context, claims models.TokenClaims) (string, error) {
	// Set expiration if not set
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = time.Now().Add(s.accessTokenDuration).Unix()
	}

	// Set issued at if not set
	if claims.IssuedAt == 0 {
		claims.IssuedAt = time.Now().Unix()
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   claims.UserID,
		"username": claims.Username,
		"email":    claims.Email,
		"roleIds":  claims.RoleIDs,
		"iat":      claims.IssuedAt,
		"exp":      claims.ExpiresAt,
	})

	// Sign token
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *TokenServiceImpl) ValidateToken(ctx context.Context, tokenString string) (TokenClaims, error) {
	// Check if token is revoked
	if s.revokedTokens[tokenString] {
		return TokenClaims{}, ErrInvalidToken
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return secret key
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return TokenClaims{}, ErrExpiredToken
		}
		return TokenClaims{}, ErrInvalidToken
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract user ID
		userID, ok := claims["userId"].(string)
		if !ok {
			return TokenClaims{}, ErrInvalidToken
		}

		// Extract username
		username, _ := claims["username"].(string)

		// Extract email
		email, _ := claims["email"].(string)

		// Extract role IDs
		roleIDs := []string{}
		if roleIDsRaw, ok := claims["roleIds"].([]interface{}); ok {
			for _, roleID := range roleIDsRaw {
				if roleIDStr, ok := roleID.(string); ok {
					roleIDs = append(roleIDs, roleIDStr)
				}
			}
		}

		// Extract issued at
		issuedAt, _ := claims["iat"].(float64)

		// Extract expiration
		expiresAt, _ := claims["exp"].(float64)

		return TokenClaims{
			UserID:    userID,
			Username:  username,
			Email:     email,
			RoleIDs:   roleIDs,
			IssuedAt:  int64(issuedAt),
			ExpiresAt: int64(expiresAt),
		}, nil
	}

	return TokenClaims{}, ErrInvalidToken
}

// GenerateRefreshToken generates a new refresh token
func (s *TokenServiceImpl) GenerateRefreshToken(ctx context.Context, userID string) (string, error) {
	// Generate random token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode token
	token := base64.URLEncoding.EncodeToString(b)

	// Store token
	s.refreshTokens[token] = userID

	return token, nil
}

// ValidateRefreshToken validates a refresh token
func (s *TokenServiceImpl) ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Get user ID from token
	userID, ok := s.refreshTokens[refreshToken]
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// RevokeToken revokes a token
func (s *TokenServiceImpl) RevokeToken(ctx context.Context, token string) error {
	// Mark token as revoked
	s.revokedTokens[token] = true

	// Remove from refresh tokens if it's a refresh token
	delete(s.refreshTokens, token)

	return nil
}

// RevokeAllUserTokens revokes all tokens for a user
func (s *TokenServiceImpl) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// Remove all refresh tokens for the user
	for token, tokenUserID := range s.refreshTokens {
		if tokenUserID == userID {
			delete(s.refreshTokens, token)
		}
	}

	// In a real implementation, we would need to keep track of
	// all access tokens for a user and mark them as revoked

	return nil
}

// GeneratePasswordResetToken generates a password reset token
func (s *TokenServiceImpl) GeneratePasswordResetToken(ctx context.Context, userID string) (string, error) {
	// Generate random token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode token
	token := base64.URLEncoding.EncodeToString(b)

	// Store token
	s.resetTokens[token] = userID

	return token, nil
}

// ValidatePasswordResetToken validates a password reset token
func (s *TokenServiceImpl) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	// Get user ID from token
	userID, ok := s.resetTokens[token]
	if !ok {
		return "", ErrInvalidToken
	}

	// Remove token (one-time use)
	delete(s.resetTokens, token)

	return userID, nil
}

// GenerateEmailVerificationToken generates an email verification token
func (s *TokenServiceImpl) GenerateEmailVerificationToken(ctx context.Context, userID string) (string, error) {
	// Generate random token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode token
	token := base64.URLEncoding.EncodeToString(b)

	// Store token
	s.verificationTokens[token] = userID

	return token, nil
}

// ValidateEmailVerificationToken validates an email verification token
func (s *TokenServiceImpl) ValidateEmailVerificationToken(ctx context.Context, token string) (string, error) {
	// Get user ID from token
	userID, ok := s.verificationTokens[token]
	if !ok {
		return "", ErrInvalidToken
	}

	// Remove token (one-time use)
	delete(s.verificationTokens, token)

	return userID, nil
}
