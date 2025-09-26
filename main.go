package main

import (
	"golang_task/models"
	"golang_task/routers"
	"golang_task/workers"
	"log"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err != nil {
		log.Println("Failed to connect to SQLite:", err)
	}
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024,
	})

	go workers.StartWorker(rdb, db)
	app.Static("/uploads", "./uploads")

	routers.UserRoutes(app, db)
	routers.PostRoute(app, db, rdb)
	routers.FollowRoute(app, db)
	db.AutoMigrate(&models.User{}, &models.Follow{}, &models.Post{})
	log.Println(app.Listen(":3000"))
}
