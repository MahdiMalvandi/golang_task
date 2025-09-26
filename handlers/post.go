package handlers

import (
	"fmt"
	"golang_task/models"
	"golang_task/repositories"
	"golang_task/utils"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func PostCreate(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// create input struct
		var input struct {
			Title     string `json:"title"`
			Content   string `json:"content"`
			MediaPath string `json:"media_path"`
			AuthorID  uint   `json:"author_id"`
		}
		// get user id from context
		userId := c.Locals("user_id").(uint)
		input.AuthorID = userId

		// get media file
		file, err := c.FormFile("media")
		if err != nil {
			if strings.Contains(err.Error(), "there is no uploaded file") { // User does not send file
				input.MediaPath = ""
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

		} else { // user send a file
			// Check file size
			maxFileSize, err := strconv.ParseUint(os.Getenv("MAX_FILE_SIZE"), 10, 64)
			if err != nil {
				fmt.Println(err.Error())
				// Default file size
				maxFileSize = 50
			}
			if file.Size > int64(maxFileSize*1024*1024) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "file size exceeds 50MB",
				})
			}

			// Check Prefixes
			allowSuffix := []string{"png", "jpg", "jpeg", "gif", "bmp", "webp", "mp4", "mov", "avi", "mkv", "flv", "wmv", "webm"}

			fileNameSplited := strings.Split(file.Filename, ".")
			fileSuffix := fileNameSplited[len(fileNameSplited)-1]
			if !slices.Contains(allowSuffix, fileSuffix) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid file type",
				})
			}

			// Uploading file
			filename := strings.ReplaceAll(fmt.Sprintf("%d_%d_%s", time.Now().Unix(), userId, file.Filename), " ", "-")

			path := fmt.Sprintf("./uploads/%s", filename)

			if err := c.SaveFile(file, path); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to save file", "message": err.Error()})
			}
			input.MediaPath = path
		}

		if err := utils.BodyParse(c, &input); err != nil {
			return err
		}

		// Create Post model in db
		var post = models.Post{
			Title:     input.Title,
			Content:   input.Content,
			MediaPath: input.MediaPath,
			AuthorID:  input.AuthorID,
		}
		if err := repo.Create(&post); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "failed to create post",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"message": "post created successfully",
		})
	}
}
func PostTimeline(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Locals("user_id").(uint)
		limit, _ := strconv.ParseUint(c.Params("limit"), 10, 64)
		page, _ := strconv.ParseUint(c.Params("page"), 10, 64)

		start := (page - 1) * limit
		end := start + (limit) -1
		posts, err := repo.GetTimeline(userId, int64(start), int64(end))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusAccepted).JSON(map[string]interface{}{
			"posts": posts,
		})
	}
}
func PostGetByID(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postIdParams, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid post id",
			})
		}

		post, err := repo.GetByID(uint(postIdParams))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "post not found",
			})
		}

		return c.JSON(post)
	}
}

func DeletePost(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postIdParams, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid post id",
			})
		}

		err = repo.DeleteByID(uint(postIdParams))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "post not found",
			})
		}

		return c.JSON(fiber.Map{
			"message": "post deleted successfully",
		})
	}
}

func PostEdit(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// create input struct
		var input struct {
			Title     string `json:"title,omitempty"`
			Content   string `json:"content,omitempty"`
			MediaPath string `json:"media_path,omitempty"`
		}

		// Get Post id
		postIdParams, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid post id",
			})
		}

		// get post from db
		post, err := repo.GetByID(uint(postIdParams))
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "post not found",
			})
		}

		// get user id from context
		userId := c.Locals("user_id").(uint)

		if post.AuthorID != userId {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "you are not the author of this post",
			})
		}

		// get media file
		file, err := c.FormFile("media")
		if err == nil { // user send a file
			oldFilePath := post.MediaPath

			// Check file size
			maxFileSize, err := strconv.ParseUint(os.Getenv("MAX_FILE_SIZE"), 10, 64)
			if err != nil {
				fmt.Println(err.Error())
				// Default file size
				maxFileSize = 50
			}
			if file.Size > int64(maxFileSize*1024*1024) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "file size exceeds 50MB",
				})
			}

			// Check Prefixes
			allowSuffix := []string{"png", "jpg", "jpeg", "gif", "bmp", "webp", "mp4", "mov", "avi", "mkv", "flv", "wmv", "webm"}

			fileNameSplited := strings.Split(file.Filename, ".")
			fileSuffix := fileNameSplited[len(fileNameSplited)-1]
			if !slices.Contains(allowSuffix, fileSuffix) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "invalid file type",
				})
			}

			// Uploading file
			filename := strings.ReplaceAll(fmt.Sprintf("%d_%d_%s", time.Now().Unix(), userId, file.Filename), " ", "-")

			path := fmt.Sprintf("./uploads/%s", filename)

			if err := c.SaveFile(file, path); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to save file", "message": err.Error()})
			}
			input.MediaPath = path

			if oldFilePath != "" {
				if err := os.Remove(oldFilePath); err != nil {
					// remove new file
					os.Remove(path)
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to remove old file", "message": err.Error()})
				}
			}

		}

		if err := utils.BodyParse(c, &input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		if err := repo.Update(post.ID, input); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "failed to update post",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"message": "post created successfully",
		})
	}
}
