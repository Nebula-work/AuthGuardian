package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rbac-system/core/rbac"
	"rbac-system/models"
	"rbac-system/pkg/common/repository"
	"time"
)

type MongoRoleRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoRoleRepository creates a new MongoDB role repository
func NewMongoRoleRepository(client *mongo.Client, database, collection string) repository.RoleRepository {
	return &MongoRoleRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// getCollection returns the MongoDB collection for roles
func (r *MongoRoleRepository) getCollection() *mongo.Collection {
	return r.client.Database(r.database).Collection(r.collection)
}

// IsConnected checks if the repository is connected to the database
func (r *MongoRoleRepository) IsConnected(ctx context.Context) bool {
	return r.client != nil && r.client.Ping(ctx, nil) == nil
}

// FindByID finds a role by ID
func (r *MongoRoleRepository) FindByID(ctx context.Context, id string) (models.Role, error) {
	var role models.Role

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return role, err
	}

	filter := bson.M{"_id": objectID}
	err = r.getCollection().FindOne(ctx, filter).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return role, repository.ErrNotFound
		}
		return role, err
	}

	return role, nil
}

// FindOne finds a single role matching the filter
func (r *MongoRoleRepository) FindOne(ctx context.Context, filter repository.Filter) (models.Role, error) {
	var role models.Role

	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	err := r.getCollection().FindOne(ctx, mongoFilter).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return role, repository.ErrNotFound
		}
		return role, err
	}

	return role, nil
}

// FindMany finds multiple roles matching the filter
func (r *MongoRoleRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]models.Role, error) {
	var roles []models.Role

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

	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// Count counts roles matching the filter
func (r *MongoRoleRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	return r.getCollection().CountDocuments(ctx, mongoFilter)
}

// Create creates a new role
func (r *MongoRoleRepository) Create(ctx context.Context, role models.Role) (string, error) {
	// Generate new ID if not provided
	if role.ID.IsZero() {
		role.ID = primitive.NewObjectID()
	}

	// Set timestamps
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	// Insert document
	result, err := r.getCollection().InsertOne(ctx, role)
	if err != nil {
		return "", err
	}

	// Return ID
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Update updates an existing role
func (r *MongoRoleRepository) Update(ctx context.Context, id string, role models.Role) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Set update timestamp
	role.UpdatedAt = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": role}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete deletes a role
func (r *MongoRoleRepository) Delete(ctx context.Context, id string) error {
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

// FindByName finds a role by its name
func (r *MongoRoleRepository) FindByName(ctx context.Context, name string) (models.Role, error) {
	filter := bson.M{"name": name}
	var role models.Role

	err := r.getCollection().FindOne(ctx, filter).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return role, repository.ErrNotFound
		}
		return role, err
	}

	return role, nil
}

// FindByOrganization finds roles for an organization
func (r *MongoRoleRepository) FindByOrganization(ctx context.Context, orgID string) ([]rbac.Role, error) {
	filter := bson.M{"organizationId": orgID}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []rbac.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

// AddPermission adds a permission to a role
func (r *MongoRoleRepository) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) error {
	objectID, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$addToSet": bson.M{"permissionIds": permissionID},
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

// RemovePermission removes a permission from a role
func (r *MongoRoleRepository) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error {
	objectID, err := primitive.ObjectIDFromHex(roleID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$pull": bson.M{"permissionIds": permissionID},
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

// FindSystemDefaults finds all system default roles
func (r *MongoRoleRepository) FindSystemDefaults(ctx context.Context) ([]rbac.Role, error) {
	filter := bson.M{"isSystemDefault": true}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []rbac.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}
