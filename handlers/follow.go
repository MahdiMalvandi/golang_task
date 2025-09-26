package handlers

import (
	"log"
	"strconv"

	"golang_task/models"
	"golang_task/repositories"

	"github.com/gofiber/fiber/v2"
)

// Response structs for Swagger

type FollowResponse struct {
	Message string `json:"message" example:"followed successfully"`
}

type UnfollowResponse struct {
	Message string `json:"message" example:"unfollowed successfully"`
}

type FollowErrorResponse struct {
	Error   string `json:"error" example:"error"`
	Message string `json:"message" example:"error message"`
}

// GetFollowers godoc
// @Summary Get followers
// @Description Get all followers of the authenticated user
// @Tags Follow
// @Security ApiKeyAuth
// @Success 200 {array} models.User
// @Failure 400 {object} FollowErrorResponse
// @Router /follows/followers [get]
func GetFollowers(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)
		log.Printf("[INFO] User %d is fetching followers", userID)
		followers, err := repo.GetFollowers(userID)
		if err != nil {
			log.Printf("[ERROR] Failed to get followers for user %d: %v", userID, err)
			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "failed to get followers",
				Message: err.Error(),
			})
		}
		log.Printf("[INFO] User %d followers fetched successfully, count=%d", userID, len(followers))
		return c.Status(fiber.StatusOK).JSON(followers)
	}
}

// GetFollowing godoc
// @Summary Get following
// @Description Get all users the authenticated user is following
// @Tags Follow
// @Security ApiKeyAuth
// @Success 200 {array} models.User
// @Failure 400 {object} FollowErrorResponse
// @Router /follows/following [get]
func GetFollowing(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)
		log.Printf("[INFO] User %d is fetching followings", userID)

		following, err := repo.GetFollowings(userID)
		if err != nil {
			log.Printf("[ERROR] [ERROR][ERROR] Failed to get followings for user %d: %v", userID, err)
			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "failed to get following",
				Message: err.Error(),
			})
		}
		if len(following) == 0 {
			following = []models.User{}
		}
		log.Printf("[INFO] User %d followings fetched successfully, count=%d", userID, len(following))

		return c.Status(fiber.StatusOK).JSON(following)
	}
}

// Follow godoc
// @Summary Follow a user
// @Description Follow another user. User cannot follow themselves and following must exist.
// @Tags Follow
// @Accept json
// @Produce json
// @Param following_id path int true "ID of the user to follow"
// @Success 200 {object} FollowResponse "followed successfully"
// @Failure 400 {object} FollowErrorResponse "validation error or user not found or trying to follow yourself"
// @Failure 404 {object} FollowErrorResponse "follower not found"
// @Router /follows/{following_id} [post]
func Follow(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)

		followingId64, err := strconv.ParseUint(c.Params("following_id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "invalid following id",
				Message: err.Error(),
			})
		}
		followingId := uint(followingId64)
		log.Printf("[INFO] User %d is trying to follow user %d", userID, followingId)
		if userID == followingId {
			log.Printf("[ERROR] User %d tried to follow themselves", userID)
			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "cannot follow yourself",
				Message: "you cannot follow yourself",
			})
		}

		if err := repo.Follow(userID, followingId); err != nil {
			log.Printf("[ERROR] User %d failed to follow user %d: %v", userID, followingId, err)

			var statusCode int
			if err.Error() == "following user not found" {
				statusCode = fiber.StatusNotFound
			} else {
				statusCode = fiber.StatusBadRequest
			}
			return c.Status(statusCode).JSON(FollowErrorResponse{
				Error:   "failed to follow",
				Message: err.Error(),
			})
		}
		log.Printf("[INFO] User %d followed user %d successfully", userID, followingId)
		return c.Status(fiber.StatusOK).JSON(FollowResponse{
			Message: "followed successfully",
		})
	}
}

// UnFollow godoc
// @Summary Follow a user
// @Description UnFollow another user. User cannot Unfollow themselves and following must exist.
// @Tags Follow
// @Accept json
// @Produce json
// @Param following_id path int true "ID of the user to unfollow"
// @Success 200 {object} FollowResponse "unfollowed successfully"
// @Failure 400 {object} FollowErrorResponse "validation error or user not found or trying to unfollow yourself"
// @Failure 404 {object} FollowErrorResponse "following not found"
// @Router /follows/{following_id} [delete]
func Unfollow(repo repositories.FollowRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)

		followingId64, err := strconv.ParseUint(c.Params("following_id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "invalid following id",
				Message: err.Error(),
			})
		}
		followingId := uint(followingId64)
		log.Printf("[INFO] User %d is trying to unfollow user %d", userID, followingId)

		if userID == followingId {
			log.Printf("[ERROR] User %d tried to unfollow themselves", userID)
			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "cannot unfollow yourself",
				Message: "you cannot unfollow yourself",
			})
		}

		if err := repo.UnFollow(userID, followingId); err != nil {
			log.Printf("[ERROR] User %d failed to unfollow user %d: %v", userID, followingId, err)

			return c.Status(fiber.StatusBadRequest).JSON(FollowErrorResponse{
				Error:   "failed to unfollow",
				Message: err.Error(),
			})
		}

		log.Printf("[INFO] User %d unfollowed user %d successfully", userID, followingId)
		return c.Status(fiber.StatusOK).JSON(UnfollowResponse{
			Message: "unfollowed successfully",
		})
	}
}
