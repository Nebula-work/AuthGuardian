package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps the MongoDB client and provides utility methods
type MongoClient struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection map[string]*mongo.Collection
}

// Collections in the database
const (
	UsersCollection         = "users"
	RolesCollection         = "roles"
	OrganizationsCollection = "organizations"
	PermissionsCollection   = "permissions"
)

// ConnectMongoDB connects to MongoDB and returns a client
func ConnectMongoDB(uri string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Initialize database and collections
	database := client.Database("rbac_system")
	collections := make(map[string]*mongo.Collection)
	collections[UsersCollection] = database.Collection(UsersCollection)
	collections[RolesCollection] = database.Collection(RolesCollection)
	collections[OrganizationsCollection] = database.Collection(OrganizationsCollection)
	collections[PermissionsCollection] = database.Collection(PermissionsCollection)

	return &MongoClient{
		client:     client,
		database:   database,
		collection: collections,
	}, nil
}

// GetCollection returns a MongoDB collection
func (m *MongoClient) GetCollection(name string) *mongo.Collection {
	return m.collection[name]
}

// Disconnect closes the MongoDB connection
func (m *MongoClient) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

// GetDatabase returns the MongoDB database
func (m *MongoClient) GetDatabase() *mongo.Database {
	return m.database
}
