package middlewares

import (
	"golang_task/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)


// This function checks user's jwt token
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header format",
			})
		}

		// Checking jwt token
		userID, check, err := utils.VerifyJwt(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}
		if !check {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token is invalid or has expired",
			})
		}

		// Add to locals
		c.Locals("user_id", userID)
		return c.Next()
	}
}
