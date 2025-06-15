package mongodb

import (
	"context"
	"errors"
	"rbac-system/core/models"
	"rbac-system/core/rbac"
	"rbac-system/pkg/common/repository"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryUserRepository implements identity.UserRepository
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]repository.User
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() repository.UserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]repository.User),
	}
}

// IsConnected checks if the repository is connected
func (r *InMemoryUserRepository) IsConnected(ctx context.Context) bool {
	return true
}

// FindByID finds a user by ID
func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return repository.User{}, errors.New("user not found")
	}

	return user, nil
}

// FindOne finds a single user matching the filter
func (r *InMemoryUserRepository) FindOne(ctx context.Context, filter repository.Filter) (repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simple implementation for common filters
	for _, user := range r.users {
		if filter["email"] != nil && user.Email == filter["email"] {
			return user, nil
		}
		if filter["username"] != nil && user.Username == filter["username"] {
			return user, nil
		}
		if filter["id"] != nil && user.ID == filter["id"] {
			return user, nil
		}
	}

	return repository.User{}, errors.New("user not found")
}

// FindMany finds multiple users matching the filter
func (r *InMemoryUserRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []repository.User
	skip := options.Skip
	limit := options.Limit

	var count int64
	for _, user := range r.users {
		// Filter
		match := true
		if filter["email"] != nil && user.Email != filter["email"] {
			match = false
		}
		if filter["username"] != nil && user.Username != filter["username"] {
			match = false
		}
		if filter["organizationIds"] != nil {
			orgID := filter["organizationIds"].(string)
			found := false
			for _, id := range user.OrganizationIDs {
				if id == orgID {
					found = true
					break
				}
			}
			if !found {
				match = false
			}
		}

		if match {
			count++
			if count > skip && (limit == 0 || count <= skip+limit) {
				users = append(users, user)
			}
		}
	}

	return users, nil
}

// Count counts users matching the filter
func (r *InMemoryUserRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, user := range r.users {
		// Filter
		match := true
		if filter["email"] != nil && user.Email != filter["email"] {
			match = false
		}
		if filter["username"] != nil && user.Username != filter["username"] {
			match = false
		}
		if filter["organizationIds"] != nil {
			orgID := filter["organizationIds"].(string)
			found := false
			for _, id := range user.OrganizationIDs {
				if id == orgID {
					found = true
					break
				}
			}
			if !found {
				match = false
			}
		}

		if match {
			count++
		}
	}

	return count, nil
}

// Create creates a new user
func (r *InMemoryUserRepository) Create(ctx context.Context, user repository.User) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	for _, u := range r.users {
		if u.Email == user.Email {
			return "", errors.New("email already exists")
		}
		if u.Username == user.Username {
			return "", errors.New("username already exists")
		}
	}

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now().Format(time.RFC3339)
	if user.User.CreatedAt == "" {
		user.User.CreatedAt = now
	}
	if user.User.UpdatedAt == "" {
		user.User.UpdatedAt = now
	}

	r.users[user.ID] = user
	return user.ID, nil
}

// Update updates an existing user
func (r *InMemoryUserRepository) Update(ctx context.Context, id string, user repository.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	_, ok := r.users[id]
	if !ok {
		return errors.New("user not found")
	}

	// Check if email already exists for another user
	if user.Email != "" {
		for uid, u := range r.users {
			if uid != id && u.Email == user.Email {
				return errors.New("email already exists")
			}
		}
	}

	// Check if username already exists for another user
	if user.Username != "" {
		for uid, u := range r.users {
			if uid != id && u.Username == user.Username {
				return errors.New("username already exists")
			}
		}
	}

	r.users[id] = user
	return nil
}

// Delete deletes a user
func (r *InMemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	_, ok := r.users[id]
	if !ok {
		return errors.New("user not found")
	}

	delete(r.users, id)
	return nil
}

// FindByUsername finds a user by username
func (r *InMemoryUserRepository) FindByUsername(ctx context.Context, username string) (repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return repository.User{}, errors.New("user not found")
}

// FindByEmail finds a user by email
func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return repository.User{}, errors.New("user not found")
}

// FindByOAuthID finds a user by OAuth provider and ID
func (r *InMemoryUserRepository) FindByOAuthID(ctx context.Context, provider, providerUserID string) (repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		for _, account := range user.OAuthAccounts {
			if account.Provider == provider && account.ProviderUserID == providerUserID {
				return user, nil
			}
		}
	}

	return repository.User{}, errors.New("user not found")
}

// UpdateLastLogin updates a user's last login time
func (r *InMemoryUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[id]
	if !ok {
		return errors.New("user not found")
	}

	// Update last login with formatted time string
	now := time.Now().Format(time.RFC3339)
	user.LastLogin = now
	user.UpdatedAt = now
	r.users[id] = user

	return nil
}

// AddRoleToUser adds a role to a user
func (r *InMemoryUserRepository) AddRoleToUser(ctx context.Context, userID, roleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	// Check if user already has the role
	for _, id := range user.RoleIDs {
		if id == roleID {
			return nil
		}
	}

	// Add role
	user.RoleIDs = append(user.RoleIDs, roleID)
	user.UpdatedAt = time.Now().Format(time.RFC3339)
	r.users[userID] = user

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *InMemoryUserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	// Remove role
	var newRoles []string
	for _, id := range user.RoleIDs {
		if id != roleID {
			newRoles = append(newRoles, id)
		}
	}

	user.RoleIDs = newRoles
	user.UpdatedAt = nowAsString()
	r.users[userID] = user

	return nil
}

// AddOrganizationToUser adds an organization to a user
func (r *InMemoryUserRepository) AddOrganizationToUser(ctx context.Context, userID, organizationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	// Check if user already has the organization
	for _, id := range user.OrganizationIDs {
		if id == organizationID {
			return nil
		}
	}

	// Add organization
	user.OrganizationIDs = append(user.OrganizationIDs, organizationID)
	user.UpdatedAt = nowAsString()
	r.users[userID] = user

	return nil
}

// RemoveOrganizationFromUser removes an organization from a user
func (r *InMemoryUserRepository) RemoveOrganizationFromUser(ctx context.Context, userID, organizationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	// Remove organization
	var newOrgs []string
	for _, id := range user.OrganizationIDs {
		if id != organizationID {
			newOrgs = append(newOrgs, id)
		}
	}

	user.OrganizationIDs = newOrgs
	user.UpdatedAt = nowAsString()
	r.users[userID] = user

	return nil
}

// FindByRoleID finds users with a specific role
func (r *InMemoryUserRepository) FindByRoleID(ctx context.Context, roleID string) ([]repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []repository.User
	for _, user := range r.users {
		for _, id := range user.RoleIDs {
			if id == roleID {
				users = append(users, user)
				break
			}
		}
	}

	return users, nil
}

// FindByOrganizationID finds users in a specific organization
func (r *InMemoryUserRepository) FindByOrganizationID(ctx context.Context, organizationID string) ([]repository.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []repository.User
	for _, user := range r.users {
		for _, id := range user.OrganizationIDs {
			if id == organizationID {
				users = append(users, user)
				break
			}
		}
	}

	return users, nil
}

// LinkOAuthAccount links an OAuth account to a user
func (r *InMemoryUserRepository) LinkOAuthAccount(ctx context.Context, userID, provider, providerUserID, refreshToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	// Check if account already linked
	for i, account := range user.OAuthAccounts {
		if account.Provider == provider {
			// Update existing account
			user.OAuthAccounts[i].ProviderUserID = providerUserID
			user.OAuthAccounts[i].RefreshToken = refreshToken
			user.OAuthAccounts[i].LinkedAt = time.Now() // OAuthAccount.LinkedAt is time.Time
			r.users[userID] = user
			return nil
		}
	}

	// Add new account
	user.OAuthAccounts = append(user.OAuthAccounts, repository.OAuthAccount{
		Provider:       provider,
		ProviderUserID: providerUserID,
		RefreshToken:   refreshToken,
		LinkedAt:       time.Now(), // OAuthAccount.LinkedAt is time.Time
	})
	user.UpdatedAt = nowAsString()
	r.users[userID] = user

	return nil
}

// UnlinkOAuthAccount unlinks an OAuth account from a user
func (r *InMemoryUserRepository) UnlinkOAuthAccount(ctx context.Context, userID, provider string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user exists
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	// Remove account
	var newAccounts []repository.OAuthAccount
	for _, account := range user.OAuthAccounts {
		if account.Provider != provider {
			newAccounts = append(newAccounts, account)
		}
	}

	user.OAuthAccounts = newAccounts
	user.UpdatedAt = nowAsString()
	r.users[userID] = user

	return nil
}

// InMemoryOrganizationRepository implements identity.OrganizationRepository
type InMemoryOrganizationRepository struct {
	mu   sync.RWMutex
	orgs map[string]repository.Organization
}

// NewInMemoryOrganizationRepository creates a new in-memory organization repository
func NewInMemoryOrganizationRepository() repository.OrganizationRepository {
	return &InMemoryOrganizationRepository{
		orgs: make(map[string]repository.Organization),
	}
}

// IsConnected checks if the repository is connected
func (r *InMemoryOrganizationRepository) IsConnected(ctx context.Context) bool {
	return true
}

// FindByID finds an organization by ID
func (r *InMemoryOrganizationRepository) FindByID(ctx context.Context, id string) (repository.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	org, ok := r.orgs[id]
	if !ok {
		return repository.Organization{}, errors.New("organization not found")
	}

	return org, nil
}

// FindOne finds a single organization matching the filter
func (r *InMemoryOrganizationRepository) FindOne(ctx context.Context, filter repository.Filter) (repository.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simple implementation for common filters
	for _, org := range r.orgs {
		if filter["name"] != nil && org.Name == filter["name"] {
			return org, nil
		}
		if filter["domain"] != nil && org.Domain == filter["domain"] {
			return org, nil
		}
		if filter["id"] != nil && org.ID == filter["id"] {
			return org, nil
		}
	}

	return repository.Organization{}, errors.New("organization not found")
}

// FindMany finds multiple organizations matching the filter
func (r *InMemoryOrganizationRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]repository.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orgs []repository.Organization
	skip := options.Skip
	limit := options.Limit

	var count int64
	for _, org := range r.orgs {
		// Filter
		match := true
		if filter["name"] != nil && org.Name != filter["name"] {
			match = false
		}
		if filter["domain"] != nil && org.Domain != filter["domain"] {
			match = false
		}

		if match {
			count++
			if count > skip && (limit == 0 || count <= skip+limit) {
				orgs = append(orgs, org)
			}
		}
	}

	return orgs, nil
}

// Count counts organizations matching the filter
func (r *InMemoryOrganizationRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, org := range r.orgs {
		// Filter
		match := true
		if filter["name"] != nil && org.Name != filter["name"] {
			match = false
		}
		if filter["domain"] != nil && org.Domain != filter["domain"] {
			match = false
		}

		if match {
			count++
		}
	}

	return count, nil
}

// Create creates a new organization
func (r *InMemoryOrganizationRepository) Create(ctx context.Context, org repository.Organization) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if name already exists
	for _, o := range r.orgs {
		if o.Name == org.Name {
			return "", errors.New("name already exists")
		}
		if o.Domain != "" && o.Domain == org.Domain {
			return "", errors.New("domain already exists")
		}
	}

	// Generate ID if not provided
	if org.ID == "" {
		org.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now().Format(time.RFC3339)
	if org.CreatedAt == "" {
		org.CreatedAt = now
	}
	if org.UpdatedAt == "" {
		org.UpdatedAt = now
	}

	r.orgs[org.ID] = org
	return org.ID, nil
}

// Update updates an existing organization
func (r *InMemoryOrganizationRepository) Update(ctx context.Context, id string, org repository.Organization) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if organization exists
	_, ok := r.orgs[id]
	if !ok {
		return errors.New("organization not found")
	}

	// Check if name already exists for another organization
	if org.Name != "" {
		for oid, o := range r.orgs {
			if oid != id && o.Name == org.Name {
				return errors.New("name already exists")
			}
		}
	}

	// Check if domain already exists for another organization
	if org.Domain != "" {
		for oid, o := range r.orgs {
			if oid != id && o.Domain != "" && o.Domain == org.Domain {
				return errors.New("domain already exists")
			}
		}
	}

	r.orgs[id] = org
	return nil
}

// Delete deletes an organization
func (r *InMemoryOrganizationRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if organization exists
	_, ok := r.orgs[id]
	if !ok {
		return errors.New("organization not found")
	}

	delete(r.orgs, id)
	return nil
}

// FindByName finds an organization by name
func (r *InMemoryOrganizationRepository) FindByName(ctx context.Context, name string) (repository.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, org := range r.orgs {
		if org.Name == name {
			return org, nil
		}
	}

	return repository.Organization{}, errors.New("organization not found")
}

// FindByDomain finds an organization by domain
func (r *InMemoryOrganizationRepository) FindByDomain(ctx context.Context, domain string) (repository.Organization, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, org := range r.orgs {
		if org.Domain == domain {
			return org, nil
		}
	}

	return repository.Organization{}, errors.New("organization not found")
}

// AddAdminToOrganization adds an admin to an organization
func (r *InMemoryOrganizationRepository) AddAdminToOrganization(ctx context.Context, orgID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if organization exists
	org, ok := r.orgs[orgID]
	if !ok {
		return errors.New("organization not found")
	}

	// Check if user already an admin
	for _, id := range org.AdminIDs {
		if id == userID {
			return nil
		}
	}

	// Add admin
	org.AdminIDs = append(org.AdminIDs, userID)
	org.UpdatedAt = nowAsString()
	r.orgs[orgID] = org

	return nil
}

// RemoveAdminFromOrganization removes an admin from an organization
func (r *InMemoryOrganizationRepository) RemoveAdminFromOrganization(ctx context.Context, orgID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if organization exists
	org, ok := r.orgs[orgID]
	if !ok {
		return errors.New("organization not found")
	}

	// Remove admin
	var newAdmins []string
	for _, id := range org.AdminIDs {
		if id != userID {
			newAdmins = append(newAdmins, id)
		}
	}

	org.AdminIDs = newAdmins
	org.UpdatedAt = nowAsString()
	r.orgs[orgID] = org

	return nil
}

// IsUserAdmin checks if a user is an admin of an organization
func (r *InMemoryOrganizationRepository) IsUserAdmin(ctx context.Context, orgID, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if organization exists
	org, ok := r.orgs[orgID]
	if !ok {
		return false, errors.New("organization not found")
	}

	// Check if user is an admin
	for _, id := range org.AdminIDs {
		if id == userID {
			return true, nil
		}
	}

	return false, nil
}

// InMemoryRoleRepository implements rbac.RoleRepository
type InMemoryRoleRepository struct {
	mu    sync.RWMutex
	roles map[string]rbac.Role
}

// NewInMemoryRoleRepository creates a new in-memory role repository
func NewInMemoryRoleRepository() rbac.RoleRepository {
	return &InMemoryRoleRepository{
		roles: make(map[string]rbac.Role),
	}
}

// IsConnected checks if the repository is connected
func (r *InMemoryRoleRepository) IsConnected(ctx context.Context) bool {
	return true
}

// FindByID finds a role by ID
func (r *InMemoryRoleRepository) FindByID(ctx context.Context, id string) (rbac.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, ok := r.roles[id]
	if !ok {
		return rbac.Role{}, errors.New("role not found")
	}

	return role, nil
}

// FindOne finds a single role matching the filter
func (r *InMemoryRoleRepository) FindOne(ctx context.Context, filter repository.Filter) (rbac.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simple implementation for common filters
	for _, role := range r.roles {
		if filter["name"] != nil && role.Name == filter["name"] {
			return role, nil
		}
		if filter["id"] != nil && role.ID == filter["id"] {
			return role, nil
		}
		if filter["organizationId"] != nil && role.OrganizationID == filter["organizationId"] {
			return role, nil
		}
	}

	return rbac.Role{}, errors.New("role not found")
}

// FindMany finds multiple roles matching the filter
func (r *InMemoryRoleRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]rbac.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var roles []rbac.Role
	skip := options.Skip
	limit := options.Limit

	var count int64
	for _, role := range r.roles {
		// Filter
		match := true
		if filter["name"] != nil && role.Name != filter["name"] {
			match = false
		}
		if filter["organizationId"] != nil && role.OrganizationID != filter["organizationId"] {
			match = false
		}

		if match {
			count++
			if count > skip && (limit == 0 || count <= skip+limit) {
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

// Count counts roles matching the filter
func (r *InMemoryRoleRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, role := range r.roles {
		// Filter
		match := true
		if filter["name"] != nil && role.Name != filter["name"] {
			match = false
		}
		if filter["organizationId"] != nil && role.OrganizationID != filter["organizationId"] {
			match = false
		}

		if match {
			count++
		}
	}

	return count, nil
}

// Create creates a new role
func (r *InMemoryRoleRepository) Create(ctx context.Context, role rbac.Role) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if name already exists within the same organization
	for _, r := range r.roles {
		if r.Name == role.Name && (r.OrganizationID == role.OrganizationID || r.OrganizationID == "" && role.OrganizationID == "") {
			return "", errors.New("role name already exists in this organization")
		}
	}

	// Generate ID if not provided
	if role.ID == "" {
		role.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now().Format(time.RFC3339)
	if role.CreatedAt == "" {
		role.CreatedAt = now
	}
	if role.UpdatedAt == "" {
		role.UpdatedAt = now
	}

	r.roles[role.ID] = role
	return role.ID, nil
}

// Update updates an existing role
func (r *InMemoryRoleRepository) Update(ctx context.Context, id string, role rbac.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if role exists
	_, ok := r.roles[id]
	if !ok {
		return errors.New("role not found")
	}

	// Check if name already exists for another role in the same organization
	if role.Name != "" {
		for rid, r := range r.roles {
			if rid != id && r.Name == role.Name && (r.OrganizationID == role.OrganizationID || r.OrganizationID == "" && role.OrganizationID == "") {
				return errors.New("role name already exists in this organization")
			}
		}
	}

	r.roles[id] = role
	return nil
}

// Delete deletes a role
func (r *InMemoryRoleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if role exists
	role, ok := r.roles[id]
	if !ok {
		return errors.New("role not found")
	}

	// Don't delete system default roles
	if role.IsSystemDefault {
		return errors.New("cannot delete system default role")
	}

	delete(r.roles, id)
	return nil
}

// FindByName finds a role by its name
func (r *InMemoryRoleRepository) FindByName(ctx context.Context, name string) (rbac.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, role := range r.roles {
		if role.Name == name {
			return role, nil
		}
	}

	return rbac.Role{}, errors.New("role not found")
}

// FindByOrganization finds roles for an organization
func (r *InMemoryRoleRepository) FindByOrganization(ctx context.Context, orgID string) ([]rbac.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var roles []rbac.Role
	for _, role := range r.roles {
		if role.OrganizationID == orgID {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

// AddPermission adds a permission to a role
func (r *InMemoryRoleRepository) AddPermission(ctx context.Context, roleID string, permissionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if role exists
	role, ok := r.roles[roleID]
	if !ok {
		return errors.New("role not found")
	}

	// Check if permission already exists
	for _, pid := range role.PermissionIDs {
		if pid == permissionID {
			return nil
		}
	}

	// Add permission
	role.PermissionIDs = append(role.PermissionIDs, permissionID)
	role.UpdatedAt = nowAsString()
	r.roles[roleID] = role

	return nil
}

// RemovePermission removes a permission from a role
func (r *InMemoryRoleRepository) RemovePermission(ctx context.Context, roleID string, permissionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if role exists
	role, ok := r.roles[roleID]
	if !ok {
		return errors.New("role not found")
	}

	// Remove permission
	var newPermissions []string
	for _, pid := range role.PermissionIDs {
		if pid != permissionID {
			newPermissions = append(newPermissions, pid)
		}
	}

	role.PermissionIDs = newPermissions
	role.UpdatedAt = nowAsString()
	r.roles[roleID] = role

	return nil
}

// AddPermissionToRole adds a permission to a role (matching the expected interface name)
func (r *InMemoryRoleRepository) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) error {
	// Delegate to AddPermission which has the same implementation
	return r.AddPermission(ctx, roleID, permissionID)
}

// RemovePermissionFromRole removes a permission from a role (matching the expected interface name)
func (r *InMemoryRoleRepository) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error {
	// Delegate to RemovePermission which has the same implementation
	return r.RemovePermission(ctx, roleID, permissionID)
}

// FindSystemDefaults finds all system default roles
func (r *InMemoryRoleRepository) FindSystemDefaults(ctx context.Context) ([]rbac.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var roles []rbac.Role
	for _, role := range r.roles {
		if role.IsSystemDefault {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

// InMemoryPermissionRepository implements rbac.PermissionRepository
type InMemoryPermissionRepository struct {
	mu          sync.RWMutex
	permissions map[string]models.Permission
}

// NewInMemoryPermissionRepository creates a new in-memory permission repository
func NewInMemoryPermissionRepository() rbac.PermissionRepository {
	return &InMemoryPermissionRepository{
		permissions: make(map[string]models.Permission),
	}
}

// FindByResourceAction finds permissions by resource and action
func (r *InMemoryPermissionRepository) FindByResourceAction(ctx context.Context, resource, action string) ([]models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permissions []models.Permission

	for _, permission := range r.permissions {
		if permission.Resource == resource && permission.Action == action {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// IsConnected checks if the repository is connected
func (r *InMemoryPermissionRepository) IsConnected(ctx context.Context) bool {
	return true
}

// FindByID finds a permission by ID
func (r *InMemoryPermissionRepository) FindByID(ctx context.Context, id string) (models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	permission, ok := r.permissions[id]
	if !ok {
		return models.Permission{}, errors.New("permission not found")
	}

	return permission, nil
}

// FindOne finds a single permission matching the filter
func (r *InMemoryPermissionRepository) FindOne(ctx context.Context, filter repository.Filter) (models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Simple implementation for common filters
	for _, permission := range r.permissions {
		if filter["name"] != nil && permission.Name == filter["name"] {
			return permission, nil
		}
		if filter["id"] != nil && permission.ID == filter["id"] {
			return permission, nil
		}
		if filter["resource"] != nil && permission.Resource == filter["resource"] && filter["action"] != nil && permission.Action == filter["action"] {
			return permission, nil
		}
	}

	return models.Permission{}, errors.New("permission not found")
}

// FindMany finds multiple permissions matching the filter
func (r *InMemoryPermissionRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permissions []models.Permission
	skip := options.Skip
	limit := options.Limit

	var count int64
	for _, permission := range r.permissions {
		// Filter
		match := true
		if filter["name"] != nil && permission.Name != filter["name"] {
			match = false
		}
		if filter["resource"] != nil && permission.Resource != filter["resource"] {
			match = false
		}
		if filter["action"] != nil && permission.Action != filter["action"] {
			match = false
		}
		if filter["organizationId"] != nil && permission.OrganizationID != filter["organizationId"] {
			match = false
		}

		if match {
			count++
			if count > skip && (limit == 0 || count <= skip+limit) {
				permissions = append(permissions, permission)
			}
		}
	}

	return permissions, nil
}

// Count counts permissions matching the filter
func (r *InMemoryPermissionRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, permission := range r.permissions {
		// Filter
		match := true
		if filter["name"] != nil && permission.Name != filter["name"] {
			match = false
		}
		if filter["resource"] != nil && permission.Resource != filter["resource"] {
			match = false
		}
		if filter["action"] != nil && permission.Action != filter["action"] {
			match = false
		}
		if filter["organizationId"] != nil && permission.OrganizationID != filter["organizationId"] {
			match = false
		}

		if match {
			count++
		}
	}

	return count, nil
}

// Create creates a new permission
func (r *InMemoryPermissionRepository) Create(ctx context.Context, permission models.Permission) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if name already exists
	for _, p := range r.permissions {
		if p.Name == permission.Name {
			return "", errors.New("permission name already exists")
		}
		if p.Resource == permission.Resource && p.Action == permission.Action && (p.OrganizationID == permission.OrganizationID || p.OrganizationID == "" && permission.OrganizationID == "") {
			return "", errors.New("permission for this resource and action already exists in this organization")
		}
	}

	// Generate ID if not provided
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now()
	if permission.CreatedAt.IsZero() {
		permission.CreatedAt = now
	}
	if permission.UpdatedAt.IsZero() {
		permission.UpdatedAt = now
	}

	r.permissions[permission.ID] = permission
	return permission.ID, nil
}

// Update updates an existing permission
func (r *InMemoryPermissionRepository) Update(ctx context.Context, id string, permission models.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if permission exists
	_, ok := r.permissions[id]
	if !ok {
		return errors.New("permission not found")
	}

	// Check if name already exists for another permission
	if permission.Name != "" {
		for pid, p := range r.permissions {
			if pid != id && p.Name == permission.Name {
				return errors.New("permission name already exists")
			}
		}
	}

	// Check if resource+action already exists for another permission in the same organization
	for pid, p := range r.permissions {
		if pid != id && p.Resource == permission.Resource && p.Action == permission.Action && (p.OrganizationID == permission.OrganizationID || p.OrganizationID == "" && permission.OrganizationID == "") {
			return errors.New("permission for this resource and action already exists in this organization")
		}
	}

	r.permissions[id] = permission
	return nil
}

// Delete deletes a permission
func (r *InMemoryPermissionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if permission exists
	permission, ok := r.permissions[id]
	if !ok {
		return errors.New("permission not found")
	}

	// Don't delete system default permissions
	if permission.IsSystemDefault {
		return errors.New("cannot delete system default permission")
	}

	delete(r.permissions, id)
	return nil
}

// FindByName finds a permission by its name
func (r *InMemoryPermissionRepository) FindByName(ctx context.Context, name string) (models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, permission := range r.permissions {
		if permission.Name == name {
			return permission, nil
		}
	}

	return models.Permission{}, errors.New("permission not found")
}

// FindByResource finds permissions for a resource
func (r *InMemoryPermissionRepository) FindByResource(ctx context.Context, resource string) ([]models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permissions []models.Permission
	for _, permission := range r.permissions {
		if permission.Resource == resource {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// FindByResourceAndAction finds a permission by resource and action
func (r *InMemoryPermissionRepository) FindByResourceAndAction(ctx context.Context, resource, action string) (models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, permission := range r.permissions {
		if permission.Resource == resource && permission.Action == action {
			return permission, nil
		}
	}

	return models.Permission{}, errors.New("permission not found")
}

// FindByOrganization finds permissions for an organization
func (r *InMemoryPermissionRepository) FindByOrganization(ctx context.Context, orgID string) ([]models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permissions []models.Permission
	for _, permission := range r.permissions {
		if permission.OrganizationID == orgID {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// FindByIDs finds permissions by their IDs
func (r *InMemoryPermissionRepository) FindByIDs(ctx context.Context, ids []string) ([]models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permissions []models.Permission
	for _, id := range ids {
		permission, ok := r.permissions[id]
		if ok {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// FindSystemDefaults finds all system default permissions
func (r *InMemoryPermissionRepository) FindSystemDefaults(ctx context.Context) ([]models.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var permissions []models.Permission
	for _, permission := range r.permissions {
		if permission.IsSystemDefault {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// Connect attempts to connect to MongoDB
func Connect(uri string) (interface{}, error) {
	if uri == "" {
		return nil, errors.New("mongodb uri is empty")
	}

	// In a real implementation, this would connect to MongoDB
	// For now, just return nil
	return nil, errors.New("mongodb connection not implemented")
}

// InMemoryRepositoryFactory creates in-memory repositories for testing and fallback
type InMemoryRepositoryFactory struct{}

// NewInMemoryRepositoryFactory creates a new in-memory repository factory
func NewInMemoryRepositoryFactory() *InMemoryRepositoryFactory {
	return &InMemoryRepositoryFactory{}
}

// CreateUserRepository creates a new in-memory user repository
func (f *InMemoryRepositoryFactory) CreateUserRepository() repository.UserRepository {
	return NewInMemoryUserRepository()
}

// CreateOrganizationRepository creates a new in-memory organization repository
func (f *InMemoryRepositoryFactory) CreateOrganizationRepository() repository.OrganizationRepository {
	return NewInMemoryOrganizationRepository()
}

// CreateRoleRepository creates a new in-memory role repository
func (f *InMemoryRepositoryFactory) CreateRoleRepository() rbac.RoleRepository {
	return NewInMemoryRoleRepository()
}

// CreatePermissionRepository creates a new in-memory permission repository
func (f *InMemoryRepositoryFactory) CreatePermissionRepository() rbac.PermissionRepository {
	return NewInMemoryPermissionRepository()
}
