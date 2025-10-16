package services

import (
	repo "ares_api/internal/interfaces/repository"
	service "ares_api/internal/interfaces/service"
	"ares_api/internal/models"
	"ares_api/internal/auth"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Ensure UserService implements the interface
var _ service.UserService = &UserService{}

type UserService struct {
	Repo repo.UserRepository
	
}

// Constructor
func NewUserService(repo repo.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

// Signup a new user
func (s *UserService) Signup(username, email, password string) error {
	if existingUser, _ := s.Repo.GetByUsername(username); existingUser != nil {
		return errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	return s.Repo.Create(user)
}

// Login user and return both access and refresh tokens
func (s *UserService) Login(username, password string) (uint, string, string, error) {
	user, err := s.Repo.GetByUsername(username)
	if err != nil {
		return 0, "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return 0,"", "", errors.New("invalid password")
	}

	// Generate Access Token
	accessToken, err := auth.GenerateJWT(user.ID)
	if err != nil {
		return 0 ,"", "", err
	}

	// Generate Refresh Token
	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return 0,"", "", err
	}

	return user.ID, accessToken, refreshToken, nil
}

// Refresh token to generate a new access token
func (s *UserService) Refresh(refreshToken string) (string, error) {
	userID, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	accessToken, err := auth.GenerateJWT(userID.UserID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.Repo.GetByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
