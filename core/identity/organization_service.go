package identity

import (
	"context"
	"errors"
	"rbac-system/pkg/common/repository"
	"time"
)

// Common errors
var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrNameAlreadyExists    = errors.New("organization name already exists")
	ErrDomainAlreadyExists  = errors.New("organization domain already exists")
)

// OrganizationServiceImpl implements OrganizationService
type OrganizationServiceImpl struct {
	orgRepo  repository.OrganizationRepository
	userRepo repository.UserRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(orgRepo repository.OrganizationRepository, userRepo repository.UserRepository) OrganizationService {
	return &OrganizationServiceImpl{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationServiceImpl) CreateOrganization(ctx context.Context, org repository.Organization) (string, error) {
	// Validate organization
	if org.Name == "" {
		return "", ErrInvalidInput
	}

	// Check if name already exists
	_, err := s.orgRepo.FindByName(ctx, org.Name)
	if err == nil {
		return "", ErrNameAlreadyExists
	}

	// Check if domain already exists (if provided)
	if org.Domain != "" {
		_, err = s.orgRepo.FindByDomain(ctx, org.Domain)
		if err == nil {
			return "", ErrDomainAlreadyExists
		}
	}

	// Set default values
	now := time.Now()
	timeStr := now.Format(time.RFC3339)
	org.CreatedAt = timeStr
	org.UpdatedAt = timeStr
	if !org.Active {
		org.Active = true
	}

	// Create organization
	return s.orgRepo.Create(ctx, org)
}

// UpdateOrganization updates an existing organization
func (s *OrganizationServiceImpl) UpdateOrganization(ctx context.Context, id string, org repository.Organization) error {
	// Get existing organization
	existingOrg, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return ErrOrganizationNotFound
	}

	// Check if name changed and already exists
	if org.Name != existingOrg.Name {
		_, err := s.orgRepo.FindByName(ctx, org.Name)
		if err == nil {
			return ErrNameAlreadyExists
		}
	}

	// Check if domain changed and already exists
	if org.Domain != existingOrg.Domain && org.Domain != "" {
		_, err := s.orgRepo.FindByDomain(ctx, org.Domain)
		if err == nil {
			return ErrDomainAlreadyExists
		}
	}

	// Preserve fields that shouldn't be updated
	org.ID = id
	org.CreatedAt = existingOrg.CreatedAt
	org.UpdatedAt = time.Now().Format(time.RFC3339)

	// Update organization
	return s.orgRepo.Update(ctx, id, org)
}

// DeleteOrganization deletes an organization
func (s *OrganizationServiceImpl) DeleteOrganization(ctx context.Context, id string) error {
	// Get organization
	_, err := s.orgRepo.FindByID(ctx, id)
	if err != nil {
		return ErrOrganizationNotFound
	}

	// Delete organization
	return s.orgRepo.Delete(ctx, id)
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationServiceImpl) GetOrganization(ctx context.Context, id string) (repository.Organization, error) {
	return s.orgRepo.FindByID(ctx, id)
}

// GetOrganizations retrieves all organizations with optional filtering
func (s *OrganizationServiceImpl) GetOrganizations(ctx context.Context, skip, limit int64) ([]repository.Organization, int64, error) {
	// Create filter
	filter := make(map[string]interface{})

	// Create options
	options := repository.QueryOptions{
		Skip:  skip,
		Limit: limit,
		Sort:  map[string]int{"name": 1},
	}

	// Get organizations
	orgs, err := s.orgRepo.FindMany(ctx, filter, options)
	if err != nil {
		return nil, 0, err
	}

	// Get count
	count, err := s.orgRepo.Count(ctx, filter)
	if err != nil {
		return orgs, 0, err
	}

	return orgs, count, nil
}

// GetOrganizationByName retrieves an organization by name
func (s *OrganizationServiceImpl) GetOrganizationByName(ctx context.Context, name string) (repository.Organization, error) {
	return s.orgRepo.FindByName(ctx, name)
}

// GetOrganizationByDomain retrieves an organization by domain
func (s *OrganizationServiceImpl) GetOrganizationByDomain(ctx context.Context, domain string) (repository.Organization, error) {
	return s.orgRepo.FindByDomain(ctx, domain)
}

// GetUserOrganizations retrieves organizations for a user
func (s *OrganizationServiceImpl) GetUserOrganizations(ctx context.Context, userID string) ([]repository.Organization, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Get organizations
	orgs := make([]repository.Organization, 0, len(user.OrganizationIDs))
	for _, orgID := range user.OrganizationIDs {
		org, err := s.orgRepo.FindByID(ctx, orgID)
		if err != nil {
			continue
		}

		orgs = append(orgs, org)
	}

	return orgs, nil
}

// AddUserToOrganization adds a user to an organization
func (s *OrganizationServiceImpl) AddUserToOrganization(ctx context.Context, orgID, userID string) error {
	// Get organization
	_, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}

	// Get user
	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Add organization to user
	return s.userRepo.AddOrganizationToUser(ctx, userID, orgID)
}

// RemoveUserFromOrganization removes a user from an organization
func (s *OrganizationServiceImpl) RemoveUserFromOrganization(ctx context.Context, orgID, userID string) error {
	// Get organization
	_, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}

	// Get user
	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Remove organization from user
	return s.userRepo.RemoveOrganizationFromUser(ctx, userID, orgID)
}

// AddAdminToOrganization adds an admin to an organization
func (s *OrganizationServiceImpl) AddAdminToOrganization(ctx context.Context, orgID, userID string) error {
	// Get organization
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}

	// Get user
	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if user is already an admin
	for _, id := range org.AdminIDs {
		if id == userID {
			return nil
		}
	}

	// Add user to organization if not already a member
	err = s.AddUserToOrganization(ctx, orgID, userID)
	if err != nil {
		// Ignore error if user is already a member
	}

	// Add admin to organization
	return s.orgRepo.AddAdminToOrganization(ctx, orgID, userID)
}

// RemoveAdminFromOrganization removes an admin from an organization
func (s *OrganizationServiceImpl) RemoveAdminFromOrganization(ctx context.Context, orgID, userID string) error {
	// Get organization
	_, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return ErrOrganizationNotFound
	}

	// Get user
	_, err = s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Remove admin from organization
	return s.orgRepo.RemoveAdminFromOrganization(ctx, orgID, userID)
}

// IsUserAdmin checks if a user is an admin of an organization
func (s *OrganizationServiceImpl) IsUserAdmin(ctx context.Context, orgID, userID string) (bool, error) {
	return s.orgRepo.IsUserAdmin(ctx, orgID, userID)
}

// GetOrganizationUsers retrieves users in an organization
func (s *OrganizationServiceImpl) GetOrganizationUsers(ctx context.Context, orgID string, skip, limit int64) ([]repository.User, int64, error) {
	// Get organization
	_, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, 0, ErrOrganizationNotFound
	}

	// Create filter
	filter := map[string]interface{}{
		"organizationIds": orgID,
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

// GetOrganizationAdmins retrieves admins of an organization
func (s *OrganizationServiceImpl) GetOrganizationAdmins(ctx context.Context, orgID string) ([]repository.User, error) {
	// Get organization
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	// Get admins
	users := make([]repository.User, 0, len(org.AdminIDs))
	for _, userID := range org.AdminIDs {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			continue
		}

		users = append(users, user)
	}

	return users, nil
}
