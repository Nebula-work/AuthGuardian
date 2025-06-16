package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Nebula-work/AuthGuardian/config"
	"github.com/Nebula-work/AuthGuardian/database"
	"github.com/Nebula-work/AuthGuardian/models"
	"github.com/Nebula-work/AuthGuardian/utils"
)

// AuthService provides authentication-related operations
type AuthService struct {
	config *config.Config
	db     *database.MongoClient
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.Config, db *database.MongoClient) *AuthService {
	return &AuthService{
		config: cfg,
		db:     db,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(input models.UserCreateInput) (*models.User, error) {
	// Check if username or email already exists
	collection := s.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{
			"$or": []bson.M{
				{"username": input.Username},
				{"email": input.Email},
			},
		},
	).Decode(&existingUser)

	if err == nil {
		return nil, errors.New("username or email already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, errors.New("failed to check for existing user")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Get the default user role
	rolesCollection := s.db.GetCollection(database.RolesCollection)
	defaultRole := models.Role{}

	err = rolesCollection.FindOne(
		context.Background(),
		bson.M{"name": models.UserRole, "isSystemDefault": true},
	).Decode(&defaultRole)

	var roleIDs []primitive.ObjectID
	if err == nil {
		roleIDs = append(roleIDs, defaultRole.ID)
	} else if err != mongo.ErrNoDocuments {
		return nil, errors.New("failed to retrieve default role")
	}

	// Create user
	user := models.User{
		Username:        input.Username,
		Email:           input.Email,
		HashedPassword:  hashedPassword,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		Active:          true,
		EmailVerified:   false,
		RoleIDs:         roleIDs,
		OrganizationIDs: input.OrganizationIDs,
		AuthProvider:    models.LocalAuth,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	return &user, nil
}

// LoginUser authenticates a user
func (s *AuthService) LoginUser(input models.UserLoginInput) (*models.User, string, error) {
	// Find user by username
	collection := s.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{"username": input.Username, "authProvider": models.LocalAuth},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, "", errors.New("invalid credentials")
	} else if err != nil {
		return nil, "", errors.New("failed to retrieve user")
	}

	// Verify password
	valid, err := utils.VerifyPassword(input.Password, user.HashedPassword)
	if err != nil || !valid {
		return nil, "", errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		return nil, "", errors.New("user account is disabled")
	}

	// Update last login
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{"lastLogin": time.Now()}},
	)
	if err != nil {
		return nil, "", errors.New("failed to update last login")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		s.config.JWTSecret,
		s.config.JWTExpirationTime,
	)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &user, token, nil
}

// ProcessOAuthUser processes the authenticated OAuth user
func (s *AuthService) ProcessOAuthUser(userInfo models.OAuthUserInfo, provider models.AuthProvider) (*models.User, string, error) {
	// Check if user already exists
	collection := s.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{
			"$or": []bson.M{
				{"email": userInfo.Email, "authProvider": provider},
				{"providerUserId": userInfo.ID, "authProvider": provider},
			},
		},
	).Decode(&existingUser)

	if err == nil {
		// User exists, update last login
		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"_id": existingUser.ID},
			bson.M{"$set": bson.M{"lastLogin": time.Now()}},
		)
		if err != nil {
			return nil, "", errors.New("failed to update last login")
		}

		// Generate JWT token
		token, err := utils.GenerateToken(
			existingUser.ID,
			existingUser.Username,
			existingUser.Email,
			existingUser.RoleIDs,
			existingUser.OrganizationIDs,
			s.config.JWTSecret,
			s.config.JWTExpirationTime,
		)
		if err != nil {
			return nil, "", errors.New("failed to generate token")
		}

		return &existingUser, token, nil
	} else if err != mongo.ErrNoDocuments {
		return nil, "", errors.New("failed to check for existing user")
	}

	// User doesn't exist, create new user
	// Get the default user role
	rolesCollection := s.db.GetCollection(database.RolesCollection)
	defaultRole := models.Role{}

	err = rolesCollection.FindOne(
		context.Background(),
		bson.M{"name": models.UserRole, "isSystemDefault": true},
	).Decode(&defaultRole)

	var roleIDs []primitive.ObjectID
	if err == nil {
		roleIDs = append(roleIDs, defaultRole.ID)
	} else if err != mongo.ErrNoDocuments {
		return nil, "", errors.New("failed to retrieve default role")
	}

	// Create user
	user := models.User{
		Username:       userInfo.Username,
		Email:          userInfo.Email,
		FirstName:      userInfo.FirstName,
		LastName:       userInfo.LastName,
		Active:         true,
		EmailVerified:  true, // Email is verified by OAuth provider
		RoleIDs:        roleIDs,
		AuthProvider:   provider,
		ProviderUserID: userInfo.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastLogin:      time.Now(),
	}

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, "", errors.New("failed to create user")
	}

	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		s.config.JWTSecret,
		s.config.JWTExpirationTime,
	)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &user, token, nil
}

// RefreshUserToken refreshes a user's token
func (s *AuthService) RefreshUserToken(userID primitive.ObjectID) (*models.User, string, error) {
	// Find user by ID
	collection := s.db.GetCollection(database.UsersCollection)
	user := models.User{}

	err := collection.FindOne(
		context.Background(),
		bson.M{"_id": userID},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, "", errors.New("user not found")
	} else if err != nil {
		return nil, "", errors.New("failed to retrieve user")
	}

	// Check if user is active
	if !user.Active {
		return nil, "", errors.New("user account is disabled")
	}

	// Generate new JWT token
	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		user.Email,
		user.RoleIDs,
		user.OrganizationIDs,
		s.config.JWTSecret,
		s.config.JWTExpirationTime,
	)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &user, token, nil
}

// VerifyToken verifies a token and returns the user ID
func (s *AuthService) VerifyToken(token string) (*utils.CustomClaims, error) {
	// Validate the token
	claims, err := utils.ValidateToken(token, s.config.JWTSecret)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
