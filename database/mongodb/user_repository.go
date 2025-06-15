package mongodb

import (
	"context"
	"errors"
	"rbac-system/pkg/common/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements identity.UserRepository using MongoDB
type MongoUserRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoUserRepository creates a new MongoDB user repository
func NewMongoUserRepository(client *mongo.Client, database, collection string) repository.UserRepository {
	return &MongoUserRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// getCollection returns the MongoDB collection for users
func (r *MongoUserRepository) getCollection() *mongo.Collection {
	return r.client.Database(r.database).Collection(r.collection)
}

// IsConnected checks if the repository is connected to the database
func (r *MongoUserRepository) IsConnected(ctx context.Context) bool {
	return r.client != nil && r.client.Ping(ctx, nil) == nil
}

// FindByID finds a user by ID
func (r *MongoUserRepository) FindByID(ctx context.Context, id string) (repository.User, error) {
	var user repository.User

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return user, err
	}

	filter := bson.M{"_id": objectID}
	err = r.getCollection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, repository.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

// FindOne finds a single user matching the filter
func (r *MongoUserRepository) FindOne(ctx context.Context, filter repository.Filter) (repository.User, error) {
	var user repository.User

	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	err := r.getCollection().FindOne(ctx, mongoFilter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, repository.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

// FindMany finds multiple users matching the filter
func (r *MongoUserRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]repository.User, error) {
	var users []repository.User

	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	findOptions := repository.ToMongoOptions(options)
	cursor, err := r.getCollection().Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Count counts users matching the filter
func (r *MongoUserRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	return r.getCollection().CountDocuments(ctx, mongoFilter)
}

// Create creates a new user
func (r *MongoUserRepository) Create(ctx context.Context, user repository.User) (string, error) {
	// Generate new ID if not provided
	if user.ID == "" {
		user.ID = primitive.NewObjectID().Hex()
	}

	// Set timestamps
	user.CreatedAt = nowAsString()
	user.UpdatedAt = nowAsString()

	// Insert document
	result, err := r.getCollection().InsertOne(ctx, user)
	if err != nil {
		return "", err
	}

	// Return ID
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Update updates an existing user
func (r *MongoUserRepository) Update(ctx context.Context, id string, user repository.User) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Set update timestamp
	user.UpdatedAt = nowAsString()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": user}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete deletes a user
func (r *MongoUserRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	result, err := r.getCollection().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// FindByUsername finds a user by username
func (r *MongoUserRepository) FindByUsername(ctx context.Context, username string) (repository.User, error) {
	filter := bson.M{"username": username}
	var user repository.User

	err := r.getCollection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, repository.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

// FindByEmail finds a user by email
func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (repository.User, error) {
	filter := bson.M{"email": email}
	var user repository.User

	err := r.getCollection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, repository.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

// FindByOAuthID finds a user by OAuth provider and ID
func (r *MongoUserRepository) FindByOAuthID(ctx context.Context, provider, providerUserID string) (repository.User, error) {
	filter := bson.M{
		"oauthAccounts.provider":       provider,
		"oauthAccounts.providerUserId": providerUserID,
	}
	var user repository.User

	err := r.getCollection().FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, repository.ErrNotFound
		}
		return user, err
	}

	return user, nil
}

// UpdateLastLogin updates a user's last login time
func (r *MongoUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	now := time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"lastLogin": now,
			"updatedAt": now,
		},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// AddRoleToUser adds a role to a user
func (r *MongoUserRepository) AddRoleToUser(ctx context.Context, userID, roleID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$addToSet": bson.M{"roleIds": roleID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *MongoUserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$pull": bson.M{"roleIds": roleID},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// AddOrganizationToUser adds an organization to a user
func (r *MongoUserRepository) AddOrganizationToUser(ctx context.Context, userID, organizationID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$addToSet": bson.M{"organizationIds": organizationID},
		"$set":      bson.M{"updatedAt": time.Now()},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// RemoveOrganizationFromUser removes an organization from a user
func (r *MongoUserRepository) RemoveOrganizationFromUser(ctx context.Context, userID, organizationID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$pull": bson.M{"organizationIds": organizationID},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// FindByRoleID finds users with a specific role
func (r *MongoUserRepository) FindByRoleID(ctx context.Context, roleID string) ([]repository.User, error) {
	filter := bson.M{"roleIds": roleID}
	opts := options.Find().SetSort(bson.D{{Key: "username", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []repository.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// FindByOrganizationID finds users in a specific organization
func (r *MongoUserRepository) FindByOrganizationID(ctx context.Context, organizationID string) ([]repository.User, error) {
	filter := bson.M{"organizationIds": organizationID}
	opts := options.Find().SetSort(bson.D{{Key: "username", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []repository.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// LinkOAuthAccount links an OAuth account to a user
func (r *MongoUserRepository) LinkOAuthAccount(ctx context.Context, userID, provider, providerUserID, refreshToken string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	oauthAccount := repository.OAuthAccount{
		Provider:       provider,
		ProviderUserID: providerUserID,
		RefreshToken:   refreshToken,
		LinkedAt:       time.Now(),
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$push": bson.M{"oauthAccounts": oauthAccount},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// UnlinkOAuthAccount unlinks an OAuth account from a user
func (r *MongoUserRepository) UnlinkOAuthAccount(ctx context.Context, userID, provider string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$pull": bson.M{"oauthAccounts": bson.M{"provider": provider}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}
