package routers

import (
	"golang_task/handlers"
	"golang_task/middlewares"
	"golang_task/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func PostRoute(app *fiber.App, db *gorm.DB, rdb *redis.Client) {
	posts := app.Group("/posts")

	repo := repositories.NewPostRepository(db, rdb)

	posts.Use(middlewares.AuthRequired())
	posts.Post("/", handlers.PostCreate(repo))
	posts.Get("/timeline/:limit/:page", handlers.PostTimeline(repo))
	posts.Get("/:id", handlers.PostGetByID(repo))
	posts.Delete("/:id", handlers.DeletePost(repo))
	posts.Put("/:id", handlers.PostEdit(repo))
	
}