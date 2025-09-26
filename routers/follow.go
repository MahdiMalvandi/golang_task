package routers

import (
	"golang_task/handlers"
	"golang_task/middlewares"
	"golang_task/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func FollowRoute(app *fiber.App, db *gorm.DB, rdb *redis.Client) {
	follows := app.Group("/follows")

	repo := repositories.NewFollowRepository(db, rdb)
	
	follows.Use(middlewares.AuthRequired())
	follows.Get("/followers", handlers.GetFollowers(repo))
	follows.Get("/followings", handlers.GetFollowing(repo))
	follows.Post("/:following_id", handlers.Follow(repo))
	follows.Delete("/:following_id", handlers.Unfollow(repo))
	
}