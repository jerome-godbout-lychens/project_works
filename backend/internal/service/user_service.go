package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// UserService orchestrates user operations.
type UserService struct {
	userStore domain.UserStore
}

// NewUserService creates a new UserService.
func NewUserService(userStore domain.UserStore) *UserService {
	return &UserService{userStore: userStore}
}

func (s *UserService) GetUserById(ctx context.Context, userId string) (*domain.User, error) {
	return s.userStore.GetUserById(ctx, userId)
}

func (s *UserService) ListUsers(ctx context.Context, limit int, offset int) ([]domain.User, error) {
	return s.userStore.ListUsers(ctx, limit, offset)
}

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	return s.userStore.CreateUser(ctx, user)
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.userStore.UpdateUser(ctx, user)
}

// GetOrCreateUserFromOIDC returns an existing user matched by external identity,
// or creates a new one if no match is found.
func (s *UserService) GetOrCreateUserFromOIDC(
	ctx context.Context,
	provider string,
	subject string,
	email string,
	displayName string,
) (*domain.User, error) {
	existingUser, err := s.userStore.GetUserByExternalIdentity(ctx, provider, subject)
	if err == nil {
		return existingUser, nil
	}

	if err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}

	// No match found — create a new user.
	user := &domain.User{
		UserId:                   uuid.New().String(),
		Email:                    email,
		DisplayName:              displayName,
		ExternalIdentityProvider: provider,
		ExternalIdentitySubject:  subject,
	}

	if err := s.userStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user from OIDC: %w", err)
	}

	return user, nil
}
