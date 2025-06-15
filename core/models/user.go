package models

// User represents a user in the system
type User struct {
	ID              string   `json:"id"`
	Username        string   `json:"username"`
	Email           string   `json:"email"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	PasswordHash    string   `json:"-"`
	Active          bool     `json:"active"`
	EmailVerified   bool     `json:"emailVerified"`
	RoleIDs         []string `json:"roleIds"`
	OrganizationIDs []string `json:"organizationIds"`
	AuthProvider    string   `json:"authProvider"`
	LastLogin       string   `json:"lastLogin,omitempty"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

// OAuthAccount represents an OAuth provider account linked to a user
type OAuthAccount struct {
	Provider       string `json:"provider"`
	ProviderUserID string `json:"providerUserId"`
	Email          string `json:"email,omitempty"`
	RefreshToken   string `json:"-"`
	AccessToken    string `json:"-"`
}
