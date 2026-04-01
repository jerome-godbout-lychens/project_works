package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type UserService struct {
	userStore domain.UserStore
}

func NewUserService(userStore domain.UserStore) *UserService {
	return &UserService{
		userStore: userStore,
	}
}

func (s *UserService) GetUserById(ctx context.Context, userId string) (*domain.User, error) {
	return s.userStore.GetUserById(ctx, userId)
}

func (s *UserService) ListUsers(ctx context.Context, query *domain.UserQuery) ([]*domain.User, error) {
	return s.userStore.ListUsers(ctx, query)
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	return s.userStore.CreateUser(ctx, user)
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.userStore.UpdateUser(ctx, user)
}

func (s *UserService) GetOrCreateUserFromOIDC(
	ctx context.Context,
	provider string,
	subject string,
	email string,
	displayName string,
) (*domain.User, error) {
	// Try to find existing user by external identity
	existingUser, err := s.userStore.GetUserByExternalIdentity(ctx, provider, subject)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	// If not found (and it's not another error), create new user
	if _, ok := err.(*domain.NotFoundError); !ok && err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}

	// Create new user
	user := &domain.User{
		Id:          uuid.New().String(),
		Email:       email,
		DisplayName: displayName,
		CreationTime: time.Now().UTC(),
	}

	// Add external identity
	user.ExternalIdentities = append(user.ExternalIdentities, &domain.ExternalIdentity{
		Provider: provider,
		Subject:  subject,
	})

	if err := s.userStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user from OIDC: %w", err)
	}

	return user, nil
}
