package routers

import (
	"golang_task/handlers"
	"golang_task/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UserRoutes(app *fiber.App, db *gorm.DB) {
	users := app.Group("/users")
	
	repo := repositories.NewUserRepository(db)

	users.Post("/signup", handlers.RegisterHandler(repo))
	users.Post("/login", handlers.LoginHandler(repo))

}