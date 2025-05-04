package middleware

import (
        "context"
        "strings"

        "github.com/gofiber/fiber/v2"
        "go.mongodb.org/mongo-driver/bson"
        "go.mongodb.org/mongo-driver/bson/primitive"

        "rbac-system/config"
        "rbac-system/database"
        "rbac-system/models"
        "rbac-system/utils"
)

// Authenticate middleware validates JWT tokens
func Authenticate(cfg *config.Config, db *database.MongoClient) fiber.Handler {
        return func(c *fiber.Ctx) error {
                // Get the Authorization header
                authHeader := c.Get("Authorization")
                if authHeader == "" {
                        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                                "success": false,
                                "error":   "missing authorization header",
                        })
                }

                // Extract the token
                tokenParts := strings.Split(authHeader, " ")
                if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
                        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                                "success": false,
                                "error":   "invalid authorization format",
                        })
                }

                // Validate the token
                claims, err := utils.ValidateToken(tokenParts[1], cfg.JWTSecret)
                if err != nil {
                        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                                "success": false,
                                "error":   "invalid token: " + err.Error(),
                        })
                }

                // Set the user claims in the context
                c.Locals("userId", claims.UserID)
                c.Locals("username", claims.Username)
                c.Locals("email", claims.Email)
                c.Locals("roleIds", claims.RoleIDs)
                c.Locals("organizationIds", claims.OrganizationIDs)

                return c.Next()
        }
}

// Authorize middleware checks if the user has the required permission
func Authorize(resource, action string, db *database.MongoClient) fiber.Handler {
        return func(c *fiber.Ctx) error {
                // Get user information from context
                _, ok := c.Locals("userId").(primitive.ObjectID)
                if !ok {
                        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                                "success": false,
                                "error":   "user not authenticated",
                        })
                }

                roleIDs, ok := c.Locals("roleIds").([]primitive.ObjectID)
                if !ok {
                        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                                "success": false,
                                "error":   "user roles not found",
                        })
                }

                // Get roles with permission
                var roles []models.Role
                rolesCollection := db.GetCollection(database.RolesCollection)
                
                cursor, err := rolesCollection.Find(context.Background(), bson.M{
                        "_id": bson.M{"$in": roleIDs},
                })
                if err != nil {
                        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                                "success": false,
                                "error":   "failed to retrieve roles",
                        })
                }
                defer cursor.Close(context.Background())

                if err = cursor.All(context.Background(), &roles); err != nil {
                        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                                "success": false,
                                "error":   "failed to decode roles",
                        })
                }

                // Extract all permission IDs from the roles
                var permissionIDs []primitive.ObjectID
                for _, role := range roles {
                        permissionIDs = append(permissionIDs, role.PermissionIDs...)
                }

                // Check if the user has any permissions
                if len(permissionIDs) == 0 {
                        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                                "success": false,
                                "error":   "access denied: no permissions",
                        })
                }

                // Get permissions
                permissionsCollection := db.GetCollection(database.PermissionsCollection)
                
                var permissions []models.Permission
                cursor, err = permissionsCollection.Find(context.Background(), bson.M{
                        "_id": bson.M{"$in": permissionIDs},
                })
                if err != nil {
                        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                                "success": false,
                                "error":   "failed to retrieve permissions",
                        })
                }
                defer cursor.Close(context.Background())

                if err = cursor.All(context.Background(), &permissions); err != nil {
                        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                                "success": false,
                                "error":   "failed to decode permissions",
                        })
                }

                // Check if the user has the required permission
                hasPermission := false
                for _, permission := range permissions {
                        if (permission.Resource == resource || permission.Resource == models.ResourceAll) &&
                                (permission.Action == action || permission.Action == models.ActionAll) {
                                hasPermission = true
                                break
                        }
                }

                if !hasPermission {
                        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                                "success": false,
                                "error":   "access denied: insufficient permissions",
                        })
                }

                return c.Next()
        }
}
