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

type PostCreateInput struct {
	Title     string `json:"title" example:"My first post"`
	Content   string `json:"content" example:"Hello world!"`
	MediaPath string `json:"media_path,omitempty" example:"/uploads/abc.png"`
	AuthorID  uint   `json:"author_id" example:"1"`
}

// PostSuccessfullResponse represents successful creation response
type PostSuccessfullResponse struct {
	Message string `json:"message" example:"operation was successfully"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error" example:"failed to create post"`
	Message string `json:"message,omitempty" example:"error message"`
}

// PostCreate godoc
// @Summary Create a new post
// @Description Create a new post with optional media file .User must be logged in and provide a valid token.
// @Tags Posts
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Post title"
// @Param content formData string true "Post content"
// @Param media formData file false "Media file (image/video)"
// @Success 201 {object} PostSuccessfullResponse "Post created successfully"
// @Failure 400 {object} ErrorResponse "Bad request or validation error"
// @Security ApiKeyAuth
// @Router /posts [post]
func PostCreate(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// create input struct
		var input PostCreateInput
		// get user id from context
		userID := c.Locals("user_id").(uint)
		input.AuthorID = userID

		// get media file
		file, err := c.FormFile("media")
		if err != nil {
			if strings.Contains(err.Error(), "there is no uploaded file") { // User does not send file
				input.MediaPath = ""
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
					Error:   "failed to create post",
					Message: err.Error(),
				})
			}

		} else { // user send a file
			// Check file size
			maxFileSize, err := strconv.ParseUint(os.Getenv("MAX_FILE_SIZE"), 10, 64)
			if err != nil {
				// Default file size
				maxFileSize = 50
			}
			if file.Size > int64(maxFileSize*1024*1024) {

				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
					Error:   "failed to create post",
					Message: "file size exceeds 50MB",
				})
			}

			// Check Prefixes
			allowSuffix := []string{"png", "jpg", "jpeg", "gif", "bmp", "webp", "mp4", "mov", "avi", "mkv", "flv", "wmv", "webm"}

			fileNameSplited := strings.Split(file.Filename, ".")
			fileSuffix := fileNameSplited[len(fileNameSplited)-1]
			if !slices.Contains(allowSuffix, fileSuffix) {

				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
					Error:   "failed to create post",
					Message: "invalid file type",
				})
			}

			// Uploading file
			filename := strings.ReplaceAll(fmt.Sprintf("%d_%d_%s", time.Now().Unix(), userID, file.Filename), " ", "-")

			path := fmt.Sprintf("./uploads/%s", filename)

			if err := c.SaveFile(file, path); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
					Error:   "failed to create post",
					Message: err.Error(),
				})
			}
			input.MediaPath = path
		}

		if err := utils.BodyParse(c, &input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to create post",
				Message: err.Error(),
			})
		}

		// Create Post model in db
		var post = models.Post{
			Title:     input.Title,
			Content:   input.Content,
			MediaPath: input.MediaPath,
			AuthorID:  input.AuthorID,
		}
		if err := repo.Create(&post); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to create post",
				Message: err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(PostSuccessfullResponse{
			Message: "post created successfully",
		})
	}
}

// PostTimeline godoc
// @Summary Get user's timeline posts
// @Description Get posts from user's followings with pagination
// @Tags Posts
// @Produce json
// @Param limit path int true "Number of posts per page"
// @Param page path int true "Page number"
// @Success 202 {object} []models.Post
// @Failure 400 {object} ErrorResponse "Bad request"
// @Security ApiKeyAuth
// @Router /timeline/{limit}/{page} [get]
func PostTimeline(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(uint)
		limit, _ := strconv.ParseUint(c.Params("limit"), 10, 64)
		page, _ := strconv.ParseUint(c.Params("page"), 10, 64)

		start := (page - 1) * limit
		end := start + (limit) - 1
		posts, err := repo.GetTimeline(userID, int64(start), int64(end))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to see timeline",
				Message: err.Error(),
			})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"posts": posts,
		})
	}
}

// PostGetByID godoc
// @Summary Get post by ID
// @Description Get a single post by its ID
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.Post
// @Failure 400 {object} ErrorResponse "Invalid post id"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Security ApiKeyAuth
// @Router /posts/{id} [get]
func PostGetByID(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postIdParams, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {

			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to get post",
				Message: "invalid post id",
			})
		}

		post, err := repo.GetByID(uint(postIdParams))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to get post",
				Message: "post not found",
			})
		}

		return c.JSON(post)
	}
}

// DeletePost godoc
// @Summary Delete a post
// @Description Delete a post by ID (only author can delete)
// @Tags Posts
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} PostSuccessfullResponse "Post deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid post id"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Failure 403 {object} ErrorResponse "Forbidden: not the author"
// @Security ApiKeyAuth
// @Router /posts/{id} [delete]
func DeletePost(repo repositories.PostRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Get Post id
		postIdParams, err := strconv.ParseUint(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to delete post",
				Message: "invalid post id",
			})
		}

		// get post from db
		post, err := repo.GetByID(uint(postIdParams))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to delete post",
				Message: err.Error(),
			})
		}

		// get user id from context
		userID := c.Locals("user_id").(uint)


		err = repo.DeletePost(post, uint(userID))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to delete post",
				Message: err.Error(),
			})
		}

		return c.JSON(PostSuccessfullResponse{
			Message: "post deleted successfully",
		})
	}
}

// PostEdit godoc
// @Summary Edit a post
// @Description Edit a post's title, content, or media (only author can edit)
// @Tags Posts
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Post ID"
// @Param title formData string false "Post title"
// @Param content formData string false "Post content"
// @Param media formData file false "Media file"
// @Success 201 {object} PostSuccessfullResponse "Post updated successfully"
// @Failure 400 {object} ErrorResponse "Bad request or validation error"
// @Failure 403 {object} ErrorResponse "Forbidden: not the author"
// @Failure 404 {object} ErrorResponse "Post not found"
// @Security ApiKeyAuth
// @Router /posts/{id} [put]
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
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to update post",
				Message: err.Error(),
			})
		}

		// get user id from context
		userID := c.Locals("user_id").(uint)

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
				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
					Error:   "failed to update post",
					Message: "file size exceeds 50MB",
				})
			}

			// Check Prefixes
			allowSuffix := []string{"png", "jpg", "jpeg", "gif", "bmp", "webp", "mp4", "mov", "avi", "mkv", "flv", "wmv", "webm"}

			fileNameSplited := strings.Split(file.Filename, ".")
			fileSuffix := fileNameSplited[len(fileNameSplited)-1]
			if !slices.Contains(allowSuffix, fileSuffix) {
				return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
					Error:   "failed to update post",
					Message: "invalid file type",
				})
			}

			// Uploading file
			filename := strings.ReplaceAll(fmt.Sprintf("%d_%d_%s", time.Now().Unix(), userID, file.Filename), " ", "-")

			path := fmt.Sprintf("./uploads/%s", filename)

			if err := c.SaveFile(file, path); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "failed to save file", "message": err.Error()})
			}
			input.MediaPath = path

			if oldFilePath != "" {
				if err := os.Remove(oldFilePath); err != nil {
					// remove new file
					os.Remove(path)
					return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
						Error:   "failed to update post",
						Message: err.Error(),
					})
				}
			}

		}

		if err := utils.BodyParse(c, &input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to update post",
				Message: err.Error(),
			})
		}

		if err := repo.UpdatePost(post, userID, input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error:   "failed to update post",
				Message: err.Error(),
			})

		}

		return c.Status(fiber.StatusCreated).JSON(PostSuccessfullResponse{
			Message: "post updated successfully",
		})
	}
}
