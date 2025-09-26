package main

import (
	"fmt"
	"golang_task/models"
	"golang_task/routers"
	"golang_task/workers"
	"log"
	"os"

	_ "golang_task/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// @title Social Media API
// @version 1.0
// @description This API allows authenticated users to create, edit, delete posts and view their timeline.\n All endpoints require login and a valid JWT token provided in the `Authorization` header as `Bearer <token>`.\n The timeline endpoint supports pagination and retrieves posts from Redis cache for fast access. \n Fan-out worker ensures that newly created posts are propagated to followers' timelines automatically.\n
func main() {
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		err := os.Mkdir("./uploads", os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create uploads directory: %v", err)
		}
		fmt.Println("Uploads directory created")
	}
	// Database Connection
dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_HOST"),
    os.Getenv("DB_NAME"),
)
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})


	// Redis Connection
rdb := redis.NewClient(&redis.Options{
    Addr: os.Getenv("REDIS_ADDR"),
})


	if err != nil {
		log.Println("[ERROR] Failed to connect to SQLite:", err)
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
