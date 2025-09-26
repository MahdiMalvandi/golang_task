package handlers

import (
	"golang_task/models"
	"golang_task/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetFollowers(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("user_id").(uint)
		followers, err := repo.GetFollowers(userId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(followers)
	}
}
func GetFollowing(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("user_id").(uint)
		following, err := repo.GetFollowings(userId)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error": err.Error(),
			})
		}
		if len(following) == 0 {
			following = []models.User{}
		}
		return c.Status(fiber.StatusOK).JSON(following)
	}
}
func Follow(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user id
		userId := c.Locals("user_id").(uint)

		// Get following id
		var followingId uint
		followingId64, err := strconv.ParseUint(c.Params("following_id"), 10, 64)
		followingId = uint(followingId64)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error":"invalid follwing id",
			})
		}

		if userId == followingId {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error":"you cannot follow yourself",
			})
		}

		if err := repo.Follow(userId, followingId); err != nil{
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error":err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(map[string]string{
			"message":"followed successfully",
		})
	}
}
func Unfollow(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user id
		userId := c.Locals("user_id").(uint)

		// Get following id
		var followingId uint
		followingId64, err := strconv.ParseUint(c.Params("following_id"), 10, 64)
		followingId = uint(followingId64)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error":"invalid follwing id",
			})
		}

		if userId == followingId {
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error":"you cannot unfollow yourself",
			})
		}
		if err := repo.UnFollow(userId, followingId); err != nil{
			return c.Status(fiber.StatusBadRequest).JSON(map[string]string{
				"error":err.Error(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(map[string]string{
			"message":"unfollowed successfully",
		})
	}
}
