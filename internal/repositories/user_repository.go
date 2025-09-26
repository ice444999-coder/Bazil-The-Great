package repositories

import (
	interfaces "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"gorm.io/gorm"
)

// Ensure UserRepository implements the interface
var _ interfaces.UserRepository = &UserRepository{}

type UserRepository struct {
	DB *gorm.DB
}

// Constructor
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// Create a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

// Get user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Get user by ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
