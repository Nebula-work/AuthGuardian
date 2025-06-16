package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Nebula-work/AuthGuardian/config"
	"github.com/Nebula-work/AuthGuardian/database"
	"github.com/Nebula-work/AuthGuardian/handlers"
	"github.com/Nebula-work/AuthGuardian/middleware"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(app *fiber.App) {
	// Load configuration
	cfg, _ := config.LoadConfig()

	// Get database client
	dbClient, _ := database.ConnectMongoDB(cfg.MongoURI)

	// Create handlers
	authHandler := handlers.NewAuthHandler(cfg, dbClient)
	userHandler := handlers.NewUserHandler(cfg, dbClient)
	orgHandler := handlers.NewOrganizationHandler(cfg, dbClient)
	roleHandler := handlers.NewRoleHandler(cfg, dbClient)
	permissionHandler := handlers.NewPermissionHandler(cfg, dbClient)
	// Auth routes
	auth := app.Group("/api/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/oauth/google", authHandler.GoogleOAuth)
	auth.Get("/oauth/google/callback", authHandler.GoogleOAuthCallback)
	auth.Get("/oauth/github", authHandler.GitHubOAuth)
	auth.Get("/oauth/github/callback", authHandler.GitHubOAuthCallback)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/refresh-token", authHandler.RefreshAccesTokenUsingRefreshToken)

	auth.Post("/logout", middleware.Authenticate(cfg, dbClient), authHandler.Logout)

	// User routes
	users := app.Group("/api/users", middleware.Authenticate(cfg, dbClient))

	users.Get("/", middleware.Authorize(database.UsersCollection, "read", dbClient), userHandler.GetUsers)
	users.Post("/", middleware.Authorize(database.UsersCollection, "create", dbClient), userHandler.CreateUser)
	users.Get("/:id", middleware.Authorize(database.UsersCollection, "read", dbClient), userHandler.GetUser)
	users.Put("/:id", middleware.Authorize(database.UsersCollection, "update", dbClient), userHandler.UpdateUser)
	users.Delete("/:id", middleware.Authorize(database.UsersCollection, "delete", dbClient), userHandler.DeleteUser)
	users.Get("/me", userHandler.GetCurrentUser)
	users.Put("/me", userHandler.UpdateCurrentUser)

	// Organization routes
	orgs := app.Group("/api/organizations", middleware.Authenticate(cfg, dbClient))
	orgs.Get("/", middleware.Authorize(database.OrganizationsCollection, "read", dbClient), orgHandler.GetOrganizations)
	orgs.Post("/", middleware.Authorize(database.OrganizationsCollection, "create", dbClient), orgHandler.CreateOrganization)
	orgs.Get("/:id", middleware.Authorize(database.OrganizationsCollection, "read", dbClient), orgHandler.GetOrganization)
	orgs.Put("/:id", middleware.Authorize(database.OrganizationsCollection, "update", dbClient), orgHandler.UpdateOrganization)
	orgs.Delete("/:id", middleware.Authorize(database.OrganizationsCollection, "delete", dbClient), orgHandler.DeleteOrganization)
	orgs.Post("/:id/users", middleware.Authorize(database.OrganizationsCollection, "update", dbClient), orgHandler.AddUserToOrganization)
	orgs.Delete("/:id/users/:userId", middleware.Authorize(database.OrganizationsCollection, "update", dbClient), orgHandler.RemoveUserFromOrganization)
	orgs.Get("/:id/users", middleware.Authorize(database.OrganizationsCollection, "read", dbClient), orgHandler.GetOrganizationUsers)

	// Role routes
	roles := app.Group("/api/roles", middleware.Authenticate(cfg, dbClient))
	roles.Get("/", middleware.Authorize(database.RolesCollection, "read", dbClient), roleHandler.GetRoles)
	roles.Post("/", middleware.Authorize(database.RolesCollection, "create", dbClient), roleHandler.CreateRole)
	roles.Get("/:id", middleware.Authorize(database.RolesCollection, "read", dbClient), roleHandler.GetRole)
	roles.Put("/:id", middleware.Authorize(database.RolesCollection, "update", dbClient), roleHandler.UpdateRole)
	roles.Delete("/:id", middleware.Authorize(database.RolesCollection, "delete", dbClient), roleHandler.DeleteRole)
	roles.Post("/:id/permissions", middleware.Authorize(database.RolesCollection, "update", dbClient), roleHandler.AddPermissionsToRole)
	roles.Delete("/:id/permissions", middleware.Authorize(database.RolesCollection, "update", dbClient), roleHandler.RemovePermissionsFromRole)
	// Permission routes
	permissions := app.Group("/api/permissions", middleware.Authenticate(cfg, dbClient))
	permissions.Get("/", middleware.Authorize(database.PermissionsCollection, "read", dbClient), permissionHandler.GetPermissions)

	// Frontend assets
	app.Static("/", "./frontend/build")

	// For SPA routing - forward all unmatched routes to index.html
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./frontend/build/index.html")
	})
}
