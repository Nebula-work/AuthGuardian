package mongodb

import (
	"context"
	"errors"
	"rbac-system/pkg/common/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoOrganizationRepository implements identity.OrganizationRepository using MongoDB
type MongoOrganizationRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoOrganizationRepository creates a new MongoDB organization repository
func NewMongoOrganizationRepository(client *mongo.Client, database, collection string) repository.OrganizationRepository {
	return &MongoOrganizationRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// getCollection returns the MongoDB collection for organizations
func (r *MongoOrganizationRepository) getCollection() *mongo.Collection {
	return r.client.Database(r.database).Collection(r.collection)
}

// IsConnected checks if the repository is connected to the database
func (r *MongoOrganizationRepository) IsConnected(ctx context.Context) bool {
	return r.client != nil && r.client.Ping(ctx, nil) == nil
}

// FindByID finds an organization by ID
func (r *MongoOrganizationRepository) FindByID(ctx context.Context, id string) (repository.Organization, error) {
	var org repository.Organization

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return org, err
	}

	filter := bson.M{"_id": objectID}
	err = r.getCollection().FindOne(ctx, filter).Decode(&org)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return org, repository.ErrNotFound
		}
		return org, err
	}

	return org, nil
}

// FindOne finds a single organization matching the filter
func (r *MongoOrganizationRepository) FindOne(ctx context.Context, filter repository.Filter) (repository.Organization, error) {
	var org repository.Organization

	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	err := r.getCollection().FindOne(ctx, mongoFilter).Decode(&org)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return org, repository.ErrNotFound
		}
		return org, err
	}

	return org, nil
}

// FindMany finds multiple organizations matching the filter
func (r *MongoOrganizationRepository) FindMany(ctx context.Context, filter repository.Filter, options repository.QueryOptions) ([]repository.Organization, error) {
	var orgs []repository.Organization

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

	if err := cursor.All(ctx, &orgs); err != nil {
		return nil, err
	}

	return orgs, nil
}

// Count counts organizations matching the filter
func (r *MongoOrganizationRepository) Count(ctx context.Context, filter repository.Filter) (int64, error) {
	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	return r.getCollection().CountDocuments(ctx, mongoFilter)
}

// Create creates a new organization
func (r *MongoOrganizationRepository) Create(ctx context.Context, org repository.Organization) (string, error) {
	// Generate new ID if not provided
	if org.ID == "" {
		org.ID = primitive.NewObjectID().Hex()
	}

	// Set timestamps
	org.CreatedAt = nowAsString()
	org.UpdatedAt = nowAsString()

	// Insert document
	result, err := r.getCollection().InsertOne(ctx, org)
	if err != nil {
		return "", err
	}

	// Return ID
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Update updates an existing organization
func (r *MongoOrganizationRepository) Update(ctx context.Context, id string, org repository.Organization) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Set update timestamp
	org.UpdatedAt = nowAsString()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": org}

	result, err := r.getCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// Delete deletes an organization
func (r *MongoOrganizationRepository) Delete(ctx context.Context, id string) error {
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

// FindByName finds an organization by name
func (r *MongoOrganizationRepository) FindByName(ctx context.Context, name string) (repository.Organization, error) {
	filter := bson.M{"name": name}
	var org repository.Organization

	err := r.getCollection().FindOne(ctx, filter).Decode(&org)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return org, repository.ErrNotFound
		}
		return org, err
	}

	return org, nil
}

// FindByDomain finds an organization by domain
func (r *MongoOrganizationRepository) FindByDomain(ctx context.Context, domain string) (repository.Organization, error) {
	filter := bson.M{"domain": domain}
	var org repository.Organization

	err := r.getCollection().FindOne(ctx, filter).Decode(&org)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return org, repository.ErrNotFound
		}
		return org, err
	}

	return org, nil
}

// AddAdminToOrganization adds an admin to an organization
func (r *MongoOrganizationRepository) AddAdminToOrganization(ctx context.Context, orgID, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$addToSet": bson.M{"adminIds": userID},
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

// RemoveAdminFromOrganization removes an admin from an organization
func (r *MongoOrganizationRepository) RemoveAdminFromOrganization(ctx context.Context, orgID, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$pull": bson.M{"adminIds": userID},
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

// IsUserAdmin checks if a user is an admin of an organization
func (r *MongoOrganizationRepository) IsUserAdmin(ctx context.Context, orgID, userID string) (bool, error) {
	org, err := r.FindByID(ctx, orgID)
	if err != nil {
		return false, err
	}

	for _, adminID := range org.AdminIDs {
		if adminID == userID {
			return true, nil
		}
	}

	return false, nil
}
