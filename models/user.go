package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthProvider represents the authentication provider for a user
type AuthProvider string

const (
	// Local authentication (username/password)
	LocalAuth AuthProvider = "local"
	// Google OAuth provider
	GoogleAuth AuthProvider = "google"
	// GitHub OAuth provider
	GitHubAuth AuthProvider = "github"
)

// User represents a user in the system
type User struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Username        string               `bson:"username" json:"username"`
	Email           string               `bson:"email" json:"email"`
	HashedPassword  string               `bson:"hashedPassword,omitempty" json:"-"`
	FirstName       string               `bson:"firstName" json:"firstName"`
	LastName        string               `bson:"lastName" json:"lastName"`
	Active          bool                 `bson:"active" json:"active"`
	EmailVerified   bool                 `bson:"emailVerified" json:"emailVerified"`
	RoleIDs         []primitive.ObjectID `bson:"roleIds" json:"roleIds"`
	OrganizationIDs []primitive.ObjectID `bson:"organizationIds" json:"organizationIds"`
	AuthProvider    AuthProvider         `bson:"authProvider" json:"authProvider"`
	ProviderUserID  string               `bson:"providerUserId,omitempty" json:"providerUserId,omitempty"`
	RefreshToken    string               `bson:"refreshToken" json:"refreshToken"`
	CreatedAt       time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time            `bson:"updatedAt" json:"updatedAt"`
	LastLogin       time.Time            `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
}

// UserCreateInput represents the input to create a new user
type UserCreateInput struct {
	Username        string               `json:"username" validate:"required"`
	Email           string               `json:"email" validate:"required,email"`
	Password        string               `json:"password,omitempty" validate:"omitempty,min=8"`
	FirstName       string               `json:"firstName" validate:"required"`
	LastName        string               `json:"lastName" validate:"required"`
	RoleIDs         []primitive.ObjectID `json:"roleIds,omitempty"`
	OrganizationIDs []primitive.ObjectID `json:"organizationIds,omitempty"`
	AuthProvider    AuthProvider         `json:"authProvider" validate:"required"`
	ProviderUserID  string               `json:"providerUserId,omitempty"`
}

// UserUpdateInput represents the input to update a user
type UserUpdateInput struct {
	Username        string               `json:"username,omitempty"`
	Email           string               `json:"email,omitempty" validate:"omitempty,email"`
	Password        string               `json:"password,omitempty" validate:"omitempty,min=8"`
	FirstName       string               `json:"firstName,omitempty"`
	LastName        string               `json:"lastName,omitempty"`
	Active          *bool                `json:"active,omitempty"`
	RoleIDs         []primitive.ObjectID `json:"roleIds,omitempty"`
	OrganizationIDs []primitive.ObjectID `json:"organizationIds,omitempty"`
}

// UserResponse represents the response for user endpoints
type UserResponse struct {
	ID              primitive.ObjectID   `json:"id,omitempty"`
	Username        string               `json:"username"`
	Email           string               `json:"email"`
	FirstName       string               `json:"firstName"`
	LastName        string               `json:"lastName"`
	Active          bool                 `json:"active"`
	EmailVerified   bool                 `json:"emailVerified"`
	RoleIDs         []primitive.ObjectID `json:"roleIds"`
	OrganizationIDs []primitive.ObjectID `json:"organizationIds"`
	AuthProvider    AuthProvider         `json:"authProvider"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	LastLogin       time.Time            `json:"lastLogin,omitempty"`
}

// ToUserResponse converts a User to a UserResponse
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		Active:          u.Active,
		EmailVerified:   u.EmailVerified,
		RoleIDs:         u.RoleIDs,
		OrganizationIDs: u.OrganizationIDs,
		AuthProvider:    u.AuthProvider,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		LastLogin:       u.LastLogin,
	}
}

// UserLoginInput represents the input for user login
type UserLoginInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// OAuthUserInfo represents user info from OAuth providers
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
}
