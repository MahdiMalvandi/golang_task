package main

import (
	"golang_task/models"
	"golang_task/routers"
	"golang_task/workers"
	"log"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	_ "golang_task/docs"
	"gorm.io/gorm"
)

// @title Social Media API
// @version 1.0
// @description This API allows authenticated users to create, edit, delete posts and view their timeline.\n All endpoints require login and a valid JWT token provided in the `Authorization` header as `Bearer <token>`.\n The timeline endpoint supports pagination and retrieves posts from Redis cache for fast access. \n Fan-out worker ensures that newly created posts are propagated to followers' timelines automatically.\n
func main() {
	// Database Connection
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err != nil {
		log.Println("[ERROR] Failed to connect to SQLite:", err)
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("[ERROR] No .env file found, using system environment variables")
	}

	// Fiber App
	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024,
	})

	// BackGround Workers
	go workers.FanOutWorker(rdb, db)
	
	// Routers
	app.Static("/uploads", "./uploads")
	app.Get("/swagger/*", swagger.HandlerDefault)

	routers.UserRoutes(app, db)
	routers.PostRoute(app, db, rdb)
	routers.FollowRoute(app, db, rdb)


	db.AutoMigrate(&models.User{}, &models.Follow{}, &models.Post{})
	
	log.Println(app.Listen(":3000"))
}
