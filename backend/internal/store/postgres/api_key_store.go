package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type APIKeyStore struct {
	db *sql.DB
}

func NewAPIKeyStore(db *sql.DB) domain.APIKeyStore {
	return &APIKeyStore{
		db: db,
	}
}

func (store *APIKeyStore) CreateAPIKey(apiKey *domain.APIKey) (*domain.APIKey, error) {
	createdAPIKey := &domain.APIKey{}

	err := store.db.QueryRow(
		`INSERT INTO api_keys (user_id, hashed_key, label, created_time)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, hashed_key, label, created_time, last_used_time`,
		apiKey.UserId,
		apiKey.HashedKey,
		apiKey.Label,
		apiKey.CreatedTime,
	).Scan(
		&createdAPIKey.APIKeyId,
		&createdAPIKey.UserId,
		&createdAPIKey.HashedKey,
		&createdAPIKey.Label,
		&createdAPIKey.CreatedTime,
		&createdAPIKey.LastUsedTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return createdAPIKey, nil
}

func (store *APIKeyStore) GetAPIKeyByHash(hashedKey string) (*domain.APIKey, error) {
	apiKey := &domain.APIKey{}

	err := store.db.QueryRow(
		`SELECT id, user_id, hashed_key, label, created_time, last_used_time
		 FROM api_keys
		 WHERE hashed_key = $1`,
		hashedKey,
	).Scan(
		&apiKey.APIKeyId,
		&apiKey.UserId,
		&apiKey.HashedKey,
		&apiKey.Label,
		&apiKey.CreatedTime,
		&apiKey.LastUsedTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("api key not found: %w", err)
		}
		return nil, fmt.Errorf("failed to query api key by hash: %w", err)
	}

	return apiKey, nil
}

func (store *APIKeyStore) ListAPIKeysByUser(userId domain.UserId, limit int, offset int) ([]*domain.APIKey, error) {
	rows, err := store.db.Query(
		`SELECT id, user_id, hashed_key, label, created_time, last_used_time
		 FROM api_keys
		 WHERE user_id = $1
		 ORDER BY created_time DESC
		 LIMIT $2 OFFSET $3`,
		userId,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query api keys by user: %w", err)
	}
	defer rows.Close()

	var apiKeys []*domain.APIKey

	for rows.Next() {
		apiKey := &domain.APIKey{}
		err := rows.Scan(
			&apiKey.APIKeyId,
			&apiKey.UserId,
			&apiKey.HashedKey,
			&apiKey.Label,
			&apiKey.CreatedTime,
			&apiKey.LastUsedTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating api key rows: %w", err)
	}

	return apiKeys, nil
}

func (store *APIKeyStore) DeleteAPIKey(apiKeyId domain.APIKeyId) error {
	result, err := store.db.Exec(
		`DELETE FROM api_keys
		 WHERE id = $1`,
		apiKeyId,
	)

	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("api key not found")
	}

	return nil
}

func (store *APIKeyStore) UpdateLastUsedTime(apiKeyId domain.APIKeyId, lastUsedTime time.Time) error {
	result, err := store.db.Exec(
		`UPDATE api_keys
		 SET last_used_time = $1
		 WHERE id = $2`,
		lastUsedTime,
		apiKeyId,
	)

	if err != nil {
		return fmt.Errorf("failed to update last used time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("api key not found")
	}

	return nil
}
