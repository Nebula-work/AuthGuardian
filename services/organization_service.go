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
)

// OrganizationService provides organization-related operations
type OrganizationService struct {
	db *database.MongoClient
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *database.MongoClient) *OrganizationService {
	return &OrganizationService{
		db: db,
	}
}

// GetOrganizations retrieves organizations with pagination
func (s *OrganizationService) GetOrganizations(limit, skip int) ([]models.Organization, int64, error) {
	// Get organizations collection
	collection := s.db.GetCollection(database.OrganizationsCollection)
	
	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"name": 1})
	
	// Execute query
	cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve organizations")
	}
	defer cursor.Close(context.Background())
	
	// Decode organizations
	var organizations []models.Organization
	if err := cursor.All(context.Background(), &organizations); err != nil {
		return nil, 0, errors.New("failed to decode organizations")
	}
	
	// Get total count
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, 0, errors.New("failed to count organizations")
	}
	
	return organizations, count, nil
}

// GetOrganizationByID retrieves an organization by ID
func (s *OrganizationService) GetOrganizationByID(id string) (*models.Organization, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid organization ID")
	}
	
	// Get organization from database
	collection := s.db.GetCollection(database.OrganizationsCollection)
	org := models.Organization{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&org)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("organization not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve organization")
	}
	
	return &org, nil
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(input models.OrganizationCreateInput, currentUserID primitive.ObjectID) (*models.Organization, error) {
	// Validate required fields
	if input.Name == "" {
		return nil, errors.New("organization name is required")
	}
	
	// Check if organization name already exists
	collection := s.db.GetCollection(database.OrganizationsCollection)
	existingOrg := models.Organization{}
	
	err := collection.FindOne(
		context.Background(),
		bson.M{"name": input.Name},
	).Decode(&existingOrg)
	
	if err == nil {
		return nil, errors.New("organization name already exists")
	} else if err != mongo.ErrNoDocuments {
		return nil, errors.New("failed to check for existing organization")
	}
	
	// If domain is provided, check for uniqueness
	if input.Domain != "" {
		err = collection.FindOne(
			context.Background(),
			bson.M{"domain": input.Domain},
		).Decode(&existingOrg)
		
		if err == nil {
			return nil, errors.New("domain already registered for another organization")
		} else if err != mongo.ErrNoDocuments {
			return nil, errors.New("failed to check for existing domain")
		}
	}
	
	// Get current user as admin if no admins provided
	var adminIDs []primitive.ObjectID
	if len(input.AdminIDs) > 0 {
		adminIDs = input.AdminIDs
	} else if !currentUserID.IsZero() {
		adminIDs = append(adminIDs, currentUserID)
	}
	
	// Create organization
	org := models.Organization{
		Name:        input.Name,
		Description: input.Description,
		Domain:      input.Domain,
		Active:      true,
		AdminIDs:    adminIDs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Insert organization
	result, err := collection.InsertOne(context.Background(), org)
	if err != nil {
		return nil, errors.New("failed to create organization")
	}
	
	// Get the inserted ID
	org.ID = result.InsertedID.(primitive.ObjectID)
	
	// Update admin users with this organization
	if len(adminIDs) > 0 {
		usersCollection := s.db.GetCollection(database.UsersCollection)
		_, err = usersCollection.UpdateMany(
			context.Background(),
			bson.M{"_id": bson.M{"$in": adminIDs}},
			bson.M{"$addToSet": bson.M{"organizationIds": org.ID}},
		)
		
		if err != nil {
			// Log error but continue
		}
	}
	
	return &org, nil
}

// UpdateOrganization updates an existing organization
func (s *OrganizationService) UpdateOrganization(id string, input models.OrganizationUpdateInput) (*models.Organization, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid organization ID")
	}
	
	// Check if organization exists
	collection := s.db.GetCollection(database.OrganizationsCollection)
	existingOrg := models.Organization{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingOrg)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("organization not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve organization")
	}
	
	// Prepare update fields
	update := bson.M{"updatedAt": time.Now()}
	
	if input.Name != "" {
		// Check if name is already taken by another organization
		if input.Name != existingOrg.Name {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"name": input.Name, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return nil, errors.New("failed to check name availability")
			}
			if count > 0 {
				return nil, errors.New("organization name already taken")
			}
		}
		update["name"] = input.Name
	}
	
	if input.Description != "" {
		update["description"] = input.Description
	}
	
	if input.Domain != "" {
		// Check if domain is already taken by another organization
		if input.Domain != existingOrg.Domain {
			count, err := collection.CountDocuments(
				context.Background(),
				bson.M{"domain": input.Domain, "_id": bson.M{"$ne": objectID}},
			)
			if err != nil {
				return nil, errors.New("failed to check domain availability")
			}
			if count > 0 {
				return nil, errors.New("domain already registered for another organization")
			}
		}
		update["domain"] = input.Domain
	}
	
	if input.Active != nil {
		update["active"] = *input.Active
	}
	
	if len(input.AdminIDs) > 0 {
		update["adminIds"] = input.AdminIDs
	}
	
	// Update organization
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	
	if err != nil {
		return nil, errors.New("failed to update organization")
	}
	
	// Get updated organization
	updatedOrg := models.Organization{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedOrg)
	
	if err != nil {
		return nil, errors.New("failed to retrieve updated organization")
	}
	
	return &updatedOrg, nil
}

// DeleteOrganization deletes an organization
func (s *OrganizationService) DeleteOrganization(id string) error {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid organization ID")
	}
	
	// Delete organization
	collection := s.db.GetCollection(database.OrganizationsCollection)
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)
	
	if err != nil {
		return errors.New("failed to delete organization")
	}
	
	if result.DeletedCount == 0 {
		return errors.New("organization not found")
	}
	
	// Update users by removing this organization
	usersCollection := s.db.GetCollection(database.UsersCollection)
	_, err = usersCollection.UpdateMany(
		context.Background(),
		bson.M{"organizationIds": objectID},
		bson.M{"$pull": bson.M{"organizationIds": objectID}},
	)
	
	if err != nil {
		// Log error but continue
	}
	
	return nil
}

// AddUserToOrganization adds a user to an organization
func (s *OrganizationService) AddUserToOrganization(orgID, userID string, roleIDs []primitive.ObjectID) error {
	// Convert string IDs to ObjectIDs
	orgObjectID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return errors.New("invalid organization ID")
	}
	
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	
	// Check if organization exists
	orgCollection := s.db.GetCollection(database.OrganizationsCollection)
	org := models.Organization{}
	
	err = orgCollection.FindOne(
		context.Background(),
		bson.M{"_id": orgObjectID},
	).Decode(&org)
	
	if err == mongo.ErrNoDocuments {
		return errors.New("organization not found")
	} else if err != nil {
		return errors.New("failed to retrieve organization")
	}
	
	// Check if user exists
	userCollection := s.db.GetCollection(database.UsersCollection)
	user := models.User{}
	
	err = userCollection.FindOne(
		context.Background(),
		bson.M{"_id": userObjectID},
	).Decode(&user)
	
	if err == mongo.ErrNoDocuments {
		return errors.New("user not found")
	} else if err != nil {
		return errors.New("failed to retrieve user")
	}
	
	// Update user's organization list
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$addToSet": bson.M{"organizationIds": orgObjectID}},
	)
	
	if err != nil {
		return errors.New("failed to add user to organization")
	}
	
	// If role IDs are provided, update user's roles
	if len(roleIDs) > 0 {
		_, err = userCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": userObjectID},
			bson.M{"$addToSet": bson.M{"roleIds": bson.M{"$each": roleIDs}}},
		)
		
		if err != nil {
			return errors.New("failed to update user roles")
		}
	}
	
	return nil
}

// RemoveUserFromOrganization removes a user from an organization
func (s *OrganizationService) RemoveUserFromOrganization(orgID, userID string) error {
	// Convert string IDs to ObjectIDs
	orgObjectID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return errors.New("invalid organization ID")
	}
	
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	
	// Check if user exists and belongs to the organization
	userCollection := s.db.GetCollection(database.UsersCollection)
	user := models.User{}
	
	err = userCollection.FindOne(
		context.Background(),
		bson.M{"_id": userObjectID, "organizationIds": orgObjectID},
	).Decode(&user)
	
	if err == mongo.ErrNoDocuments {
		return errors.New("user not found or not a member of the organization")
	} else if err != nil {
		return errors.New("failed to retrieve user")
	}
	
	// Update user's organization list
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$pull": bson.M{"organizationIds": orgObjectID}},
	)
	
	if err != nil {
		return errors.New("failed to remove user from organization")
	}
	
	return nil
}
