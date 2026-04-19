package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// APIKeyStore implements domain.APIKeyStore using PostgreSQL.
type APIKeyStore struct {
	db *sql.DB
}

func NewAPIKeyStore(db *sql.DB) domain.APIKeyStore {
	return &APIKeyStore{db: db}
}

func (store *APIKeyStore) CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	err := store.db.QueryRowContext(ctx,
		`INSERT INTO api_keys (api_key_identifier, user_identifier, hashed_key, label, created_time)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING api_key_identifier`,
		apiKey.APIKeyIdentifier, apiKey.UserIdentifier, apiKey.HashedKey, apiKey.Label, apiKey.CreatedTime,
	).Scan(&apiKey.APIKeyIdentifier)
	if err != nil {
		return fmt.Errorf("failed to create api key: %w", err)
	}
	return nil
}

func (store *APIKeyStore) GetAPIKeyByHash(ctx context.Context, hashedKey string) (*domain.APIKey, error) {
	apiKey := &domain.APIKey{}
	err := store.db.QueryRowContext(ctx,
		`SELECT api_key_identifier, user_identifier, hashed_key, label, created_time, last_used_time
		 FROM api_keys WHERE hashed_key = $1`, hashedKey,
	).Scan(&apiKey.APIKeyIdentifier, &apiKey.UserIdentifier, &apiKey.HashedKey,
		&apiKey.Label, &apiKey.CreatedTime, &apiKey.LastUsedTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("failed to query api key by hash: %w", err)
	}
	return apiKey, nil
}

func (store *APIKeyStore) ListAPIKeysByUser(ctx context.Context, userIdentifier string) ([]domain.APIKey, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT api_key_identifier, user_identifier, hashed_key, label, created_time, last_used_time
		 FROM api_keys WHERE user_identifier = $1 ORDER BY created_time DESC`, userIdentifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query api keys by user: %w", err)
	}
	defer rows.Close()

	var apiKeys []domain.APIKey
	for rows.Next() {
		var apiKey domain.APIKey
		if err := rows.Scan(&apiKey.APIKeyIdentifier, &apiKey.UserIdentifier, &apiKey.HashedKey,
			&apiKey.Label, &apiKey.CreatedTime, &apiKey.LastUsedTime); err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}
	return apiKeys, rows.Err()
}

func (store *APIKeyStore) DeleteAPIKey(ctx context.Context, apiKeyIdentifier string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM api_keys WHERE api_key_identifier = $1`, apiKeyIdentifier)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

func (store *APIKeyStore) UpdateLastUsedTime(ctx context.Context, apiKeyIdentifier string) error {
	result, err := store.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_time = NOW() WHERE api_key_identifier = $1`, apiKeyIdentifier)
	if err != nil {
		return fmt.Errorf("failed to update last used time: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}
