package handlers

import (
	"golang_task/models"
	"golang_task/repositories"
	"golang_task/utils"
	"log"

	"github.com/gofiber/fiber/v2"
)

func RegisterHandler(repo repositories.UserRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			Password  string `json:"password"`
		}

		if err := utils.BodyParse(c, &input); err != nil {
			return err
		}

		user := models.User{
			Firstname: input.FirstName,
			Lastname:  input.LastName,
			Username:  input.Username,
			Password:  input.Password,
			Email:     input.Email,
		}

		if err := repo.Create(&user); err != nil {
			log.Println("failed to create user:", err)

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "failed to create user",
				"message": err.Error(),
			})
		}

		jwtToken, err := utils.CreateJwt(user.ID)
		if err != nil {

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "failed to create JWT",
				"message": err.Error(),
			})
		}
		log.Println("User Created Successfully username:", user.Username)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "user created successfully",
			"token":   jwtToken,
		})
	}
}

func LoginHandler(repo repositories.UserRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := utils.BodyParse(c, &input); err != nil {
			return err
		}

		var getDataErr error
		var user *models.User
		if input.Email != "" && input.Username != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "send either email or username, not both",
			})
		} else if input.Email != "" {
			user, getDataErr = repo.GetByEmail(input.Email)
			if getDataErr != nil {
				log.Printf("failed to get user by email %s: %v", input.Email, getDataErr)
			}
		} else if input.Username != "" {
			user, getDataErr = repo.GetByUsername(input.Username)
			if getDataErr != nil {
				log.Printf("failed to get user by username %s: %v", input.Username, getDataErr)
			}
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "email or username must be provided",
			})
		}

		if getDataErr != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "user not found",
				"message": getDataErr.Error(),
			})
		}

		if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
			log.Printf("Password is wrong for user %s:%s", user.Username, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "password is wrong",
			})
		}

		jwtToken, err := utils.CreateJwt(user.ID)
		if err != nil {
			c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "failed to create JWT",
				"message": err.Error(),
			})
			return err
		}
		log.Println("User Logged In Successfully username:", user.Username)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "user logged in successfully",
			"token":   jwtToken,
		})
	}
}
