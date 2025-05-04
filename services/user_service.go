package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rbac-system/database"
	"rbac-system/models"
	"rbac-system/utils"
)

// UserService provides user-related operations
type UserService struct {
	db *database.MongoClient
}

// NewUserService creates a new user service
func NewUserService(db *database.MongoClient) *UserService {
	return &UserService{
		db: db,
	}
}

// GetUsers retrieves users with pagination
func (s *UserService) GetUsers(limit, skip int, orgID string) ([]models.User, int64, error) {
	// Prepare query
	query := bson.M{}
	
	// Add organization filter if provided
	if orgID != "" {
		objectID, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			return nil, 0, errors.New("invalid organization ID")
		}
		query["organizationIds"] = objectID
	}
	
	// Get users collection
	collection := s.db.GetCollection(database.UsersCollection)
	
	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"username": 1})
	
	// Execute query
	cursor, err := collection.Find(context.Background(), query, findOptions)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve users")
	}
	defer cursor.Close(context.Background())
	
	// Decode users
	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, 0, errors.New("failed to decode users")
	}
	
	// Get total count
	count, err := collection.CountDocuments(context.Background(), query)
	if err != nil {
		return nil, 0, errors.New("failed to count users")
	}
	
	return users, count, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	
	// Get user from database
	collection := s.db.GetCollection(database.UsersCollection)
	user := models.User{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&user)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve user")
	}
	
	return &user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(input models.UserCreateInput) (*models.User, error) {
	// Validate required fields
	if input.Username == "" || input.Email == "" {
		return nil, errors.New("username and email are required")
	}
	
	// For local auth, password is required
	if input.AuthProvider == models.LocalAuth && input.Password == "" {
		return nil, errors.New("password is required for local authentication")
	}
	
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
	
	// Create user
	user := models.User{
		Username:      input.Username,
		Email:         input.Email,
		FirstName:     input.FirstName,
		LastName:      input.LastName,
		Active:        true,
		EmailVerified: false,
		RoleIDs:       input.RoleIDs,
		OrganizationIDs: input.OrganizationIDs,
		AuthProvider:  input.AuthProvider,
		ProviderUserID: input.ProviderUserID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Hash password for local auth
	if input.AuthProvider == models.LocalAuth {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.HashedPassword = hashedPassword
	}
	
	// Insert user
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}
	
	// Get the inserted ID
	user.ID = result.InsertedID.(primitive.ObjectID)
	
	return &user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(id string, input models.UserUpdateInput) (*models.User, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	
	// Check if user exists
	collection := s.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingUser)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve user")
	}
	
	// Prepare update fields
	update := bson.M{"updatedAt": time.Now()}
	
	if input.Username != "" {
		// Check if username is already taken by another user
		if input.Username != existingUser.Username {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"username": input.Username, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return nil, errors.New("failed to check username availability")
			}
			if count > 0 {
				return nil, errors.New("username already taken")
			}
		}
		update["username"] = input.Username
	}
	
	if input.Email != "" {
		// Check if email is already taken by another user
		if input.Email != existingUser.Email {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"email": input.Email, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return nil, errors.New("failed to check email availability")
			}
			if count > 0 {
				return nil, errors.New("email already taken")
			}
		}
		update["email"] = input.Email
	}
	
	if input.FirstName != "" {
		update["firstName"] = input.FirstName
	}
	
	if input.LastName != "" {
		update["lastName"] = input.LastName
	}
	
	if input.Active != nil {
		update["active"] = *input.Active
	}
	
	if len(input.RoleIDs) > 0 {
		update["roleIds"] = input.RoleIDs
	}
	
	if len(input.OrganizationIDs) > 0 {
		update["organizationIds"] = input.OrganizationIDs
	}
	
	// Update password if provided
	if input.Password != "" && existingUser.AuthProvider == models.LocalAuth {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		update["hashedPassword"] = hashedPassword
	}
	
	// Update user
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	
	if err != nil {
		return nil, errors.New("failed to update user")
	}
	
	// Get updated user
	updatedUser := models.User{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedUser)
	
	if err != nil {
		return nil, errors.New("failed to retrieve updated user")
	}
	
	return &updatedUser, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id string) error {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}
	
	// Delete user
	collection := s.db.GetCollection(database.UsersCollection)
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)
	
	if err != nil {
		return errors.New("failed to delete user")
	}
	
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

// GetUsersByOrgID retrieves users by organization ID
func (s *UserService) GetUsersByOrgID(orgID string, limit, skip int) ([]models.User, int64, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, 0, errors.New("invalid organization ID")
	}
	
	// Get users in organization
	collection := s.db.GetCollection(database.UsersCollection)
	
	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"username": 1})
	
	// Execute query
	cursor, err := collection.Find(
		context.Background(),
		bson.M{"organizationIds": objectID},
		findOptions,
	)
	
	if err != nil {
		return nil, 0, errors.New("failed to retrieve users")
	}
	defer cursor.Close(context.Background())
	
	// Decode users
	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, 0, errors.New("failed to decode users")
	}
	
	// Get total count
	count, err := collection.CountDocuments(
		context.Background(),
		bson.M{"organizationIds": objectID},
	)
	if err != nil {
		return nil, 0, errors.New("failed to count users")
	}
	
	return users, count, nil
}

// UpdateUserSelf updates the current user's own profile
func (s *UserService) UpdateUserSelf(userID primitive.ObjectID, input models.UserUpdateInput) (*models.User, error) {
	// Check if user exists
	collection := s.db.GetCollection(database.UsersCollection)
	existingUser := models.User{}
	
	err := collection.FindOne(
		context.Background(),
		bson.M{"_id": userID},
	).Decode(&existingUser)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve user")
	}
	
	// Prepare update fields (limited fields for self-update)
	update := bson.M{"updatedAt": time.Now()}
	
	if input.FirstName != "" {
		update["firstName"] = input.FirstName
	}
	
	if input.LastName != "" {
		update["lastName"] = input.LastName
	}
	
	// Update password if provided
	if input.Password != "" && existingUser.AuthProvider == models.LocalAuth {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		update["hashedPassword"] = hashedPassword
	}
	
	// Update user
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": update},
	)
	
	if err != nil {
		return nil, errors.New("failed to update user")
	}
	
	// Get updated user
	updatedUser := models.User{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": userID},
	).Decode(&updatedUser)
	
	if err != nil {
		return nil, errors.New("failed to retrieve updated user")
	}
	
	return &updatedUser, nil
}
