package utils

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func BodyParse(c *fiber.Ctx, input interface{}) error {
	if err := c.BodyParser(input); err != nil {
		log.Println("failed to parse request body:", err)
		return err
	}
	return nil
}