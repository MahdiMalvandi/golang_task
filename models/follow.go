package models

import "time"

type Follow struct {
	ID          uint `gorm:"primaryKey"`
	FollowerID  uint `gorm:"not null;uniqueIndex:idx_follower_following"`
	Follower    User `gorm:"foreignKey:FollowerID"`
	FollowingID uint `gorm:"not null;uniqueIndex:idx_follower_following"`
	Following   User `gorm:"foreignKey:FollowingID"`
	CreatedAt   time.Time
}
