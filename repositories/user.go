package repositories

import (
	"errors"
	"fmt"
	"golang_task/models"
	"golang_task/utils"
	"log"
	"strings"

	"gorm.io/gorm"
)

// User Repository interface
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(id uint, updates map[string]interface{}) error
	DeleteById(id uint) error
	DeleteByUsername(username string) error
}

// User Repository
type userRepository struct {
	db *gorm.DB
}

// User Repository Constructor
func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{
		db: db,
	}
}

// User Repository methods

// This method get a user data and create it and return error
//
// If the error is nil, the user was created successfully.
func (r *userRepository) Create(user *models.User) error {
	var err error

	// Hashing password
	user.Password, err = utils.HashPassword(user.Password)
	if err != nil {
		log.Printf("[ERROR] Error hashing password: %v", err)
		return err
	}

	// Create User and Error Handling
	if err := r.db.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") ||
			strings.Contains(err.Error(), "constraint failed") ||
			strings.Contains(err.Error(), "Duplicate") {
			log.Printf("[ERROR] User with username %s or email %s already exists", user.Username, user.Email)
			return fmt.Errorf("username or email already exists")
		}
		return err
	}

	return err
}

// This method retrieves a user by ID
//
// If the user is found, it returns the user. If not, it returns an error.
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] User with id %d not found", id)
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// This method retrieves a user by username
//
// If the user is found, it returns the user. If not, it returns an error.
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] User with username %s not found", username)
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// This method retrieves a user by email
//
// If the user is found, it returns the user. If not, it returns an error.
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] User with email %s not found", email)
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// This method updates a user by ID
//
// If the user is found, it updates the user with the provided data. If not, it returns an error.
func (r *userRepository) Update(id uint, updates map[string]interface{}) error {
	var err error

	user := models.User{
		ID: id,
	}
	if err = r.db.Model(&user).Updates(updates).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "Duplicate") {
			return fmt.Errorf("username or email already exists")
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] User with id %d not found", id)
			return err
		}
	}
	return err
}

// This method deletes a user by ID
//
// If the user is found, it deletes the user. If not, it returns an error.
func (r *userRepository) DeleteById(id uint) error {
	var err error
	if err = r.db.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] User with id %d not found", id)
			return fmt.Errorf("user not found")
		}
	}
	return err
}

// This method deletes a user by username
//
// If the user is found, it deletes the user. If not, it returns an error.
func (r *userRepository) DeleteByUsername(username string) error {
	var err error
	if err = r.db.Where("username = ?", username).Delete(&models.User{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[ERROR] User with username %s not found", username)
			return fmt.Errorf("user not found")
		}
	}
	return err
}
