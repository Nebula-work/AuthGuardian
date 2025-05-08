package auth

import (
	"context"
	"errors"
	"rbac-system/core/identity"
	"rbac-system/core/models"
	"rbac-system/core/password"
	"time"
)

// Common errors
var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserInactive          = errors.New("user account is inactive")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidToken          = errors.New("invalid token")
)

// AuthServiceImpl implements AuthService
type AuthServiceImpl struct {
	userRepo        identity.UserRepository
	tokenService    TokenService
	passwordManager password.PasswordManager
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo identity.UserRepository, tokenService TokenService, passwordManager password.PasswordManager) AuthService {
	return &AuthServiceImpl{
		userRepo:        userRepo,
		tokenService:    tokenService,
		passwordManager: passwordManager,
	}
}

// Login authenticates a user
func (s *AuthServiceImpl) Login(ctx context.Context, username, password string) (AuthInfo, error) {
	// Get user by username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		// Try by email
		user, err = s.userRepo.FindByEmail(ctx, username)
		if err != nil {
			return AuthInfo{}, ErrInvalidCredentials
		}
	}

	// Check if user is active
	if !user.Active {
		return AuthInfo{}, ErrUserInactive
	}

	// Verify password
	err = s.passwordManager.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return AuthInfo{}, ErrInvalidCredentials
	}

	// Generate JWT token
	claims := TokenClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		RoleIDs:   user.RoleIDs,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	token, err := s.tokenService.GenerateToken(ctx, claims)
	if err != nil {
		return AuthInfo{}, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return AuthInfo{}, err
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, user.User.ID)
	if err != nil {
		// Non-critical error, continue
	}

	return AuthInfo{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Register creates a new user
func (s *AuthServiceImpl) Register(ctx context.Context, request RegisterRequest) (AuthInfo, error) {
	// Check if email already exists
	_, err := s.userRepo.FindByEmail(ctx, request.Email)
	if err == nil {
		return AuthInfo{}, ErrEmailAlreadyExists
	}

	// Check if username already exists
	_, err = s.userRepo.FindByUsername(ctx, request.Username)
	if err == nil {
		return AuthInfo{}, ErrUsernameAlreadyExists
	}

	// Hash password
	passwordHash, err := s.passwordManager.HashPassword(request.Password)
	if err != nil {
		return AuthInfo{}, err
	}

	// Create user
	now := time.Now()
	createdAt := now.Format(time.RFC3339)
	updatedAt := createdAt

	user := identity.User{
		User: models.User{
			Username:      request.Username,
			Email:         request.Email,
			PasswordHash:  passwordHash,
			FirstName:     request.FirstName,
			LastName:      request.LastName,
			Active:        true,
			EmailVerified: false,
			AuthProvider:  "local",
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		},
	}

	// Save user
	userID, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return AuthInfo{}, err
	}

	// Get created user
	createdUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return AuthInfo{}, err
	}

	// Generate JWT token
	claims := TokenClaims{
		UserID:    createdUser.User.ID,
		Username:  createdUser.User.Username,
		Email:     createdUser.User.Email,
		RoleIDs:   createdUser.User.RoleIDs,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	token, err := s.tokenService.GenerateToken(ctx, claims)
	if err != nil {
		return AuthInfo{}, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(ctx, createdUser.User.ID)
	if err != nil {
		return AuthInfo{}, err
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, createdUser.User.ID)
	if err != nil {
		// Non-critical error, continue
	}

	return AuthInfo{
		Token:        token,
		RefreshToken: refreshToken,
		User:         createdUser,
	}, nil
}

// ValidateToken validates a token
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (string, error) {
	// Validate token
	claims, err := s.tokenService.ValidateToken(ctx, token)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}

// RefreshToken refreshes a token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (AuthInfo, error) {
	// Validate refresh token
	userID, err := s.tokenService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return AuthInfo{}, err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return AuthInfo{}, ErrUserNotFound
	}

	// Check if user is active
	if !user.Active {
		return AuthInfo{}, ErrUserInactive
	}

	// Generate JWT token
	claims := TokenClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		RoleIDs:   user.RoleIDs,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	token, err := s.tokenService.GenerateToken(ctx, claims)
	if err != nil {
		return AuthInfo{}, err
	}

	// Generate new refresh token
	newRefreshToken, err := s.tokenService.GenerateRefreshToken(ctx, user.ID)
	if err != nil {
		return AuthInfo{}, err
	}

	// Revoke old refresh token
	err = s.tokenService.RevokeToken(ctx, refreshToken)
	if err != nil {
		// Non-critical error, continue
	}

	return AuthInfo{
		Token:        token,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

// Logout logs out a user
func (s *AuthServiceImpl) Logout(ctx context.Context, token string) error {
	// Revoke token
	return s.tokenService.RevokeToken(ctx, token)
}

// RequestPasswordReset initiates a password reset
func (s *AuthServiceImpl) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	// Get user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Generate password reset token
	resetToken, err := s.tokenService.GeneratePasswordResetToken(ctx, user.ID)
	if err != nil {
		return "", err
	}

	// In a real implementation, we would send an email with the reset token

	return resetToken, nil
}

// CompletePasswordReset completes a password reset
func (s *AuthServiceImpl) CompletePasswordReset(ctx context.Context, token, newPassword string) error {
	// Validate password reset token
	userID, err := s.tokenService.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Hash new password
	passwordHash, err := s.passwordManager.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Update user password
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	// Save user
	err = s.userRepo.Update(ctx, userID, user)
	if err != nil {
		return err
	}

	// Revoke all tokens for the user
	err = s.tokenService.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		// Non-critical error, continue
	}

	return nil
}

// VerifyEmail verifies a user's email
func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	// Validate email verification token
	userID, err := s.tokenService.ValidateEmailVerificationToken(ctx, token)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Update user
	user.EmailVerified = true
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	// Save user
	return s.userRepo.Update(ctx, userID, user)
}
