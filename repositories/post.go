package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang_task/models"
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
	GetByAuthorID(authorID uint, limit, offset int) ([]models.Post, error)
	GetByAuthorUsername(username string, limit, offset int) ([]models.Post, error)
	Update(id uint, updates interface{}) error
	DeleteByID(id uint) error
	GetTimeline(userID uint, start, end int64) ([]models.Post, error)
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
		return err
	}
	queueKey := "post_queue"
	data, _ := json.Marshal(map[string]uint{
		"post_id":   post.ID,
		"author_id": post.AuthorID,
	})

	r.rdb.RPush(context.Background(), queueKey, data)

	return nil
}

// This method retrieves a post by ID
//
// If the post is found, it returns the post. If not, it returns an error.
func (r *postRepository) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	if err := r.db.Preload("Author").First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Post with id %d not found", id)
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}
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
			log.Printf("Posts with ids %v not found", postIds)
			return nil, fmt.Errorf("posts not found")
		}
		return nil, err
	}
	return posts, nil
}

// This method retrieves posts by their author ID
//
// If the error is nil, the posts were retrieved successfully.
func (r *postRepository) GetByAuthorID(authorID uint, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	if err := r.db.Where("author_id = ?", authorID).Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Posts with author_id %d not found", authorID)
			return nil, fmt.Errorf("posts not found")
		}
		return nil, err
	}
	return posts, nil
}

// This method gets posts by their author username
//
// If the error is nil, the posts were retrieved successfully.
func (r *postRepository) GetByAuthorUsername(username string, limit, offset int) ([]models.Post, error) {
	var posts []models.Post
	if err := r.db.Where("author_username = ?", username).Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Posts with author_username %s not found", username)
			return nil, fmt.Errorf("posts not found")
		}
		return nil, err
	}
	return posts, nil
}

// This method updates a post
//
// If the error is nil, the post was updated successfully.
func (r *postRepository) Update(id uint, updates interface{}) error {
	var err error

	if err = r.db.Model(&models.Post{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Post with id %d not found or you are not the author", id)
			return err
		}
	}
	return err
}

// This method deletes a post
//
// If the error is nil, the post was deleted successfully.
func (r *postRepository) DeleteByID(id uint) error {
	var err error
	if err = r.db.Where("id = ?", id).Delete(&models.Post{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Post with id %d not found", id)
			return fmt.Errorf("post not found")
		}
	}
	return err
}

func (r *postRepository) GetTimeline(userID uint, start, end int64) (posts []models.Post, err error) {
	ctx := context.Background()
	result, err := r.rdb.ZRevRange(ctx, fmt.Sprintf("timeline:%d", userID), start, end).Result()
	if err != nil {
		return posts, fmt.Errorf("error")
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
	posts, err = r.GetPostsByIDs(postIds)
	if err != nil {
		return posts, err
	}
	idOrder := map[uint]int{}
	for i, id := range postIds {
		idOrder[id] = i
	}
	sort.Slice(posts, func(i, j int)bool{
		return idOrder[posts[i].ID] < idOrder[posts[j].ID]
	})
	return posts, nil
}
