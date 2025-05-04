package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rbac-system/config"
	"rbac-system/models"
	"rbac-system/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping the MongoDB server to verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB successfully")

	// Initialize database
	db := client.Database(cfg.DatabaseName)

	// Create collections
	usersCollection := db.Collection("users")
	rolesCollection := db.Collection("roles")
	permissionsCollection := db.Collection("permissions")
	organizationsCollection := db.Collection("organizations")

	// Initialize permissions
	initializePermissions(ctx, permissionsCollection)

	// Initialize roles
	initializeRoles(ctx, rolesCollection, permissionsCollection)

	// Create admin user if not exists
	createAdminUser(ctx, usersCollection, rolesCollection)

	log.Println("Database initialization completed")
}

func initializePermissions(ctx context.Context, permissionsCollection *mongo.Collection) {
	log.Println("Initializing permissions...")

	// Define system permissions
	permissions := []struct {
		name        string
		description string
		resource    string
		action      string
	}{
		// User permissions
		{"user.create", "Create users", models.ResourceUser, models.ActionCreate},
		{"user.read", "Read users", models.ResourceUser, models.ActionRead},
		{"user.update", "Update users", models.ResourceUser, models.ActionUpdate},
		{"user.delete", "Delete users", models.ResourceUser, models.ActionDelete},

		// Role permissions
		{"role.create", "Create roles", models.ResourceRole, models.ActionCreate},
		{"role.read", "Read roles", models.ResourceRole, models.ActionRead},
		{"role.update", "Update roles", models.ResourceRole, models.ActionUpdate},
		{"role.delete", "Delete roles", models.ResourceRole, models.ActionDelete},

		// Organization permissions
		{"organization.create", "Create organizations", models.ResourceOrganization, models.ActionCreate},
		{"organization.read", "Read organizations", models.ResourceOrganization, models.ActionRead},
		{"organization.update", "Update organizations", models.ResourceOrganization, models.ActionUpdate},
		{"organization.delete", "Delete organizations", models.ResourceOrganization, models.ActionDelete},

		// Permission permissions
		{"permission.create", "Create permissions", models.ResourcePermission, models.ActionCreate},
		{"permission.read", "Read permissions", models.ResourcePermission, models.ActionRead},
		{"permission.update", "Update permissions", models.ResourcePermission, models.ActionUpdate},
		{"permission.delete", "Delete permissions", models.ResourcePermission, models.ActionDelete},

		// All permissions
		{"all", "All permissions", models.ResourceAll, models.ActionAll},
	}

	// Check if permissions already exist
	count, err := permissionsCollection.CountDocuments(ctx, bson.M{"isSystemDefault": true})
	if err != nil {
		log.Fatalf("Failed to count permissions: %v", err)
	}

	if count == 0 {
		// Create permissions
		var permissionDocs []interface{}
		for _, p := range permissions {
			permissionDocs = append(permissionDocs, models.Permission{
				Name:           p.name,
				Description:    p.description,
				Resource:       p.resource,
				Action:         p.action,
				IsSystemDefault: true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			})
		}

		_, err := permissionsCollection.InsertMany(ctx, permissionDocs)
		if err != nil {
			log.Fatalf("Failed to insert permissions: %v", err)
		}

		log.Printf("Inserted %d default permissions", len(permissionDocs))
	} else {
		log.Println("Default permissions already exist, skipping creation")
	}
}

func initializeRoles(ctx context.Context, rolesCollection, permissionsCollection *mongo.Collection) {
	log.Println("Initializing roles...")

	// Check if default roles already exist
	count, err := rolesCollection.CountDocuments(ctx, bson.M{"isSystemDefault": true})
	if err != nil {
		log.Fatalf("Failed to count roles: %v", err)
	}

	if count > 0 {
		log.Println("Default roles already exist, skipping creation")
		return
	}

	// Get all permissions
	var permissions []models.Permission
	cursor, err := permissionsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to retrieve permissions: %v", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &permissions); err != nil {
		log.Fatalf("Failed to decode permissions: %v", err)
	}

	// Get permission IDs
	var allPermissionIDs []primitive.ObjectID
	var userReadPermissionID primitive.ObjectID
	var userUpdatePermissionID primitive.ObjectID
	var orgReadPermissionID primitive.ObjectID
	var roleReadPermissionID primitive.ObjectID

	for _, perm := range permissions {
		allPermissionIDs = append(allPermissionIDs, perm.ID)
		
		// Store specific permission IDs for regular user role
		if perm.Name == "user.read" {
			userReadPermissionID = perm.ID
		} else if perm.Name == "user.update" {
			userUpdatePermissionID = perm.ID
		} else if perm.Name == "organization.read" {
			orgReadPermissionID = perm.ID
		} else if perm.Name == "role.read" {
			roleReadPermissionID = perm.ID
		}
	}

	// Define default roles
	roles := []models.Role{
		{
			Name:            models.SystemAdminRole,
			Description:     "System administrator with full access",
			PermissionIDs:   allPermissionIDs,
			IsSystemDefault: true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Name:            models.OrganizationAdminRole,
			Description:     "Organization administrator with full access to organization resources",
			PermissionIDs:   getOrgAdminPermissions(permissions),
			IsSystemDefault: true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			Name:            models.UserRole,
			Description:     "Regular user with limited access",
			PermissionIDs:   []primitive.ObjectID{userReadPermissionID, userUpdatePermissionID, orgReadPermissionID, roleReadPermissionID},
			IsSystemDefault: true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Insert roles
	var roleDocs []interface{}
	for _, role := range roles {
		roleDocs = append(roleDocs, role)
	}

	_, err = rolesCollection.InsertMany(ctx, roleDocs)
	if err != nil {
		log.Fatalf("Failed to insert roles: %v", err)
	}

	log.Printf("Inserted %d default roles", len(roleDocs))
}

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

func createAdminUser(ctx context.Context, usersCollection, rolesCollection *mongo.Collection) {
	log.Println("Creating admin user if not exists...")

	// Check if admin user already exists
	count, err := usersCollection.CountDocuments(ctx, bson.M{"username": "admin"})
	if err != nil {
		log.Fatalf("Failed to check for admin user: %v", err)
	}

	if count > 0 {
		log.Println("Admin user already exists, skipping creation")
		return
	}

	// Get the system admin role
	var adminRole models.Role
	err = rolesCollection.FindOne(ctx, bson.M{"name": models.SystemAdminRole}).Decode(&adminRole)
	if err != nil {
		log.Fatalf("Failed to retrieve admin role: %v", err)
	}

	// Create admin password
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	adminUser := models.User{
		Username:       "admin",
		Email:          "admin@example.com",
		HashedPassword: hashedPassword,
		FirstName:      "System",
		LastName:       "Administrator",
		Active:         true,
		EmailVerified:  true,
		RoleIDs:        []primitive.ObjectID{adminRole.ID},
		AuthProvider:   models.LocalAuth,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = usersCollection.InsertOne(ctx, adminUser)
	if err != nil {
		log.Fatalf("Failed to insert admin user: %v", err)
	}

	log.Println("Created admin user with username 'admin' and password 'admin123'")
}
