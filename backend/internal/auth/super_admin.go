package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

const (
	superAdminProvider = "local"
	superAdminSubject  = "super-admin"
	superAdminKeyLabel = "super-admin-config-key"
)

// SeedSuperAdmin ensures that a super admin user and API key exist in the
// database, based on the values provided in the application configuration.
// If the user already exists it is updated (email / display name). The API
// key is replaced whenever the configured raw key changes.
//
// Returns the super admin user ID (empty string when no key is configured).
// This function is idempotent — safe to call on every startup.
func SeedSuperAdmin(
	ctx context.Context,
	userStore domain.UserStore,
	apiKeyStore domain.APIKeyStore,
	email string,
	displayName string,
	rawAPIKey string,
) (string, error) {
	if rawAPIKey == "" {
		log.Println("super admin: no api key configured, skipping seed")
		return "", nil
	}

	if email == "" {
		email = "admin@localhost"
	}
	if displayName == "" {
		displayName = "Super Admin"
	}

	// ── Upsert user ─────────────────────────────────────────────
	existingUser, err := userStore.GetUserByExternalIdentity(ctx, superAdminProvider, superAdminSubject)
	var userIdentifier string

	if errors.Is(err, domain.ErrUserNotFound) {
		userIdentifier = uuid.New().String()
		newUser := &domain.User{
			UserIdentifier:                   userIdentifier,
			Email:                    email,
			DisplayName:              displayName,
			ExternalIdentityProvider: superAdminProvider,
			ExternalIdentitySubject:  superAdminSubject,
		}
		if createErr := userStore.CreateUser(ctx, newUser); createErr != nil {
			return "", createErr
		}
		log.Printf("super admin: created user %s (%s)", userIdentifier, email)
	} else if err != nil {
		return "", err
	} else {
		userIdentifier = existingUser.UserIdentifier
		// Update email / display name if they changed.
		if existingUser.Email != email || existingUser.DisplayName != displayName {
			existingUser.Email = email
			existingUser.DisplayName = displayName
			if updateErr := userStore.UpdateUser(ctx, existingUser); updateErr != nil {
				return "", updateErr
			}
			log.Printf("super admin: updated user %s (%s)", userIdentifier, email)
		}
	}

	// ── Upsert API key ──────────────────────────────────────────
	hashedKey := HashAPIKey(rawAPIKey)

	// Check whether the current hash already exists.
	existingKey, err := apiKeyStore.GetAPIKeyByHash(ctx, hashedKey)
	if err == nil && existingKey.UserIdentifier == userIdentifier {
		// Key is already present and belongs to the super admin — nothing to do.
		log.Printf("super admin: api key already up-to-date for user %s", userIdentifier)
		return userIdentifier, nil
	}

	// Remove any previous super-admin config key for this user so we don't
	// accumulate stale keys on every config change.
	existingKeys, listErr := apiKeyStore.ListAPIKeysByUser(ctx, userIdentifier)
	if listErr == nil {
		for _, key := range existingKeys {
			if key.Label == superAdminKeyLabel {
				_ = apiKeyStore.DeleteAPIKey(ctx, key.APIKeyIdentifier)
			}
		}
	}

	newAPIKey := &domain.APIKey{
		APIKeyIdentifier:    uuid.New().String(),
		UserIdentifier:      userIdentifier,
		HashedKey:   hashedKey,
		Label:       superAdminKeyLabel,
		CreatedTime: time.Now(),
	}
	if createErr := apiKeyStore.CreateAPIKey(ctx, newAPIKey); createErr != nil {
		return "", createErr
	}
	log.Printf("super admin: api key seeded for user %s", userIdentifier)
	return userIdentifier, nil
}
