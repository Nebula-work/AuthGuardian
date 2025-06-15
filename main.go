package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"rbac-system/core/auth"
	"rbac-system/core/identity"
	"rbac-system/core/rbac"
	"rbac-system/database"
	"rbac-system/database/mongodb"
	"rbac-system/middleware"
	"rbac-system/pkg/common/repository"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"rbac-system/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Use default CORS if config failed to load
	corsAllowOrigins := "*"
	if cfg != nil {
		corsAllowOrigins = cfg.CORSAllowOrigins
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsAllowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	var userRepo repository.UserRepository
	var orgRepo repository.OrganizationRepository
	var roleRepo repository.RoleRepository
	var permRepo repository.PermissionRepository
	var tokenRepo repository.TokenRepository
	// Check if MongoDB connection is available
	mongoClient, mongoErr := database.ConnectMongoDB(os.Getenv("MONGO_URI"))
	if mongoErr != nil {
		//log.Printf("Warning: MongoDB connection failed: %v", mongoErr)
		//log.Println("Using in-memory repositories")
		//
		//// Use in-memory repositories as fallback
		//userRepo = mongodb.NewInMemoryUserRepository()
		//orgRepo = mongodb.NewInMemoryOrganizationRepository()
		//roleRepo = mongodb.NewInMemoryRoleRepository()
		//permRepo = mongodb.NewInMemoryPermissionRepository()
		//tokenRepo = mongodb.NewInMemoryTokenRepository()
	} else {
		log.Println("Connected to MongoDB")
		// Use MongoDB repositories with type assertion
		mongoDbClient := mongoClient.Client
		userRepo = mongodb.NewMongoUserRepository(mongoDbClient, cfg.DatabaseName, "users")
		orgRepo = mongodb.NewMongoOrganizationRepository(mongoDbClient, cfg.DatabaseName, "organizations")
		roleRepo = mongodb.NewMongoRoleRepository(mongoDbClient, cfg.DatabaseName, "roles")
		permRepo = mongodb.NewMongoPermissionRepository(mongoDbClient, cfg.DatabaseName, "permissions")
		tokenRepo = mongodb.NewMongoTokenRepository(mongoDbClient, cfg.DatabaseName, "tokens")
	}
	// Initialize services
	// Token service with repository
	tokenService := auth.NewTokenServiceWithRepository("your-jwt-secret", 24*time.Hour, tokenRepo)

	// Password manager
	passwordManager := auth.NewBcryptPasswordManager(10)

	// User service
	userService := identity.NewUserService(userRepo, passwordManager)

	// Organization service
	orgService := identity.NewOrganizationService(orgRepo, userRepo)

	// Auth service
	authService := auth.NewAuthService(userRepo, tokenService, passwordManager)

	// OAuth service (placeholder)
	oauthService := auth.NewOAuthService(userRepo, tokenService)

	// Role service
	roleService := rbac.NewRoleService(roleRepo, userRepo)

	// Permission service
	permissionService := rbac.NewPermissionService(permRepo, roleRepo)

	// Access control service
	accessService := rbac.NewAccessControlService(roleRepo, permRepo, userRepo)

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userRepo)

	// Initialize API handlers
	// Auth handler
	authHandler := rest.NewAuthHandler(authService, oauthService, passwordManager)

	// User handler
	userHandler := rest.NewUserHandler(userService)

	// Organization handler
	orgHandler := rest.NewOrganizationHandler(orgService)

	// Role handler
	roleHandler := rest.NewRoleHandler(roleService, permissionService)

	// Permission handler
	permHandler := rest.NewPermissionHandler(permissionService)

	// Access control handler
	accessHandler := rest.NewAccessControlHandler(accessService)

	// Register routes
	// Common endpoints
	setupCommonRoutes(app)

	// Auth endpoints
	authHandler.RegisterRoutes(app)

	// User endpoints
	userHandler.RegisterRoutes(app, authMiddleware.Middleware)

	// Organization endpoints
	orgHandler.RegisterRoutes(app, authMiddleware.Middleware)

	// Role endpoints
	roleHandler.RegisterRoutes(app, authMiddleware.Middleware)

	// Permission endpoints
	permHandler.RegisterRoutes(app, authMiddleware.Middleware)

	// Access control endpoints
	accessHandler.RegisterRoutes(app, authMiddleware.Middleware)
	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(fmt.Sprintf("0.0.0.0:%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
