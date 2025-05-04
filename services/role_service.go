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

// RoleService provides role-related operations
type RoleService struct {
	db *database.MongoClient
}

// NewRoleService creates a new role service
func NewRoleService(db *database.MongoClient) *RoleService {
	return &RoleService{
		db: db,
	}
}

// GetRoles retrieves roles with pagination
func (s *RoleService) GetRoles(limit, skip int, orgID string) ([]models.Role, int64, error) {
	// Prepare query
	query := bson.M{}
	
	// Add organization filter if provided
	if orgID != "" {
		objectID, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			return nil, 0, errors.New("invalid organization ID")
		}
		query["organizationId"] = objectID
	}
	
	// Get roles collection
	collection := s.db.GetCollection(database.RolesCollection)
	
	// Set options
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"name": 1})
	
	// Execute query
	cursor, err := collection.Find(context.Background(), query, findOptions)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve roles")
	}
	defer cursor.Close(context.Background())
	
	// Decode roles
	var roles []models.Role
	if err := cursor.All(context.Background(), &roles); err != nil {
		return nil, 0, errors.New("failed to decode roles")
	}
	
	// Get total count
	count, err := collection.CountDocuments(context.Background(), query)
	if err != nil {
		return nil, 0, errors.New("failed to count roles")
	}
	
	return roles, count, nil
}

// GetRoleByID retrieves a role by ID
func (s *RoleService) GetRoleByID(id string) (*models.Role, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid role ID")
	}
	
	// Get role from database
	collection := s.db.GetCollection(database.RolesCollection)
	role := models.Role{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&role)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("role not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve role")
	}
	
	return &role, nil
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(input models.RoleCreateInput) (*models.Role, error) {
	// Validate required fields
	if input.Name == "" {
		return nil, errors.New("role name is required")
	}
	
	// Check if role name already exists in the same organization
	collection := s.db.GetCollection(database.RolesCollection)
	
	// Query to check for role with same name in same organization or system default
	query := bson.M{"name": input.Name}
	if !input.OrganizationID.IsZero() {
		query["$or"] = []bson.M{
			{"organizationId": input.OrganizationID},
			{"isSystemDefault": true},
		}
	} else {
		query["isSystemDefault"] = true
	}
	
	existingRole := models.Role{}
	err := collection.FindOne(context.Background(), query).Decode(&existingRole)
	
	if err == nil {
		return nil, errors.New("role name already exists in this organization or as a system default")
	} else if err != mongo.ErrNoDocuments {
		return nil, errors.New("failed to check for existing role")
	}
	
	// Create role
	role := models.Role{
		Name:           input.Name,
		Description:    input.Description,
		OrganizationID: input.OrganizationID,
		PermissionIDs:  input.PermissionIDs,
		IsSystemDefault: false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Insert role
	result, err := collection.InsertOne(context.Background(), role)
	if err != nil {
		return nil, errors.New("failed to create role")
	}
	
	// Get the inserted ID
	role.ID = result.InsertedID.(primitive.ObjectID)
	
	return &role, nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(id string, input models.RoleUpdateInput) (*models.Role, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid role ID")
	}
	
	// Check if role exists
	collection := s.db.GetCollection(database.RolesCollection)
	existingRole := models.Role{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingRole)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("role not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve role")
	}
	
	// Prevent updating system default roles
	if existingRole.IsSystemDefault {
		return nil, errors.New("system default roles cannot be modified")
	}
	
	// Prepare update fields
	update := bson.M{"updatedAt": time.Now()}
	
	if input.Name != "" {
		// Check if name is already taken by another role in the same organization
		if input.Name != existingRole.Name {
			query := bson.M{
				"name":        input.Name,
				"_id":         bson.M{"$ne": objectID},
			}
			
			if !existingRole.OrganizationID.IsZero() {
				query["organizationId"] = existingRole.OrganizationID
			}
			
			count, err := collection.CountDocuments(context.Background(), query)
			if err != nil {
				return nil, errors.New("failed to check name availability")
			}
			if count > 0 {
				return nil, errors.New("role name already taken in this organization")
			}
		}
		update["name"] = input.Name
	}
	
	if input.Description != "" {
		update["description"] = input.Description
	}
	
	if len(input.PermissionIDs) > 0 {
		update["permissionIds"] = input.PermissionIDs
	}
	
	// Update role
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	
	if err != nil {
		return nil, errors.New("failed to update role")
	}
	
	// Get updated role
	updatedRole := models.Role{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedRole)
	
	if err != nil {
		return nil, errors.New("failed to retrieve updated role")
	}
	
	return &updatedRole, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(id string) error {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid role ID")
	}
	
	// Check if role exists and is not a system default
	collection := s.db.GetCollection(database.RolesCollection)
	existingRole := models.Role{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&existingRole)
	
	if err == mongo.ErrNoDocuments {
		return errors.New("role not found")
	} else if err != nil {
		return errors.New("failed to retrieve role")
	}
	
	// Prevent deleting system default roles
	if existingRole.IsSystemDefault {
		return errors.New("system default roles cannot be deleted")
	}
	
	// Delete role
	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)
	
	if err != nil {
		return errors.New("failed to delete role")
	}
	
	if result.DeletedCount == 0 {
		return errors.New("role not found")
	}
	
	// Update users by removing this role from their role list
	usersCollection := s.db.GetCollection(database.UsersCollection)
	_, err = usersCollection.UpdateMany(
		context.Background(),
		bson.M{"roleIds": objectID},
		bson.M{"$pull": bson.M{"roleIds": objectID}},
	)
	
	if err != nil {
		// Log error but continue
	}
	
	return nil
}

// AddPermissionsToRole adds permissions to a role
func (s *RoleService) AddPermissionsToRole(id string, permissionIDs []primitive.ObjectID) (*models.Role, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid role ID")
	}
	
	// Check if role exists
	collection := s.db.GetCollection(database.RolesCollection)
	role := models.Role{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&role)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("role not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve role")
	}
	
	// Check if permissions exist
	permissionsCollection := s.db.GetCollection(database.PermissionsCollection)
	var permissionCount int64
	permissionCount, err = permissionsCollection.CountDocuments(
		context.Background(),
		bson.M{"_id": bson.M{"$in": permissionIDs}},
	)
	
	if err != nil {
		return nil, errors.New("failed to check permissions")
	}
	
	if int(permissionCount) != len(permissionIDs) {
		return nil, errors.New("one or more permissions do not exist")
	}
	
	// Add permissions to role
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$addToSet": bson.M{
				"permissionIds": bson.M{"$each": permissionIDs},
			},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)
	
	if err != nil {
		return nil, errors.New("failed to add permissions to role")
	}
	
	// Get updated role
	updatedRole := models.Role{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedRole)
	
	if err != nil {
		return nil, errors.New("failed to retrieve updated role")
	}
	
	return &updatedRole, nil
}

// RemovePermissionsFromRole removes permissions from a role
func (s *RoleService) RemovePermissionsFromRole(id string, permissionIDs []primitive.ObjectID) (*models.Role, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid role ID")
	}
	
	// Check if role exists
	collection := s.db.GetCollection(database.RolesCollection)
	role := models.Role{}
	
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&role)
	
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("role not found")
	} else if err != nil {
		return nil, errors.New("failed to retrieve role")
	}
	
	// Remove permissions from role
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$pull": bson.M{
				"permissionIds": bson.M{"$in": permissionIDs},
			},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)
	
	if err != nil {
		return nil, errors.New("failed to remove permissions from role")
	}
	
	// Get updated role
	updatedRole := models.Role{}
	err = collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedRole)
	
	if err != nil {
		return nil, errors.New("failed to retrieve updated role")
	}
	
	return &updatedRole, nil
}

// InitializeDefaultRoles creates default system roles if they don't exist
func (s *RoleService) InitializeDefaultRoles() error {
	collection := s.db.GetCollection(database.RolesCollection)
	
	// Check if default roles already exist
	count, err := collection.CountDocuments(
		context.Background(),
		bson.M{"isSystemDefault": true},
	)
	
	if err != nil {
		return errors.New("failed to check existing default roles")
	}
	
	// If default roles exist, skip initialization
	if count > 0 {
		return nil
	}
	
	// Get all permissions for admin role
	permissionsCollection := s.db.GetCollection(database.PermissionsCollection)
	cursor, err := permissionsCollection.Find(
		context.Background(),
		bson.M{},
	)
	
	if err != nil {
		return errors.New("failed to retrieve permissions")
	}
	defer cursor.Close(context.Background())
	
	var permissions []models.Permission
	if err := cursor.All(context.Background(), &permissions); err != nil {
		return errors.New("failed to decode permissions")
	}
	
	var allPermissionIDs []primitive.ObjectID
	for _, perm := range permissions {
		allPermissionIDs = append(allPermissionIDs, perm.ID)
	}
	
	// Create default roles
	
	// System Admin role - has all permissions
	systemAdminRole := models.Role{
		Name:            models.SystemAdminRole,
		Description:     "System-wide administrator with full access",
		PermissionIDs:   allPermissionIDs,
		IsSystemDefault: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// Organization Admin role - has org-specific permissions
	orgAdminRole := models.Role{
		Name:            models.OrganizationAdminRole,
		Description:     "Organization administrator with full access to organization resources",
		PermissionIDs:   getOrgAdminPermissions(permissions),
		IsSystemDefault: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// User role - has basic user permissions
	userRole := models.Role{
		Name:            models.UserRole,
		Description:     "Regular user with limited access",
		PermissionIDs:   getUserPermissions(permissions),
		IsSystemDefault: true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// Insert default roles
	_, err = collection.InsertMany(
		context.Background(),
		[]interface{}{systemAdminRole, orgAdminRole, userRole},
	)
	
	if err != nil {
		return errors.New("failed to create default roles")
	}
	
	return nil
}

// Helper function to get organization admin permissions
func getOrgAdminPermissions(permissions []models.Permission) []primitive.ObjectID {
	var permIDs []primitive.ObjectID
	
	for _, perm := range permissions {
		// Include all org-related permissions and user management within org
		if perm.Resource == models.ResourceOrganization || 
		   perm.Resource == models.ResourceUser || 
		   perm.Resource == models.ResourceRole {
			permIDs = append(permIDs, perm.ID)
		}
	}
	
	return permIDs
}

// Helper function to get regular user permissions
func getUserPermissions(permissions []models.Permission) []primitive.ObjectID {
	var permIDs []primitive.ObjectID
	
	for _, perm := range permissions {
		// Include only read permissions and self-update
		if perm.Action == models.ActionRead || 
		   (perm.Resource == models.ResourceUser && perm.Action == models.ActionUpdate) {
			permIDs = append(permIDs, perm.ID)
		}
	}
	
	return permIDs
}
