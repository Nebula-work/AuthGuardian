package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Common errors
var (
	// ErrInvalidToken is now defined in auth_service.go
	ErrExpiredToken  = errors.New("expired token")
	ErrTokenNotFound = errors.New("token not found")
)

// TokenServiceImpl implements TokenService
type TokenServiceImpl struct {
	jwtSecret           string
	accessTokenDuration time.Duration
	tokenRepository     repository.TokenRepository
}

// NewTokenService creates a new token service
func NewTokenService(jwtSecret string, accessTokenDuration time.Duration) repository.TokenService {
	// Use in-memory repository as fallback
	return NewTokenServiceWithRepository(jwtSecret, accessTokenDuration, nil)
}

// NewTokenServiceWithRepository creates a new token service with a specific repository
func NewTokenServiceWithRepository(jwtSecret string, accessTokenDuration time.Duration, repository repository.TokenRepository) repository.TokenService {
	// TODO Create in-memory repository if none provided

	return &TokenServiceImpl{
		jwtSecret:           jwtSecret,
		accessTokenDuration: accessTokenDuration,
		tokenRepository:     repository,
	}
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
func (s *TokenServiceImpl) ValidateToken(ctx context.Context, tokenString string) (models.TokenClaims, error) {
	// Check if token is revoked
	_, err := s.tokenRepository.FindTokenByValue(ctx, models.TokenTypeRevoked, tokenString)
	if err == nil {
		// Token is found in revoked tokens
		return models.TokenClaims{}, ErrInvalidToken
	}
	// If error is not ErrTokenNotFound, then just continue
	// We only care if token is explicitly revoked

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
			return models.TokenClaims{}, ErrExpiredToken
		}
		return models.TokenClaims{}, ErrInvalidToken
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Extract user ID
		userID, ok := claims["userId"].(string)
		if !ok {
			return models.TokenClaims{}, ErrInvalidToken
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

		return models.TokenClaims{
			UserID:    userID,
			Username:  username,
			Email:     email,
			RoleIDs:   roleIDs,
			IssuedAt:  int64(issuedAt),
			ExpiresAt: int64(expiresAt),
		}, nil
	}

	return models.TokenClaims{}, ErrInvalidToken
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
	tokenValue := base64.URLEncoding.EncodeToString(b)
	// Store token using repository
	token := models.Token{
		TokenType:  models.TokenTypeRefresh,
		TokenValue: tokenValue,
		UserID:     userID,
		CreatedAt:  time.Now().Format(time.RFC3339),
		// Set ExpiresAt if needed
	}

	// Store token
	err = s.tokenRepository.StoreToken(ctx, token)
	if err != nil {
		return "", err
	}

	return tokenValue, nil
}

// ValidateRefreshToken validates a refresh token
func (s *TokenServiceImpl) ValidateRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Get token from reposityr
	token, err := s.tokenRepository.FindTokenByValue(ctx, models.TokenTypeRefresh, refreshToken)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return "", ErrTokenNotFound
		}
		if errors.Is(err, ErrExpiredToken) {
			return "", ErrExpiredToken
		}
	}

	return token.UserID, nil
}

// RevokeToken revokes a token
func (s *TokenServiceImpl) RevokeToken(ctx context.Context, tokenValue string) error {
	token := models.Token{
		TokenType:  models.TokenTypeRevoked,
		TokenValue: tokenValue,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	if err := s.tokenRepository.StoreToken(ctx, token); err != nil {
		return err
	}

	// Remove from refresh tokens if it's a refresh token
	if err := s.tokenRepository.DeleteToken(ctx, models.TokenTypeRefresh, tokenValue); err != nil {
		// Log the error but continue
		fmt.Printf("Error deleting refresh token: %v\n", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a user
func (s *TokenServiceImpl) RevokeAllUserTokens(ctx context.Context, userID string) error {
	// Get all refresh tokens for the user
	refreshTokens, err := s.tokenRepository.FindTokensByUser(ctx, models.TokenTypeRefresh, userID)
	if err != nil {
		return err
	}

	// Revoke all refresh tokens
	for _, token := range refreshTokens {
		// Store as revoked
		revokedToken := models.Token{
			TokenType:  models.TokenTypeRevoked,
			TokenValue: token.TokenValue,
			CreatedAt:  time.Now().Format(time.RFC3339),
		}
		if err := s.tokenRepository.StoreToken(ctx, revokedToken); err != nil {
			// Log the error but continue
			fmt.Printf("Error storing revoked token: %v\n", err)
		}
	}

	// Delete all refresh tokens for the user
	return s.tokenRepository.DeleteTokensByUser(ctx, models.TokenTypeRefresh, userID)
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
	tokenValue := base64.URLEncoding.EncodeToString(b)

	// Calculate expiration (24 hours)
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	// Store token
	token := models.Token{
		UserID:     userID,
		TokenType:  models.TokenTypeReset,
		TokenValue: tokenValue,
		CreatedAt:  time.Now().Format(time.RFC3339),
		ExpiresAt:  expiresAt,
	}

	err = s.tokenRepository.StoreToken(ctx, token)
	if err != nil {
		return "", err
	}

	return tokenValue, nil
}

// ValidatePasswordResetToken validates a password reset token
func (s *TokenServiceImpl) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	// Get token from repository
	restoken, err := s.tokenRepository.FindTokenByValue(ctx, models.TokenTypeReset, token)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return "", ErrInvalidToken
		}
		if errors.Is(err, ErrExpiredToken) {
			return "", ErrExpiredToken
		}
		return "", err
	}

	// Remove token (one-time use)
	err = s.tokenRepository.DeleteToken(ctx, models.TokenTypeReset, token)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error deleting reset token: %v\n", err)
	}

	return restoken.UserID, nil
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
	tokenValue := base64.URLEncoding.EncodeToString(b)

	// Calculate expiration (7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)

	// Store token
	token := models.Token{
		UserID:     userID,
		TokenType:  models.TokenTypeVerification,
		TokenValue: tokenValue,
		CreatedAt:  time.Now().Format(time.RFC3339),
		ExpiresAt:  expiresAt,
	}

	err = s.tokenRepository.StoreToken(ctx, token)
	if err != nil {
		return "", err
	}

	return tokenValue, nil
}

// ValidateEmailVerificationToken validates an email verification token
func (s *TokenServiceImpl) ValidateEmailVerificationToken(ctx context.Context, token string) (string, error) {
	// Get token from repository
	tokenValue, err := s.tokenRepository.FindTokenByValue(ctx, models.TokenTypeVerification, token)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return "", ErrInvalidToken
		}
		if errors.Is(err, ErrExpiredToken) {
			return "", ErrExpiredToken
		}
		return "", err
	}

	// Remove token (one-time use)
	err = s.tokenRepository.DeleteToken(ctx, models.TokenTypeVerification, token)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error deleting verification token: %v\n", err)
	}

	return tokenValue.UserID, nil
}
