package mongodb

import (
	"context"
	"errors"
	"rbac-system/core/models"
	"rbac-system/pkg/common/repository"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserInactive          = errors.New("user account is inactive")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidToken          = errors.New("invalid token")
	ErrExpiredToken          = errors.New("expired token")
	ErrTokenNotFound         = errors.New("token not found")
	ErrInvalidSignature      = errors.New("invalid token signature")
)

// MongoTokenRepository implements auth.TokenRepository using MongoDB
type MongoTokenRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// Helper function to get current time as formatted string
func nowAsString() string {
	return time.Now().Format(time.RFC3339)
}

// NewMongoTokenRepository creates a new MongoDB token repository
func NewMongoTokenRepository(client *mongo.Client, database, collection string) repository.TokenRepository {
	return &MongoTokenRepository{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// getCollection returns the MongoDB collection for tokens
func (r *MongoTokenRepository) getCollection() *mongo.Collection {
	return r.client.Database(r.database).Collection(r.collection)
}

// IsConnected checks if the repository is connected to the database
func (r *MongoTokenRepository) IsConnected(ctx context.Context) bool {
	return r.client != nil && r.client.Ping(ctx, nil) == nil
}

// StoreToken stores a token
func (r *MongoTokenRepository) StoreToken(ctx context.Context, token models.Token) error {
	// Generate ID if not provided
	if token.ID == "" {
		token.ID = uuid.New().String()
	}

	// Set created at if not provided
	if token.CreatedAt == "" {
		token.CreatedAt = nowAsString()
	}

	// Insert token
	_, err := r.getCollection().InsertOne(ctx, token)
	return err
}

// FindTokenByValue finds a token by its value
func (r *MongoTokenRepository) FindTokenByValue(ctx context.Context, tokenType models.TokenType, tokenValue string) (models.Token, error) {
	// Find token
	var token models.Token
	filter := bson.M{"tokenType": tokenType, "tokenValue": tokenValue}

	err := r.getCollection().FindOne(ctx, filter).Decode(&token)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Token{}, ErrTokenNotFound
		}
		return models.Token{}, err
	}

	// Check if token is expired
	if token.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt)
		if err == nil && expiresAt.Before(time.Now()) {
			// Token is expired, delete it
			_ = r.DeleteToken(ctx, tokenType, tokenValue)
			return models.Token{}, ErrExpiredToken
		}
	}

	return token, nil
}

// FindTokensByUser finds all tokens for a user
func (r *MongoTokenRepository) FindTokensByUser(ctx context.Context, tokenType models.TokenType, userID string) ([]models.Token, error) {
	// Find tokens
	filter := bson.M{"tokenType": tokenType, "userId": userID}
	cursor, err := r.getCollection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode tokens
	var tokens []models.Token
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// DeleteToken deletes a token
func (r *MongoTokenRepository) DeleteToken(ctx context.Context, tokenType models.TokenType, tokenValue string) error {
	// Delete token
	filter := bson.M{"tokenType": tokenType, "tokenValue": tokenValue}
	_, err := r.getCollection().DeleteOne(ctx, filter)
	return err
}

// DeleteTokensByUser deletes all tokens for a user
func (r *MongoTokenRepository) DeleteTokensByUser(ctx context.Context, tokenType models.TokenType, userID string) error {
	// Delete tokens
	filter := bson.M{"tokenType": tokenType, "userId": userID}
	_, err := r.getCollection().DeleteMany(ctx, filter)
	return err
}

// DeleteExpiredTokens deletes all expired tokens
func (r *MongoTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	// Get current time
	now := time.Now().Format(time.RFC3339)

	// Delete tokens
	filter := bson.M{"expiresAt": bson.M{"$lt": now}}
	_, err := r.getCollection().DeleteMany(ctx, filter)
	return err
}

// InMemoryTokenRepository implements auth.TokenRepository using in-memory storage
type InMemoryTokenRepository struct {
	tokens map[string]models.Token // key is tokenValue
}

//// NewInMemoryTokenRepository creates a new in-memory token repository
//func NewInMemoryTokenRepository() *InMemoryTokenRepository{
//	return &InMemoryTokenRepository{
//		tokens: make(map[string]models.Token),
//	}
//}
//
//// IsConnected checks if the repository is connected
//func (r *InMemoryTokenRepository) IsConnected(ctx context.Context) bool {
//	return true
//}
//
//// StoreToken stores a token
//func (r *InMemoryTokenRepository) StoreToken(ctx context.Context, token models.Token) error {
//	// Generate ID if not provided
//	if token.ID == "" {
//		token.ID = uuid.New().String()
//	}
//
//	// Set created at if not provided
//	if token.CreatedAt == "" {
//		token.CreatedAt = nowAsString()
//	}
//
//	// Store token
//	r.tokens[token.TokenValue] = token
//
//	return nil
//}
//
//// FindTokenByValue finds a token by its value
//func (r *InMemoryTokenRepository) FindTokenByValue(ctx context.Context, tokenType models.TokenType, tokenValue string) (models.Token, error) {
//	// Find token
//	token, ok := r.tokens[tokenValue]
//	if !ok || token.TokenType != tokenType {
//		return models.Token{}, auth.ErrTokenNotFound
//	}
//
//	// Check if token is expired
//	if token.ExpiresAt != "" {
//		expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt)
//		if err == nil && expiresAt.Before(time.Now()) {
//			// Token is expired, delete it
//			delete(r.tokens, tokenValue)
//			return models.Token{}, auth.ErrExpiredToken
//		}
//	}
//
//	return token, nil
//}
//
//// FindTokensByUser finds all tokens for a user
//func (r *InMemoryTokenRepository) FindTokensByUser(ctx context.Context, tokenType models.TokenType, userID string) ([]models.Token, error) {
//	// Find tokens
//	var tokens []models.Token
//	for _, token := range r.tokens {
//		if token.TokenType == tokenType && token.UserID == userID {
//			tokens = append(tokens, token)
//		}
//	}
//
//	return tokens, nil
//}
//
//// DeleteToken deletes a token
//func (r *InMemoryTokenRepository) DeleteToken(ctx context.Context, tokenType models.TokenType, tokenValue string) error {
//	// Check if token exists
//	token, ok := r.tokens[tokenValue]
//	if !ok || token.TokenType != tokenType {
//		return nil // No error if token doesn't exist
//	}
//
//	// Delete token
//	delete(r.tokens, tokenValue)
//
//	return nil
//}
//
//// DeleteTokensByUser deletes all tokens for a user
//func (r *InMemoryTokenRepository) DeleteTokensByUser(ctx context.Context, tokenType models.TokenType, userID string) error {
//	// Find tokens to delete
//	var tokensToDelete []string
//	for tokenValue, token := range r.tokens {
//		if token.TokenType == tokenType && token.UserID == userID {
//			tokensToDelete = append(tokensToDelete, tokenValue)
//		}
//	}
//
//	// Delete tokens
//	for _, tokenValue := range tokensToDelete {
//		delete(r.tokens, tokenValue)
//	}
//
//	return nil
//}
//
//// DeleteExpiredTokens deletes all expired tokens
//func (r *InMemoryTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
//	// Get current time
//	now := time.Now()
//
//	// Find tokens to delete
//	var tokensToDelete []string
//	for tokenValue, token := range r.tokens {
//		if token.ExpiresAt != "" {
//			expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt)
//			if err == nil && expiresAt.Before(now) {
//				tokensToDelete = append(tokensToDelete, tokenValue)
//			}
//		}
//	}
//
//	// Delete tokens
//	for _, tokenValue := range tokensToDelete {
//		delete(r.tokens, tokenValue)
//	}
//
//	return nil
//}
