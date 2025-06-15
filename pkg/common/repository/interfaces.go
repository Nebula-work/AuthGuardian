package repository

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Common errors
var (
	ErrNotFound          = errors.New("entity not found")
	ErrDuplicateKey      = errors.New("duplicate key error")
	ErrInvalidID         = errors.New("invalid id")
	ErrConnectionFailed  = errors.New("database connection failed")
	ErrTransactionFailed = errors.New("transaction failed")
)

// Filter represents a map of field names to values for filtering
type Filter map[string]interface{}

// QueryOptions represents options for querying data
type QueryOptions struct {
	Skip  int64
	Limit int64
	Sort  map[string]int // field name -> sort order (1 for asc, -1 for desc)
}

func ToMongoOptions(queryOptions QueryOptions) *options.FindOptions {
	findOptions := options.Find()

	// Set skip value
	if queryOptions.Skip > 0 {
		findOptions.SetSkip(queryOptions.Skip)
	}

	// Set limit value
	if queryOptions.Limit > 0 {
		findOptions.SetLimit(queryOptions.Limit)
	}

	// Set sort options
	if len(queryOptions.Sort) > 0 {
		sort := bson.D{}
		for field, order := range queryOptions.Sort {
			sort = append(sort, bson.E{Key: field, Value: order})
		}
		findOptions.SetSort(sort)
	}

	return findOptions
}

// Repository defines a generic repository interface
type Repository[T any] interface {
	// IsConnected checks if the repository is connected to the database
	IsConnected(ctx context.Context) bool

	// FindByID finds an entity by ID
	FindByID(ctx context.Context, id string) (T, error)

	// FindOne finds a single entity matching the filter
	FindOne(ctx context.Context, filter Filter) (T, error)

	// FindMany finds multiple entities matching the filter
	FindMany(ctx context.Context, filter Filter, options QueryOptions) ([]T, error)

	// Count counts entities matching the filter
	Count(ctx context.Context, filter Filter) (int64, error)

	// Create creates a new entity
	Create(ctx context.Context, entity T) (string, error)

	// Update updates an existing entity
	Update(ctx context.Context, id string, entity T) error

	// Delete deletes an entity
	Delete(ctx context.Context, id string) error
}

// TransactionManager defines an interface for managing database transactions
type TransactionManager interface {
	// BeginTransaction begins a new transaction
	BeginTransaction(ctx context.Context) (context.Context, error)

	// CommitTransaction commits a transaction
	CommitTransaction(ctx context.Context) error

	// RollbackTransaction rolls back a transaction
	RollbackTransaction(ctx context.Context) error
}

// RepositoryFactory defines an interface for creating repositories
type RepositoryFactory interface {
	// GetUserRepository returns a user repository
	GetUserRepository() interface{}

	// GetRoleRepository returns a role repository
	GetRoleRepository() interface{}

	// GetPermissionRepository returns a permission repository
	GetPermissionRepository() interface{}

	// GetOrganizationRepository returns an organization repository
	GetOrganizationRepository() interface{}
}

// MongoDB-specific interfaces

// MongoIndexCreator defines an interface for creating MongoDB indexes
type MongoIndexCreator interface {
	// CreateIndexes creates indexes for the collection
	CreateIndexes(ctx context.Context) error
}

// PostgreSQL-specific interfaces

// PostgreSQLMigrator defines an interface for performing PostgreSQL migrations
type PostgreSQLMigrator interface {
	// MigrateUp performs database migrations up
	MigrateUp(ctx context.Context) error

	// MigrateDown performs database migrations down
	MigrateDown(ctx context.Context) error
}
