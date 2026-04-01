package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) domain.UserStore {
	return &UserStore{
		db: db,
	}
}

func (store *UserStore) GetUserById(userId domain.UserId) (*domain.User, error) {
	user := &domain.User{}

	err := store.db.QueryRow(
		`SELECT id, email, display_name, external_identity_provider, external_identity_subject
		 FROM users
		 WHERE id = $1`,
		userId,
	).Scan(
		&user.UserId,
		&user.Email,
		&user.DisplayName,
		&user.ExternalIdentityProvider,
		&user.ExternalIdentitySubject,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	return user, nil
}

func (store *UserStore) GetUserByExternalIdentity(provider string, subject string) (*domain.User, error) {
	user := &domain.User{}

	err := store.db.QueryRow(
		`SELECT id, email, display_name, external_identity_provider, external_identity_subject
		 FROM users
		 WHERE external_identity_provider = $1 AND external_identity_subject = $2`,
		provider,
		subject,
	).Scan(
		&user.UserId,
		&user.Email,
		&user.DisplayName,
		&user.ExternalIdentityProvider,
		&user.ExternalIdentitySubject,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to query user by external identity: %w", err)
	}

	return user, nil
}

func (store *UserStore) CreateUser(user *domain.User) (*domain.User, error) {
	createdUser := &domain.User{}

	err := store.db.QueryRow(
		`INSERT INTO users (email, display_name, external_identity_provider, external_identity_subject)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, display_name, external_identity_provider, external_identity_subject`,
		user.Email,
		user.DisplayName,
		user.ExternalIdentityProvider,
		user.ExternalIdentitySubject,
	).Scan(
		&createdUser.UserId,
		&createdUser.Email,
		&createdUser.DisplayName,
		&createdUser.ExternalIdentityProvider,
		&createdUser.ExternalIdentitySubject,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (store *UserStore) UpdateUser(userId domain.UserId, email string, displayName string) (*domain.User, error) {
	updatedUser := &domain.User{}

	err := store.db.QueryRow(
		`UPDATE users
		 SET email = $1, display_name = $2
		 WHERE id = $3
		 RETURNING id, email, display_name, external_identity_provider, external_identity_subject`,
		email,
		displayName,
		userId,
	).Scan(
		&updatedUser.UserId,
		&updatedUser.Email,
		&updatedUser.DisplayName,
		&updatedUser.ExternalIdentityProvider,
		&updatedUser.ExternalIdentitySubject,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (store *UserStore) ListUsers(limit int, offset int) ([]*domain.User, error) {
	rows, err := store.db.Query(
		`SELECT id, email, display_name, external_identity_provider, external_identity_subject
		 FROM users
		 ORDER BY id
		 LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User

	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.UserId,
			&user.Email,
			&user.DisplayName,
			&user.ExternalIdentityProvider,
			&user.ExternalIdentitySubject,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
