package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/model"
	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/repository"
	"github.com/davidbadelllab/go-microservice-grpc-2023/pkg/cache"
)

// UserService handles user business logic
type UserService struct {
	repo  *repository.UserRepository
	cache *cache.Redis
}

// NewUserService creates a new UserService instance
func NewUserService(repo *repository.UserRepository, cache *cache.Redis) *UserService {
	return &UserService{
		repo:  repo,
		cache: cache,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, email, name string) (*model.User, error) {
	user := &model.User{
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Invalidate cache
	s.cache.Delete(ctx, "users:list")

	slog.Info("user created",
		slog.Int64("user_id", user.ID),
		slog.String("email", user.Email))

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)

	// Try to get from cache
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var user model.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			slog.Debug("cache hit", slog.String("key", cacheKey))
			return &user, nil
		}
	}

	// Get from database
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Cache the result
	if data, err := json.Marshal(user); err == nil {
		s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return user, nil
}

// ListUsers lists all users with pagination
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	offset := (page - 1) * pageSize

	users, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id int64, email, name string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.Email = email
	user.Name = name
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	s.cache.Delete(ctx, cacheKey)
	s.cache.Delete(ctx, "users:list")

	slog.Info("user updated",
		slog.Int64("user_id", user.ID),
		slog.String("email", user.Email))

	return user, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	s.cache.Delete(ctx, cacheKey)
	s.cache.Delete(ctx, "users:list")

	slog.Info("user deleted", slog.Int64("user_id", id))

	return nil
}
