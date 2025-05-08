package identity

import (
	"context"
	"errors"
	"rbac-system/core/password"
	"rbac-system/pkg/common/repository"
	"time"
)

// Common errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidInput          = errors.New("invalid input")
)

// UserServiceImpl implements UserService
type UserServiceImpl struct {
	userRepo        UserRepository
	passwordManager password.PasswordManager
}

// NewUserService creates a new user service
func NewUserService(userRepo UserRepository, passwordManager password.PasswordManager) UserService {
	return &UserServiceImpl{
		userRepo:        userRepo,
		passwordManager: passwordManager,
	}
}

// CreateUser creates a new user
func (s *UserServiceImpl) CreateUser(ctx context.Context, user User) (string, error) {
	// Validate user
	if user.Username == "" || user.Email == "" {
		return "", ErrInvalidInput
	}

	// Check if email already exists
	_, err := s.userRepo.FindByEmail(ctx, user.Email)
	if err == nil {
		return "", ErrEmailAlreadyExists
	}

	// Check if username already exists
	_, err = s.userRepo.FindByUsername(ctx, user.Username)
	if err == nil {
		return "", ErrUsernameAlreadyExists
	}

	// Set default values
	now := time.Now()
	timeStr := now.Format(time.RFC3339)
	user.CreatedAt = timeStr
	user.UpdatedAt = timeStr
	if user.Active == false {
		user.Active = true
	}

	// Create user
	return s.userRepo.Create(ctx, user)
}

// UpdateUser updates an existing user
func (s *UserServiceImpl) UpdateUser(ctx context.Context, id string, user User) error {
	// Get existing user
	existingUser, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if email changed and already exists
	if user.Email != existingUser.Email {
		_, err := s.userRepo.FindByEmail(ctx, user.Email)
		if err == nil {
			return ErrEmailAlreadyExists
		}
	}

	// Check if username changed and already exists
	if user.Username != existingUser.Username {
		_, err := s.userRepo.FindByUsername(ctx, user.Username)
		if err == nil {
			return ErrUsernameAlreadyExists
		}
	}

	// Preserve fields that shouldn't be updated
	user.ID = id
	user.PasswordHash = existingUser.PasswordHash // Don't update password through this method
	user.CreatedAt = existingUser.CreatedAt
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	// Update user
	return s.userRepo.Update(ctx, id, user)
}

// DeleteUser deletes a user
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id string) error {
	// Get user
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Delete user
	return s.userRepo.Delete(ctx, id)
}

// GetUser retrieves a user by ID
func (s *UserServiceImpl) GetUser(ctx context.Context, id string) (User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// GetUsers retrieves all users with optional filtering
func (s *UserServiceImpl) GetUsers(ctx context.Context, orgID string, skip, limit int64) ([]User, int64, error) {
	// Create filter
	filter := make(map[string]interface{})
	if orgID != "" {
		filter["organizationIds"] = orgID
	}

	// Create options
	options := repository.QueryOptions{
		Skip:  skip,
		Limit: limit,
		Sort:  map[string]int{"username": 1},
	}

	// Get users
	users, err := s.userRepo.FindMany(ctx, filter, options)
	if err != nil {
		return nil, 0, err
	}

	// Get count
	count, err := s.userRepo.Count(ctx, filter)
	if err != nil {
		return users, 0, err
	}

	return users, count, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserServiceImpl) GetUserByUsername(ctx context.Context, username string) (User, error) {
	return s.userRepo.FindByUsername(ctx, username)
}

// GetUserByEmail retrieves a user by email
func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// GetUserByOAuthID retrieves a user by OAuth provider and ID
func (s *UserServiceImpl) GetUserByOAuthID(ctx context.Context, provider, providerUserID string) (User, error) {
	return s.userRepo.FindByOAuthID(ctx, provider, providerUserID)
}

// UpdateLastLogin updates a user's last login time
func (s *UserServiceImpl) UpdateLastLogin(ctx context.Context, id string) error {
	return s.userRepo.UpdateLastLogin(ctx, id)
}

// SetPassword sets a user's password
func (s *UserServiceImpl) SetPassword(ctx context.Context, id, password string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Validate password strength
	err = s.passwordManager.ValidatePasswordStrength(password)
	if err != nil {
		return err
	}

	// Hash password
	passwordHash, err := s.passwordManager.HashPassword(password)
	if err != nil {
		return err
	}

	// Update user
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	return s.userRepo.Update(ctx, id, user)
}

// VerifyPassword verifies a user's password
func (s *UserServiceImpl) VerifyPassword(ctx context.Context, id, password string) (bool, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return false, ErrUserNotFound
	}

	// Verify password
	err = s.passwordManager.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// VerifyEmail marks a user's email as verified
func (s *UserServiceImpl) VerifyEmail(ctx context.Context, id string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Update user
	user.EmailVerified = true
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	return s.userRepo.Update(ctx, id, user)
}

// AddRoleToUser adds a role to a user
func (s *UserServiceImpl) AddRoleToUser(ctx context.Context, userID, roleID string) error {
	return s.userRepo.AddRoleToUser(ctx, userID, roleID)
}

// RemoveRoleFromUser removes a role from a user
func (s *UserServiceImpl) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	return s.userRepo.RemoveRoleFromUser(ctx, userID, roleID)
}

// AddOrganizationToUser adds an organization to a user
func (s *UserServiceImpl) AddOrganizationToUser(ctx context.Context, userID, organizationID string) error {
	return s.userRepo.AddOrganizationToUser(ctx, userID, organizationID)
}

// RemoveOrganizationFromUser removes an organization from a user
func (s *UserServiceImpl) RemoveOrganizationFromUser(ctx context.Context, userID, organizationID string) error {
	return s.userRepo.RemoveOrganizationFromUser(ctx, userID, organizationID)
}

// GetUserRoles retrieves roles for a user
func (s *UserServiceImpl) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user.RoleIDs, nil
}

// GetUserOrganizations retrieves organizations for a user
func (s *UserServiceImpl) GetUserOrganizations(ctx context.Context, userID string) ([]string, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user.OrganizationIDs, nil
}

// LinkOAuthAccount links an OAuth account to a user
func (s *UserServiceImpl) LinkOAuthAccount(ctx context.Context, userID, provider, providerUserID, refreshToken string) error {
	return s.userRepo.LinkOAuthAccount(ctx, userID, provider, providerUserID, refreshToken)
}

// UnlinkOAuthAccount unlinks an OAuth account from a user
func (s *UserServiceImpl) UnlinkOAuthAccount(ctx context.Context, userID, provider string) error {
	return s.userRepo.UnlinkOAuthAccount(ctx, userID, provider)
}
