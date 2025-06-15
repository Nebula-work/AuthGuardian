package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"time"
)

// MongoPermissionRepository implements rbac.PermissionRepository using MongoDB
type MongoPermissionRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoPermissionRepository creates a new MongoDB permission repository
func NewMongoPermissionRepository(client *mongo.Client, database, collection string) repository.PermissionRepository {
	return &MongoPermissionRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// getCollection returns the MongoDB collection for permissions
func (r *MongoPermissionRepository) getCollection() *mongo.Collection {
	return r.client.Database(r.database).Collection(r.collection)
}

// IsConnected checks if the repository is connected to the database
func (r *MongoPermissionRepository) IsConnected(ctx context.Context) bool {
	return r.client != nil && r.client.Ping(ctx, nil) == nil
}

// FindByID finds a permission by ID
func (r *MongoPermissionRepository) FindByID(ctx context.Context, id string) (models.Permission, error) {
	var permission models.Permission

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return permission, err
	}

	filter := bson.M{"_id": objectID}
	err = r.getCollection().FindOne(ctx, filter).Decode(&permission)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return permission, repository.ErrNotFound
		}
		return permission, err
	}

	return permission, nil
}

// FindOne finds a single permission matching the filter
func (r *MongoPermissionRepository) FindOne(ctx context.Context, filter repository.Filter) (models.Permission, error) {
	var permission models.Permission

	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	err := r.getCollection().FindOne(ctx, mongoFilter).Decode(&permission)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return permission, repository.ErrNotFound
		}
		return permission, err
	}

	return permission, nil
}

// FindMany finds multiple permissions matching the filter
func (r *MongoPermissionRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]models.Permission, error) {
	var permissions []models.Permission

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

	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// Count counts permissions matching the filter
func (r *MongoPermissionRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	return r.getCollection().CountDocuments(ctx, mongoFilter)
}

// Create creates a new permission
func (r *MongoPermissionRepository) Create(ctx context.Context, permission models.Permission) (string, error) {
	// Generate new ID if not provided
	if permission.ID == "" {
		permission.ID = primitive.NewObjectID().Hex()
	}

	// Set timestamps
	now := time.Now()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	// Insert document
	result, err := r.getCollection().InsertOne(ctx, permission)
	if err != nil {
		return "", err
	}

	// Return ID
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Update updates an existing permission
func (r *MongoPermissionRepository) Update(ctx context.Context, id string, permission models.Permission) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Set update timestamp
	permission.UpdatedAt = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": permission}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete deletes a permission
func (r *MongoPermissionRepository) Delete(ctx context.Context, id string) error {
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

// FindByName finds a permission by its name
func (r *MongoPermissionRepository) FindByName(ctx context.Context, name string) (models.Permission, error) {
	filter := bson.M{"name": name}
	var permission models.Permission

	err := r.getCollection().FindOne(ctx, filter).Decode(&permission)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return permission, repository.ErrNotFound
		}
		return permission, err
	}

	return permission, nil
}

// FindByResource finds permissions for a resource
func (r *MongoPermissionRepository) FindByResource(ctx context.Context, resource string) ([]models.Permission, error) {
	filter := bson.M{"resource": resource}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []models.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// FindByResourceAndAction finds a permission by resource and action
func (r *MongoPermissionRepository) FindByResourceAction(ctx context.Context, resource, action string) ([]models.Permission, error) {
	filter := bson.M{"resource": resource, "action": action}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []models.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// FindByOrganization finds permissions for an organization
func (r *MongoPermissionRepository) FindByOrganization(ctx context.Context, orgID string) ([]models.Permission, error) {
	filter := bson.M{"organizationId": orgID}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []models.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// FindByIDs finds permissions by their IDs
func (r *MongoPermissionRepository) FindByIDs(ctx context.Context, ids []string) ([]models.Permission, error) {
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objectID)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []models.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

// FindSystemDefaults finds all system default permissions
func (r *MongoPermissionRepository) FindSystemDefaults(ctx context.Context) ([]models.Permission, error) {
	filter := bson.M{"isSystemDefault": true}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var permissions []models.Permission
	if err := cursor.All(ctx, &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}
