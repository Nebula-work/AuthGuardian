package main

import (
        "fmt"
        "log"
        "os"

        "github.com/gofiber/fiber/v2"
        "github.com/gofiber/fiber/v2/middleware/cors"
        "github.com/gofiber/fiber/v2/middleware/logger"
        "github.com/gofiber/fiber/v2/middleware/recover"

        "rbac-system/config"
        "rbac-system/database"
        "rbac-system/routes"
)

func main() {
        // Load configuration
        cfg, err := config.LoadConfig()
        if err != nil {
                log.Fatalf("Failed to load config: %v", err)
        }

        // Get MongoDB URI from environment or config
        mongoURI := os.Getenv("MONGO_URI")
        if mongoURI == "" {
                mongoURI = cfg.MongoURI
                log.Println("Using MongoDB URI from config file")
        } else {
                log.Println("Using MongoDB URI from environment variable")
        }

        // Connect to MongoDB
        client, err := database.ConnectMongoDB(mongoURI)
        if err != nil {
                log.Fatalf("Failed to connect to MongoDB: %v", err)
        }
        defer client.Disconnect()

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
        app.Use(cors.New(cors.Config{
                AllowOrigins:     "http://localhost:3000",
                AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
                AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
                AllowCredentials: true,
        }))

        // Setup routes
        routes.SetupRoutes(app)

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
