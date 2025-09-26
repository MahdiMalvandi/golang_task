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
	UpdatePost(post *models.Post ,userID uint, updates interface{}) error
	DeletePost(post *models.Post ,userID uint) error
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
		return err
	}

	// Send post id and author id to queue to add in followers timeline
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
func (r *postRepository) UpdatePost(post *models.Post, userID uint, updates interface{}) error {

	if post.AuthorID != userID {
		return fmt.Errorf("you are not the author of this post")
	}

	if err := r.db.Model(post).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// This method deletes a post
//
// If the error is nil, the post was deleted successfully.
func (r *postRepository) DeletePost(post *models.Post, userID uint) error {

	if post.AuthorID != userID {
		return fmt.Errorf("you are not the author of this post")
	}

	if err := r.db.Delete(post).Error; err != nil {
		return err
	}

	return nil
}


func (r *postRepository) GetFollowingsPosts(userID uint, start, end int64) (posts []models.Post, err error) {
	fmt.Println("got from database")
	followRepo := NewFollowRepository(r.db)
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
	return posts, nil
}

func (r *postRepository) GetTimeline(userID uint, start, end int64) ([]models.Post, error) {
	var posts []models.Post
	ctx := context.Background()
	key := fmt.Sprintf("timeline:%d", userID)
	result, err := r.rdb.ZRevRange(ctx, key, start, end).Result()
	if err != nil {
		return posts, fmt.Errorf("error")
	}

	fmt.Println(result)
	if len(result) == 0 {
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
		fmt.Println("got from redis")
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
	
	return posts, nil
}
