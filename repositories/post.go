package repositories

import (
	"context"
	"errors"
	"fmt"
	"golang_task/models"
	"golang_task/utils"
	"log"
	"sort"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Post Repository interface
type PostRepositoryInterface interface {
	Create(post *models.Post) error
	GetByID(id uint) (*models.Post, error)
	GetPostsByIDs(postIds []uint) ([]models.Post, error)
	GetByAuthorID(authorID uint) ([]models.Post, error)
	GetByAuthorUsername(username string) ([]models.Post, error)
	UpdatePost(post *models.Post, userID uint, updates interface{}) error
	DeletePost(post *models.Post, userID uint) error
	GetTimeline(userID uint, start, end int64) ([]models.Post, error)
	GetFollowingsPosts(userID uint, start, end int64) ([]models.Post, error)
}

// Post repository struct
type postRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

// Post repository constructor
func NewPostRepository(db *gorm.DB, rdb *redis.Client) PostRepositoryInterface {
	return &postRepository{
		db:  db,
		rdb: rdb,
	}
}

// Post repository methods

// This method creates a new post
//
// If the error is nil, the post was created successfully.
func (r *postRepository) Create(post *models.Post) error {
	var err error
	if err = r.db.Create(post).Error; err != nil {
		log.Printf("[ERROR] Failed to create post for author %d: %v", post.AuthorID, err)

		return err
	}

	utils.PostQueue(post, r.rdb, true)
	log.Printf("[INFO] Post added to queue successfully: ID=%d, AuthorID=%d", post.ID, post.AuthorID)


	log.Printf("[INFO] Post created successfully: ID=%d, AuthorID=%d", post.ID, post.AuthorID)

	return nil
}

// This method retrieves a post by ID
//
// If the post is found, it returns the post. If not, it returns an error.
func (r *postRepository) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.Preload("Author").First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] Post with id %d not found", id)
			return nil, fmt.Errorf("post not found")
		}
		log.Printf("[ERROR] Error fetching post id %d: %v", id, err)

		return nil, err
	}
	log.Printf("[INFO] Post fetched successfully: ID=%d", id)

	return &post, nil
}

// This method retrieves posts by their IDs
//
// If the error is nil, the posts were retrieved successfully.
func (r *postRepository) GetPostsByIDs(postIds []uint) ([]models.Post, error) {
	if len(postIds) == 0 {
		return []models.Post{}, nil
	}
	var posts []models.Post
	if err := r.db.Preload("Author").Find(&posts, postIds).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] Posts with ids %v not found", postIds)
			return nil, fmt.Errorf("posts not found")
		}
		return nil, err
	}
	return posts, nil
}

// This method retrieves posts by their author ID
//
// If the error is nil, the posts were retrieved successfully.
func (r *postRepository) GetByAuthorID(authorID uint) ([]models.Post, error) {
	var posts []models.Post
	if err := r.db.Where("author_id = ?", authorID).Find(&posts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] Posts with author_id %d not found", authorID)
			return nil, fmt.Errorf("posts not found")
		}
		return nil, err
	}
	return posts, nil
}

// This method gets posts by their author username
//
// If the error is nil, the posts were retrieved successfully.
func (r *postRepository) GetByAuthorUsername(username string,) ([]models.Post, error) {
	var posts []models.Post
	if err := r.db.Where("author_username = ?", username).Find(&posts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] Posts with author_username %s not found", username)
			return nil, fmt.Errorf("posts not found")
		}
		return nil, err
	}
	return posts, nil
}

// This method updates a post
//
// If the error is nil, the post was updated successfully.
func (r *postRepository) UpdatePost(post *models.Post, userID uint, updates interface{}) error {

	if post.AuthorID != userID {
		log.Printf("[ERROR] User %d is not the author of post %d", userID, post.ID)

		return fmt.Errorf("you are not the author of this post")
	}

	if err := r.db.Model(post).Updates(updates).Error; err != nil {
		log.Printf("[ERROR] Failed to update post %d by user %d: %v", post.ID, userID, err)

		return err
	}
	log.Printf("[INFO] Post %d updated successfully by user %d", post.ID, userID)

	return nil
}

// This method deletes a post
//
// If the error is nil, the post was deleted successfully.
func (r *postRepository) DeletePost(post *models.Post, userID uint) error {

	if post.AuthorID != userID {
		log.Printf("[ERROR] User %d tried to delete post %d but is not the author", userID, post.ID)

		return fmt.Errorf("you are not the author of this post")
	}

	if err := r.db.Delete(post).Error; err != nil {
		log.Printf("[ERROR] User %d tried to delete post %d error %v", userID, post.ID, err)

		return err
	}
	utils.PostQueue(post, r.rdb, false)
	log.Printf("[INFO] Post added to queue for delete successfully: ID=%d, AuthorID=%d", post.ID, post.AuthorID)

	log.Printf("[INFO] Post %d deleted successfully by user %d", post.ID, userID)

	return nil
}

func (r *postRepository) GetFollowingsPosts(userID uint, start, end int64) (posts []models.Post, err error) {
	log.Printf("[INFO] Fetching posts from followings of user %d", userID)

	followRepo := NewFollowRepository(r.db, r.rdb)
	followings, err := followRepo.GetFollowings(userID)
	if err != nil {
		return posts, err
	}
	if len(followings) == 0 {
		return []models.Post{}, nil
	}
	followingIDs := make([]uint, 0, len(followings))

	for _, f := range followings {
		followingIDs = append(followingIDs, f.ID)
	}

	limit := int(end - start + 1)
	page := int((int(start) / limit) + 1)
	if err := r.db.Preload("Author").
		Where("author_id IN ?", followingIDs).
		Order("created_at DESC").
		Offset((page - 1) * limit).Limit(limit).Find(&posts).Error; err != nil {
		return posts, err
	}
	log.Printf("[INFO] Fetched %d posts from followings of user %d", len(posts), userID)

	return posts, nil
}

func (r *postRepository) GetTimeline(userID uint, start, end int64) ([]models.Post, error) {
	var posts []models.Post
	log.Printf("[INFO] Fetching timeline for user %d, start=%d, end=%d", userID, start, end)

	ctx := context.Background()
	key := fmt.Sprintf("timeline:%d", userID)
	result, err := r.rdb.ZRevRange(ctx, key, start, end).Result()
	if err != nil {
		log.Printf("[ERROR] Error for fetching timeline for user %d error %v", userID, err)

		return posts, err
	}

	if len(result) == 0 {
		log.Printf("[INFO] Timeline for user %d fetched from DB because redis was empty", userID)

		posts, err = r.GetFollowingsPosts(userID, start, end)
		if err != nil {
			return posts, err
		}
		ctx := context.Background()
		for _, post := range posts {
			r.rdb.ZAdd(ctx, key, redis.Z{
				Score:  float64(post.CreatedAt.Unix()),
				Member: post.ID,
			})
		}
		return posts, nil

	}
	postIds := []uint{}
	for _, s := range result {
		num, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			fmt.Println("Conversion error:", err)
			continue
		}
		postIds = append(postIds, uint(num))
	}
	log.Printf("[INFO] Timeline for user %d fetched from Redis", userID)
	a, _ := r.rdb.ZRevRange(ctx, key, 0, 100).Result()
	fmt.Println(a)
	posts, err = r.GetPostsByIDs(postIds)
	if err != nil {
		return posts, err
	}
	idOrder := map[uint]int{}
	for i, id := range postIds {
		idOrder[id] = i
	}
	sort.Slice(posts, func(i, j int) bool {
		return idOrder[posts[i].ID] < idOrder[posts[j].ID]
	})
	log.Printf("[INFO] Timeline was sent for user %d", userID)

	return posts, nil
}
