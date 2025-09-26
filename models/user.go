package models

import (
	"time"

)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Firstname string `gorm:"size:100;not null" json:"firstname"`
	Lastname  string `gorm:"size:100;not null" json:"lastname"`
	Username  string `gorm:"size:50;unique;not null" json:"username"`
	Email     string `gorm:"size:100;unique;not null" json:"email"`
	Password  string `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Posts     []Post     `json:"posts"`
}