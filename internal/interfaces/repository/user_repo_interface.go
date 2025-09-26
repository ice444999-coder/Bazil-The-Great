package Repositories
import "ares_api/internal/models"

// UserRepository defines methods for user data
type UserRepository interface {
	Create(user *models.User) error
	GetByUsername(username string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
}