package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// UserStore implements domain.UserStore using PostgreSQL.
type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) domain.UserStore {
	return &UserStore{db: db}
}

func (store *UserStore) GetUserByIdentifier(ctx context.Context, userIdentifier string) (*domain.User, error) {
	user := &domain.User{}
	err := store.db.QueryRowContext(ctx,
		`SELECT user_identifier, email, display_name, external_identity_provider, external_identity_subject
		 FROM users WHERE user_identifier = $1`, userIdentifier,
	).Scan(&user.UserIdentifier, &user.Email, &user.DisplayName,
		&user.ExternalIdentityProvider, &user.ExternalIdentitySubject)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return user, nil
}

func (store *UserStore) GetUserByExternalIdentity(ctx context.Context, provider string, subject string) (*domain.User, error) {
	user := &domain.User{}
	err := store.db.QueryRowContext(ctx,
		`SELECT user_identifier, email, display_name, external_identity_provider, external_identity_subject
		 FROM users WHERE external_identity_provider = $1 AND external_identity_subject = $2`,
		provider, subject,
	).Scan(&user.UserIdentifier, &user.Email, &user.DisplayName,
		&user.ExternalIdentityProvider, &user.ExternalIdentitySubject)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by external identity: %w", err)
	}
	return user, nil
}

func (store *UserStore) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := store.db.ExecContext(ctx,
		`INSERT INTO users (user_identifier, email, display_name, external_identity_provider, external_identity_subject)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.UserIdentifier, user.Email, user.DisplayName,
		user.ExternalIdentityProvider, user.ExternalIdentitySubject,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (store *UserStore) UpdateUser(ctx context.Context, user *domain.User) error {
	result, err := store.db.ExecContext(ctx,
		`UPDATE users SET email = $1, display_name = $2 WHERE user_identifier = $3`,
		user.Email, user.DisplayName, user.UserIdentifier,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (store *UserStore) ListUsers(ctx context.Context, limit int, offset int) ([]domain.User, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT user_identifier, email, display_name, external_identity_provider, external_identity_subject
		 FROM users ORDER BY user_identifier LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserIdentifier, &user.Email, &user.DisplayName,
			&user.ExternalIdentityProvider, &user.ExternalIdentitySubject); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
