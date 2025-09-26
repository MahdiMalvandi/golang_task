package repositories

import (
	"errors"
	"fmt"
	"golang_task/models"
	"log"
	"strings"

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
	db *gorm.DB
}

// Follow repository constructor
func NewFollowRepository(db *gorm.DB) FollowRepositoryInterface {
	return &followRepository{
		db: db,

	}
}

// Follow repository methods

// This method allows a user to follow another user
func (r *followRepository) Follow(followerID, followingID uint) error {
	var followObject = models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
    var following models.User
	    if err := r.db.First(&following, followObject.FollowingID).Error; err != nil {
        return fmt.Errorf("following user not found")
    }
	
	if err := r.db.Create(&followObject).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") ||
			strings.Contains(err.Error(), "constraint failed") ||
			strings.Contains(err.Error(), "duplicate") {
			log.Println("You have already followed this user")
			return fmt.Errorf("you have already followed this user")
		}
		return err
	}
	return nil
}

// This method allows a user to unfollow another user
func (r *followRepository) UnFollow(followerID, followingID uint) error {
	var followObject = models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
    var following models.User

    if err := r.db.First(&following, followObject.FollowingID).Error; err != nil {
        return fmt.Errorf("following user not found")
    }
	
	result := r.db.Where("follower_id = ? AND following_id = ? ", followObject.FollowerID, followObject.FollowingID).Delete(&models.Follow{})
	if result.RowsAffected == 0{
		return fmt.Errorf("you have not followed this user")
	}
	return result.Error
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

func (r *followRepository) GetFollowers(followingID uint) ([]models.User, error) {
	var followers []models.User
	if err := r.db.Model(&models.Follow{}).Select("users.*").
		Joins("JOIN users ON users.id = follows.follower_id").
		Where("follows.following_id = ?", followingID).
		Scan(&followers).Error; err != nil {
		log.Println("Error getting followers:", err)
		return nil, err
	}
	return followers, nil
}

func (r *followRepository) GetFollowings(followerID uint) ([]models.User, error) {
	var followings []models.User
	if err := r.db.Model(&models.Follow{}).Select("users.*").
		Joins("JOIN users ON users.id = follows.following_id").
		Where("follows.follower_id = ?", followerID).
		Scan(&followings).Error; err != nil {
		log.Println("Error getting followings:", err)
		return nil, err
	}
	return followings, nil
}
