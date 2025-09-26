package handlers

import (
	"golang_task/models"
	"golang_task/repositories"
	"golang_task/utils"
	"log"

	"github.com/gofiber/fiber/v2"
)
// UserRegisterRequest represents the request body for user registration
type UserRegisterRequest struct {
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Username  string `json:"username" example:"johndoe"`
	Email     string `json:"email" example:"john@example.com"`
	Password  string `json:"password" example:"password123"`
}

// UserRegisterResponse represents the response for user registration
type UserRegisterResponse struct {
	Message string `json:"message" example:"user created successfully"`
	Token   string `json:"token" example:"jwt_token_string"`
}

// UserLoginRequest represents the request body for user login
type UserLoginRequest struct {
	Email    string `json:"email,omitempty" example:"john@example.com"`
	Username string `json:"username,omitempty" example:"johndoe"`
	Password string `json:"password" example:"password123"`
}

// UserLoginResponse represents the response for user login
type UserLoginResponse struct {
	Message string `json:"message" example:"user logged in successfully"`
	Token   string `json:"token" example:"jwt_token_string"`
}

// ErrorResponse represents the error response format
type UserErrorResponse struct {
	Error   string `json:"error" example:"failed to create user"`
	Message string `json:"message" example:"detailed error message"`
}

// RegisterHandler godoc
// @Summary Register a new user
// @Description Register a new user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body UserRegisterRequest true "User registration info"
// @Success 201 {object} UserRegisterResponse "User created successfully"
// @Failure 400 {object} ErrorResponse "Failed to create user"
// @Router /register [post]
func RegisterHandler(repo repositories.UserRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UserRegisterRequest
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
			return c.Status(fiber.StatusBadRequest).JSON(UserErrorResponse{
				Error:   "failed to create user",
				Message: err.Error(),
			})
		}

		jwtToken, err := utils.CreateJwt(user.ID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(UserErrorResponse{
				Error:   "failed to create JWT",
				Message: err.Error(),
			})
		}

		log.Println("User Created Successfully username:", user.Username)
		return c.Status(fiber.StatusCreated).JSON(UserRegisterResponse{
			Message: "user created successfully",
			Token:   jwtToken,
		})
	}
}


// LoginHandler godoc
// @Summary Login user
// @Description Authenticate user using email or username and password, returns JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body UserLoginRequest true "User Login Data.  NOTE: Send either username or email for login, but do not provide both at the same time."
// @Success 200 {object} UserLoginResponse "User Logged in successfully"
// @Failure 400 {object} UserErrorResponse "Validation Error."
// @Failure 401 {object} UserErrorResponse "Password is wrong"
// @Failure 404 {object} UserErrorResponse "User not found"
// @Router /auth/login [post]
func LoginHandler(repo repositories.UserRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input UserLoginRequest
		if err := utils.BodyParse(c, &input); err != nil {
			return err
		}

		var getDataErr error
		var user *models.User

		if input.Email != "" && input.Username != "" {
			return c.Status(fiber.StatusBadRequest).JSON(UserErrorResponse{
				Error:   "invalid input",
				Message: "send either email or username, not both",
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
			return c.Status(fiber.StatusBadRequest).JSON(UserErrorResponse{
				Error:   "invalid input",
				Message: "email or username must be provided",
			})
		}

		if getDataErr != nil {
			return c.Status(fiber.StatusNotFound).JSON(UserErrorResponse{
				Error:   "user not found",
				Message: getDataErr.Error(),
			})
		}

		if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
			log.Printf("Password is wrong for user %s:%s", user.Username, err)
			return c.Status(fiber.StatusUnauthorized).JSON(UserErrorResponse{
				Error:   "unauthorized",
				Message: "password is wrong",
			})
		}

		jwtToken, err := utils.CreateJwt(user.ID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(UserErrorResponse{
				Error:   "failed to create JWT",
				Message: err.Error(),
			})
		}

		log.Println("User Logged In Successfully username:", user.Username)
		return c.Status(fiber.StatusOK).JSON(UserLoginResponse{
			Message: "user logged in successfully",
			Token:   jwtToken,
		})
	}
}
