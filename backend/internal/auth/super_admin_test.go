package auth

import (
	"context"
	"testing"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// fakeUserStore is an in-memory implementation of domain.UserStore for tests.
type fakeUserStore struct {
	usersByIdentifier       map[string]*domain.User
	usersByExternalIdentity map[string]*domain.User // key = provider|subject
	createCalls             int
	updateCalls             int
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{
		usersByIdentifier:       make(map[string]*domain.User),
		usersByExternalIdentity: make(map[string]*domain.User),
	}
}

func externalIdentityKey(provider, subject string) string {
	return provider + "|" + subject
}

func (store *fakeUserStore) GetUserByIdentifier(_ context.Context, userIdentifier string) (*domain.User, error) {
	if user, ok := store.usersByIdentifier[userIdentifier]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (store *fakeUserStore) GetUserByExternalIdentity(_ context.Context, provider string, subject string) (*domain.User, error) {
	if user, ok := store.usersByExternalIdentity[externalIdentityKey(provider, subject)]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (store *fakeUserStore) CreateUser(_ context.Context, user *domain.User) error {
	store.createCalls++
	store.usersByIdentifier[user.UserIdentifier] = user
	store.usersByExternalIdentity[externalIdentityKey(user.ExternalIdentityProvider, user.ExternalIdentitySubject)] = user
	return nil
}

func (store *fakeUserStore) UpdateUser(_ context.Context, user *domain.User) error {
	store.updateCalls++
	store.usersByIdentifier[user.UserIdentifier] = user
	store.usersByExternalIdentity[externalIdentityKey(user.ExternalIdentityProvider, user.ExternalIdentitySubject)] = user
	return nil
}

func (store *fakeUserStore) ListUsers(_ context.Context, _ int, _ int) ([]domain.User, error) {
	users := make([]domain.User, 0, len(store.usersByIdentifier))
	for _, user := range store.usersByIdentifier {
		users = append(users, *user)
	}
	return users, nil
}

// fakeAPIKeyStore is an in-memory implementation of domain.APIKeyStore for tests.
type fakeAPIKeyStore struct {
	keysByIdentifier map[string]*domain.APIKey
	keysByHash       map[string]*domain.APIKey
	createCalls      int
	deleteCalls      int
}

func newFakeAPIKeyStore() *fakeAPIKeyStore {
	return &fakeAPIKeyStore{
		keysByIdentifier: make(map[string]*domain.APIKey),
		keysByHash:       make(map[string]*domain.APIKey),
	}
}

func (store *fakeAPIKeyStore) CreateAPIKey(_ context.Context, apiKey *domain.APIKey) error {
	store.createCalls++
	store.keysByIdentifier[apiKey.APIKeyIdentifier] = apiKey
	store.keysByHash[apiKey.HashedKey] = apiKey
	return nil
}

func (store *fakeAPIKeyStore) GetAPIKeyByHash(_ context.Context, hashedKey string) (*domain.APIKey, error) {
	if key, ok := store.keysByHash[hashedKey]; ok {
		return key, nil
	}
	return nil, domain.ErrAPIKeyNotFound
}

func (store *fakeAPIKeyStore) ListAPIKeysByUser(_ context.Context, userIdentifier string) ([]domain.APIKey, error) {
	keys := make([]domain.APIKey, 0)
	for _, key := range store.keysByIdentifier {
		if key.UserIdentifier == userIdentifier {
			keys = append(keys, *key)
		}
	}
	return keys, nil
}

func (store *fakeAPIKeyStore) DeleteAPIKey(_ context.Context, apiKeyIdentifier string) error {
	store.deleteCalls++
	if existing, ok := store.keysByIdentifier[apiKeyIdentifier]; ok {
		delete(store.keysByHash, existing.HashedKey)
	}
	delete(store.keysByIdentifier, apiKeyIdentifier)
	return nil
}

func (store *fakeAPIKeyStore) UpdateLastUsedTime(_ context.Context, _ string) error {
	return nil
}

// ─── Tests ────────────────────────────────────────────────────────────────

func TestSeedSuperAdmin_SkipsWhenNoAPIKeyConfigured(t *testing.T) {
	userStore := newFakeUserStore()
	apiKeyStore := newFakeAPIKeyStore()

	userIdentifier, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userIdentifier != "" {
		t.Errorf("expected empty user identifier when no key configured, got %q", userIdentifier)
	}
	if userStore.createCalls != 0 {
		t.Errorf("expected no user creation, got %d", userStore.createCalls)
	}
	if apiKeyStore.createCalls != 0 {
		t.Errorf("expected no api key creation, got %d", apiKeyStore.createCalls)
	}
}

func TestSeedSuperAdmin_CreatesUserAndKeyOnFirstRun(t *testing.T) {
	userStore := newFakeUserStore()
	apiKeyStore := newFakeAPIKeyStore()
	rawKey := "my-admin-key"

	userIdentifier, err := SeedSuperAdmin(
		context.Background(),
		userStore,
		apiKeyStore,
		"admin@example.com",
		"Admin",
		rawKey,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userIdentifier == "" {
		t.Fatal("expected non-empty user identifier")
	}
	if userStore.createCalls != 1 {
		t.Errorf("expected 1 user creation, got %d", userStore.createCalls)
	}
	if apiKeyStore.createCalls != 1 {
		t.Errorf("expected 1 api key creation, got %d", apiKeyStore.createCalls)
	}

	createdUser, err := userStore.GetUserByExternalIdentity(context.Background(), superAdminProvider, superAdminSubject)
	if err != nil {
		t.Fatalf("expected super admin user to exist: %v", err)
	}
	if createdUser.Email != "admin@example.com" || createdUser.DisplayName != "Admin" {
		t.Errorf("unexpected user fields: %+v", createdUser)
	}

	storedKey, err := apiKeyStore.GetAPIKeyByHash(context.Background(), HashAPIKey(rawKey))
	if err != nil {
		t.Fatalf("expected api key to be stored: %v", err)
	}
	if storedKey.UserIdentifier != userIdentifier {
		t.Errorf("expected key tied to user %q, got %q", userIdentifier, storedKey.UserIdentifier)
	}
	if storedKey.Label != superAdminKeyLabel {
		t.Errorf("expected label %q, got %q", superAdminKeyLabel, storedKey.Label)
	}
}

func TestSeedSuperAdmin_UsesDefaultEmailAndNameWhenOmitted(t *testing.T) {
	userStore := newFakeUserStore()
	apiKeyStore := newFakeAPIKeyStore()

	userIdentifier, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "", "", "raw-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user := userStore.usersByIdentifier[userIdentifier]
	if user == nil {
		t.Fatal("expected user to be stored")
	}
	if user.Email != "admin@localhost" {
		t.Errorf("expected default email admin@localhost, got %q", user.Email)
	}
	if user.DisplayName != "Super Admin" {
		t.Errorf("expected default display name 'Super Admin', got %q", user.DisplayName)
	}
}

func TestSeedSuperAdmin_IsIdempotentWhenKeyUnchanged(t *testing.T) {
	userStore := newFakeUserStore()
	apiKeyStore := newFakeAPIKeyStore()
	rawKey := "stable-key"

	firstIdentifier, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "a@b.com", "A", rawKey)
	if err != nil {
		t.Fatalf("first seed error: %v", err)
	}
	secondIdentifier, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "a@b.com", "A", rawKey)
	if err != nil {
		t.Fatalf("second seed error: %v", err)
	}

	if firstIdentifier != secondIdentifier {
		t.Errorf("expected same user identifier across runs, got %q and %q", firstIdentifier, secondIdentifier)
	}
	if userStore.createCalls != 1 {
		t.Errorf("expected only 1 user creation across both runs, got %d", userStore.createCalls)
	}
	if apiKeyStore.createCalls != 1 {
		t.Errorf("expected only 1 api key creation across both runs, got %d", apiKeyStore.createCalls)
	}
	if apiKeyStore.deleteCalls != 0 {
		t.Errorf("expected no api key deletions when key is unchanged, got %d", apiKeyStore.deleteCalls)
	}
}

func TestSeedSuperAdmin_ReplacesKeyWhenRawKeyChanges(t *testing.T) {
	userStore := newFakeUserStore()
	apiKeyStore := newFakeAPIKeyStore()

	firstIdentifier, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "a@b.com", "A", "old-key")
	if err != nil {
		t.Fatalf("first seed error: %v", err)
	}

	secondIdentifier, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "a@b.com", "A", "new-key")
	if err != nil {
		t.Fatalf("second seed error: %v", err)
	}

	if firstIdentifier != secondIdentifier {
		t.Errorf("expected same user identifier, got %q and %q", firstIdentifier, secondIdentifier)
	}

	if _, err := apiKeyStore.GetAPIKeyByHash(context.Background(), HashAPIKey("old-key")); err == nil {
		t.Error("expected old key to be removed")
	}

	newKey, err := apiKeyStore.GetAPIKeyByHash(context.Background(), HashAPIKey("new-key"))
	if err != nil {
		t.Fatalf("expected new key to be stored: %v", err)
	}
	if newKey.UserIdentifier != firstIdentifier {
		t.Errorf("expected new key tied to the same user, got %q", newKey.UserIdentifier)
	}

	if apiKeyStore.deleteCalls < 1 {
		t.Errorf("expected old super-admin key to be deleted, got %d deletions", apiKeyStore.deleteCalls)
	}
}

func TestSeedSuperAdmin_UpdatesUserWhenEmailOrNameChanges(t *testing.T) {
	userStore := newFakeUserStore()
	apiKeyStore := newFakeAPIKeyStore()

	if _, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "old@b.com", "Old", "key"); err != nil {
		t.Fatalf("first seed error: %v", err)
	}

	if _, err := SeedSuperAdmin(context.Background(), userStore, apiKeyStore, "new@b.com", "New", "key"); err != nil {
		t.Fatalf("second seed error: %v", err)
	}

	user, err := userStore.GetUserByExternalIdentity(context.Background(), superAdminProvider, superAdminSubject)
	if err != nil {
		t.Fatalf("expected user to exist: %v", err)
	}
	if user.Email != "new@b.com" || user.DisplayName != "New" {
		t.Errorf("expected updated email/name, got %+v", user)
	}
	if userStore.updateCalls < 1 {
		t.Error("expected at least one UpdateUser call")
	}
}
