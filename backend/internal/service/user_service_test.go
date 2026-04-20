package service

import (
	"context"
	"testing"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

func TestUserService_GetOrCreateUserFromOIDC_CreatesNewUser(t *testing.T) {
	store := newFakeUserStore()
	service := NewUserService(store)

	user, err := service.GetOrCreateUserFromOIDC(
		context.Background(),
		"office365",
		"sub-abc",
		"new@example.com",
		"New User",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserIdentifier == "" {
		t.Error("expected UserIdentifier to be assigned")
	}
	if user.Email != "new@example.com" || user.DisplayName != "New User" {
		t.Errorf("unexpected user fields: %+v", user)
	}
	if _, err := store.GetUserByExternalIdentity(context.Background(), "office365", "sub-abc"); err != nil {
		t.Errorf("expected newly-created user to be persisted: %v", err)
	}
}

func TestUserService_GetOrCreateUserFromOIDC_ReturnsExistingUser(t *testing.T) {
	store := newFakeUserStore()
	existing := &domain.User{
		UserIdentifier:           "user-1",
		Email:                    "existing@example.com",
		DisplayName:              "Existing",
		ExternalIdentityProvider: "office365",
		ExternalIdentitySubject:  "sub-xyz",
	}
	_ = store.CreateUser(context.Background(), existing)

	service := NewUserService(store)
	got, err := service.GetOrCreateUserFromOIDC(
		context.Background(),
		"office365",
		"sub-xyz",
		"should@be.ignored",
		"Should Ignore",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserIdentifier != "user-1" {
		t.Errorf("expected existing user-1, got %q", got.UserIdentifier)
	}
	if got.Email != "existing@example.com" {
		t.Errorf("expected existing email preserved, got %q", got.Email)
	}
}

func TestUserService_CreateAndGetUser(t *testing.T) {
	store := newFakeUserStore()
	service := NewUserService(store)

	user := &domain.User{
		UserIdentifier: "u-1",
		Email:          "a@b.com",
		DisplayName:    "A",
	}
	if err := service.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := service.GetUserByIdentifier(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email != "a@b.com" {
		t.Errorf("unexpected email: %q", got.Email)
	}
}
