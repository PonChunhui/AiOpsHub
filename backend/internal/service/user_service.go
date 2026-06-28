package service

import (
	"errors"

	"github.com/aiops/AiOpsHub/backend/internal/database"
	"github.com/aiops/AiOpsHub/backend/internal/model"
	"github.com/aiops/AiOpsHub/backend/internal/repository"
	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo: repository.NewUserRepository(),
	}
}

func (s *UserService) Register(username, email, password, role string) (*model.User, error) {
	logger.Info("Register called: username=" + username + ", email=" + email)

	if database.DB == nil {
		logger.Error("database.DB is nil in Register")
		return nil, errors.New("database not initialized")
	}

	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		logger.Error("username already exists")
		return nil, errors.New("username already exists")
	}

	database.DB.Model(&model.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		logger.Error("email already exists")
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("bcrypt failed: " + err.Error())
		return nil, err
	}

	user := &model.User{
		ID:       uuid.New().String(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	logger.Info("Creating user in database")
	if err := s.repo.Create(user); err != nil {
		logger.Error("repo.Create failed: " + err.Error())
		return nil, err
	}

	logger.Info("User created successfully: " + user.ID)
	return user, nil
}

func (s *UserService) Login(username, password string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

func (s *UserService) GetByID(id string) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) List(limit, offset int) ([]model.User, error) {
	return s.repo.List(limit, offset)
}

func (s *UserService) Update(id, email, role string) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if email != "" {
		user.Email = email
	}
	if role != "" {
		user.Role = role
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(id string) error {
	return s.repo.Delete(id)
}
