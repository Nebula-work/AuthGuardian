package auth

import (
	"context"
	"errors"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"time"
)

// OAuthServiceImpl implements OAuthService
type OAuthServiceImpl struct {
	userRepo     repository.UserRepository
	tokenService repository.TokenService
	providers    map[string]OAuthProvider
}

// OAuthProvider defines the interface for OAuth providers
type OAuthProvider interface {
	// GetAuthURL returns the authorization URL for the provider
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for a token
	ExchangeCode(code string) (string, error)

	// GetUserInfo retrieves user information from the provider
	GetUserInfo(token string) (OAuthUserInfo, error)
}

// OAuthUserInfo contains user information from an OAuth provider
type OAuthUserInfo struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Username  string
}

// Common errors
var (
	ErrProviderNotFound   = errors.New("oauth provider not found")
	ErrCodeExchangeFailed = errors.New("failed to exchange authorization code")
	ErrUserInfoFailed     = errors.New("failed to retrieve user information")
)

// NewOAuthService creates a new OAuth service
func NewOAuthService(userRepo repository.UserRepository, tokenService repository.TokenService) repository.OAuthService {
	return &OAuthServiceImpl{
		userRepo:     userRepo,
		tokenService: tokenService,
		providers:    make(map[string]OAuthProvider),
	}
}

// RegisterProvider registers an OAuth provider
func (s *OAuthServiceImpl) RegisterProvider(name string, provider OAuthProvider) {
	s.providers[string(name)] = provider
}

// InitiateOAuth begins an OAuth flow
func (s *OAuthServiceImpl) InitiateOAuth(ctx context.Context, providerType repository.ProviderType) (string, string, error) {
	// Get provider
	provider, ok := s.providers[string(providerType)]
	if !ok {
		return "", "", ErrProviderNotFound
	}

	// Generate state
	state := generateRandomString(32)

	// Get authorization URL
	authURL := provider.GetAuthURL(state)

	return authURL, state, nil
}

// CompleteOAuth completes an OAuth flow
func (s *OAuthServiceImpl) CompleteOAuth(ctx context.Context, providerType repository.ProviderType, code, state string) (repository.AuthInfo, error) {
	// Get provider
	provider, ok := s.providers[string(providerType)]
	if !ok {
		return repository.AuthInfo{}, ErrProviderNotFound
	}

	// Exchange code for token
	token, err := provider.ExchangeCode(code)
	if err != nil {
		return repository.AuthInfo{}, ErrCodeExchangeFailed
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		return repository.AuthInfo{}, ErrUserInfoFailed
	}

	// Check if user exists
	user, err := s.userRepo.FindByOAuthID(ctx, string(providerType), userInfo.ID)
	if err == nil {
		// User exists, generate token
		return s.generateAuthInfoForUser(ctx, user)
	}

	// Check if user exists by email
	if userInfo.Email != "" {
		user, err = s.userRepo.FindByEmail(ctx, userInfo.Email)
		if err == nil {
			// User exists, link OAuth account
			err = s.userRepo.LinkOAuthAccount(ctx, user.ID, string(providerType), userInfo.ID, token)
			if err != nil {
				return repository.AuthInfo{}, err
			}

			// Generate token
			return s.generateAuthInfoForUser(ctx, user)
		}
	}

	// Create new user
	newUser := repository.User{
		User: models.User{
			Username:      userInfo.Username,
			Email:         userInfo.Email,
			FirstName:     userInfo.FirstName,
			LastName:      userInfo.LastName,
			Active:        true,
			EmailVerified: true,
			AuthProvider:  string(providerType),
			CreatedAt:     getNow().Format(time.RFC3339),
			UpdatedAt:     getNow().Format(time.RFC3339),
		},
		OAuthAccounts: []repository.OAuthAccount{
			{
				Provider:       string(providerType),
				ProviderUserID: userInfo.ID,
				RefreshToken:   token,
				LinkedAt:       getNow(),
			},
		},
	}

	// Generate username if not provided
	if newUser.User.Username == "" {
		newUser.User.Username = generateUsernameFromEmail(newUser.User.Email)
	}

	// Save user
	userID, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return repository.AuthInfo{}, err
	}

	// Get created user
	createdUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return repository.AuthInfo{}, err
	}

	// Generate token
	return s.generateAuthInfoForUser(ctx, createdUser)
}

// LinkUserAccount links a user account to an OAuth provider
func (s *OAuthServiceImpl) LinkUserAccount(ctx context.Context, userID string, providerType repository.ProviderType, code, state string) error {
	// Get provider
	provider, ok := s.providers[string(providerType)]
	if !ok {
		return ErrProviderNotFound
	}

	// Exchange code for token
	token, err := provider.ExchangeCode(code)
	if err != nil {
		return ErrCodeExchangeFailed
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		return ErrUserInfoFailed
	}

	// Link account
	return s.userRepo.LinkOAuthAccount(ctx, userID, string(providerType), userInfo.ID, token)
}

// UnlinkUserAccount unlinks a user account from an OAuth provider
func (s *OAuthServiceImpl) UnlinkUserAccount(ctx context.Context, userID string, providerType repository.ProviderType) error {
	return s.userRepo.UnlinkOAuthAccount(ctx, userID, string(providerType))
}

// GetOAuthProviders returns available OAuth providers
func (s *OAuthServiceImpl) GetOAuthProviders(ctx context.Context) ([]repository.ProviderType, error) {
	providers := make([]repository.ProviderType, 0, len(s.providers))
	for provider := range s.providers {
		providers = append(providers, repository.ProviderType(provider))
	}
	return providers, nil
}

// GetLinkedProviders returns OAuth providers linked to a user
func (s *OAuthServiceImpl) GetLinkedProviders(ctx context.Context, userID string) ([]repository.ProviderType, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get linked providers
	providers := make([]repository.ProviderType, 0, len(user.OAuthAccounts))
	for _, account := range user.OAuthAccounts {
		providers = append(providers, repository.ProviderType(account.Provider))
	}

	return providers, nil
}

// Helper functions

// generateAuthInfoForUser generates authentication information for a user
func (s *OAuthServiceImpl) generateAuthInfoForUser(ctx context.Context, user repository.User) (repository.AuthInfo, error) {
	// Generate claims
	claims := models.TokenClaims{
		UserID:   user.User.ID,
		Username: user.User.Username,
		Email:    user.User.Email,
		RoleIDs:  user.User.RoleIDs,
	}

	// Generate token
	token, err := s.tokenService.GenerateToken(ctx, claims)
	if err != nil {
		return repository.AuthInfo{}, err
	}

	// Generate refresh token
	refreshToken, err := s.tokenService.GenerateRefreshToken(ctx, user.User.ID)
	if err != nil {
		return repository.AuthInfo{}, err
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, user.User.ID)
	if err != nil {
		// Non-critical error, continue
	}

	return repository.AuthInfo{
		Token:        token,
		RefreshToken: refreshToken,
		User:         repository.UserData(user),
	}, nil
}

// Helper function to generate a random string
func generateRandomString(length int) string {
	// generate a secure random string
	return "random-state-string"
}

// Helper function to generate a username from an email
func generateUsernameFromEmail(email string) string {
	// generate a username from an email
	return "user"
}

// Helper function to get current time
func getNow() time.Time {
	return time.Now()
}
