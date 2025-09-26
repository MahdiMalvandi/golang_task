package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang_task/models"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Follow Repository interface
type FollowRepositoryInterface interface {
	Follow(followerID, followingID uint) error
	IsFollowing(followerID, followingID uint) (bool, error)
	GetFollowers(followingID uint) ([]models.User, error)
	GetFollowings(followerID uint) ([]models.User, error)
	UnFollow(followerID, followingID uint) error
}

// Follow repository struct
type followRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

// Follow repository constructor
func NewFollowRepository(db *gorm.DB, rdb *redis.Client) FollowRepositoryInterface {
	return &followRepository{
		db:  db,
		rdb: rdb,
	}
}

// Follow repository methods

// This method allows a user to follow another user
func (r *followRepository) Follow(followerID, followingID uint) error {
	var followObject = models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	// Get following from db and check if it exists
	var following models.User
	if err := r.db.First(&following, followObject.FollowingID).Error; err != nil {
		log.Printf("[ERROR] User %d tried to follow user %d but the following user not found", followerID, followingID)

		return fmt.Errorf("following user not found")
	}

	// Check follow situation
	if err := r.db.Create(&followObject).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") ||
			strings.Contains(err.Error(), "constraint failed") ||
			strings.Contains(err.Error(), "Duplicate") {
			log.Printf("[ERROR] User %d already followed user %d", followerID, followingID)

			return fmt.Errorf("you have already followed this user")
		}
		log.Printf("[ERROR] Error while user %d tried to follow user %d: %v", followerID, followingID, err)

		return err
	}
	log.Printf("[INFO] User %d followed user %d successfully", followerID, followingID)
	postRepo := NewPostRepository(r.db, r.rdb)
	posts, err := postRepo.GetByAuthorID(followingID)
	if err != nil {
		log.Printf("[ERROR] Could not fetch posts of user %d for follower %d: %v", followingID, followerID, err)
	} else {
		var pushData []interface{}
		for _, post := range posts {
			data, _ := json.Marshal(map[string]interface{}{
				"post_id":    post.ID,
				"author_id":  followingID,
				"created_at": uint(post.CreatedAt.Unix()),
				"is_add":     true,
			})
			pushData = append(pushData, data)
		}

		if len(pushData) > 0 {
			err := r.rdb.RPush(context.Background(), "post_queue", pushData...).Err()
			if err != nil {
				log.Printf("[ERROR] Failed to push posts of user %d into queue for follower %d: %v", followingID, followerID, err)
			} else {
				log.Printf("[INFO] Queued %d posts of user %d for follower %d", len(pushData), followingID, followerID)
			}
		}
	}
	return nil
}

// This method allows a user to unfollow another user
func (r *followRepository) UnFollow(followerID, followingID uint) error {
	var followObject = models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	// Get following from db
	var following models.User

	if err := r.db.First(&following, followObject.FollowingID).Error; err != nil {
		log.Printf("[ERROR] User %d tried to unfollow user %d but following user not found", followerID, followingID)
		return fmt.Errorf("following user not found")
	}

	// Checking if the user is following the user with id followingID
	result := r.db.Where("follower_id = ? AND following_id = ? ", followObject.FollowerID, followObject.FollowingID).Delete(&models.Follow{})
	if result.RowsAffected == 0 {
		log.Printf("[ERROR] User %d tried to unfollow user %d but was not following", followerID, followingID)

		return fmt.Errorf("you have not followed this user")
	}
	if result.Error != nil {
		log.Printf("[ERROR] Error while user %d tried to unfollow user %d: %v", followerID, followingID, result.Error)
		return result.Error
	}
	log.Printf("[INFO] User %d unfollowed user %d successfully", followerID, followingID)

	timelineKey := fmt.Sprintf("timeline:%d", followerID)

	ctx := context.Background()
	posts, _ := NewPostRepository(r.db, r.rdb).GetByAuthorID(followingID)
	var postIDs []interface{}

	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}
	removedCount, err := r.rdb.ZRem(ctx, timelineKey, postIDs...).Result()
	if err != nil {
		log.Printf("[ERROR] Failed to remove posts: %v", err)
	} else {
		log.Printf("[INFO] Removed %d posts from timeline", removedCount)
	}

	return nil
}

func (r *followRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var followObject models.Follow
	if err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&followObject).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// This function get the user's followers
func (r *followRepository) GetFollowers(followingID uint) ([]models.User, error) {
	var followers []models.User
	if err := r.db.Model(&models.Follow{}).Select("users.*").
		Joins("JOIN users ON users.id = follows.follower_id").
		Where("follows.following_id = ?", followingID).
		Scan(&followers).Error; err != nil {
		log.Printf("[ERROR] [ERROR]Error getting followers for user %d: %v", followingID, err)

		return nil, err
	}
	log.Printf("[INFO] Fetched %d followers for user %d", len(followers), followingID)

	return followers, nil
}

// This function get the user's followings
func (r *followRepository) GetFollowings(followerID uint) ([]models.User, error) {
	var followings []models.User
	if err := r.db.Model(&models.Follow{}).Select("users.*").
		Joins("JOIN users ON users.id = follows.following_id").
		Where("follows.follower_id = ?", followerID).
		Scan(&followings).Error; err != nil {
		log.Printf("[ERROR] Error getting followings for user %d: %v", followerID, err)

		return nil, err
	}
	log.Printf("[INFO] Fetched %d followings for user %d", len(followings), followerID)

	return followings, nil
}
