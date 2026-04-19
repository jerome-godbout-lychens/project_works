package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// GroupStore implements domain.GroupStore using PostgreSQL.
type GroupStore struct {
	db *sql.DB
}

func NewGroupStore(db *sql.DB) domain.GroupStore {
	return &GroupStore{db: db}
}

func (store *GroupStore) GetGroupByIdentifier(ctx context.Context, groupIdentifier string) (*domain.Group, error) {
	group := &domain.Group{}
	err := store.db.QueryRowContext(ctx,
		`SELECT group_identifier, group_name FROM groups WHERE group_identifier = $1`, groupIdentifier,
	).Scan(&group.GroupIdentifier, &group.GroupName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to query group by id: %w", err)
	}
	return group, nil
}

func (store *GroupStore) ListGroups(ctx context.Context) ([]domain.Group, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT group_identifier, group_name FROM groups ORDER BY group_identifier`)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(&group.GroupIdentifier, &group.GroupName); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (store *GroupStore) CreateGroup(ctx context.Context, group *domain.Group) error {
	err := store.db.QueryRowContext(ctx,
		`INSERT INTO groups (group_name) VALUES ($1) RETURNING group_identifier`,
		group.GroupName,
	).Scan(&group.GroupIdentifier)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (store *GroupStore) DeleteGroup(ctx context.Context, groupIdentifier string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM groups WHERE group_identifier = $1`, groupIdentifier)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

func (store *GroupStore) AddUserToGroup(ctx context.Context, groupIdentifier string, userIdentifier string) error {
	_, err := store.db.ExecContext(ctx,
		`INSERT INTO group_memberships (group_identifier, user_identifier)
		 VALUES ($1, $2) ON CONFLICT (group_identifier, user_identifier) DO NOTHING`,
		groupIdentifier, userIdentifier,
	)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	return nil
}

func (store *GroupStore) RemoveUserFromGroup(ctx context.Context, groupIdentifier string, userIdentifier string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM group_memberships WHERE group_identifier = $1 AND user_identifier = $2`,
		groupIdentifier, userIdentifier,
	)
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found in group")
	}
	return nil
}

func (store *GroupStore) ListGroupsByUser(ctx context.Context, userIdentifier string) ([]domain.Group, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT g.group_identifier, g.group_name
		 FROM groups g
		 INNER JOIN group_memberships gm ON g.group_identifier = gm.group_identifier
		 WHERE gm.user_identifier = $1
		 ORDER BY g.group_identifier`, userIdentifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups by user: %w", err)
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(&group.GroupIdentifier, &group.GroupName); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (store *GroupStore) ListUsersByGroup(ctx context.Context, groupIdentifier string) ([]domain.User, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT u.user_identifier, u.email, u.display_name, u.external_identity_provider, u.external_identity_subject
		 FROM users u
		 INNER JOIN group_memberships gm ON u.user_identifier = gm.user_identifier
		 WHERE gm.group_identifier = $1
		 ORDER BY u.user_identifier`, groupIdentifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by group: %w", err)
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

// GroupProjectAccessStore implements domain.GroupProjectAccessStore using PostgreSQL.
type GroupProjectAccessStore struct {
	db *sql.DB
}

func NewGroupProjectAccessStore(db *sql.DB) domain.GroupProjectAccessStore {
	return &GroupProjectAccessStore{db: db}
}

func (store *GroupProjectAccessStore) SetAccess(ctx context.Context, access *domain.GroupProjectAccess) error {
	_, err := store.db.ExecContext(ctx,
		`INSERT INTO group_project_access (group_identifier, project_identifier, access_level)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (group_identifier, project_identifier) DO UPDATE SET access_level = EXCLUDED.access_level`,
		access.GroupIdentifier, access.ProjectIdentifier, access.AccessLevel,
	)
	if err != nil {
		return fmt.Errorf("failed to set access: %w", err)
	}
	return nil
}

func (store *GroupProjectAccessStore) RemoveAccess(ctx context.Context, groupIdentifier string, projectIdentifier string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM group_project_access WHERE group_identifier = $1 AND project_identifier = $2`,
		groupIdentifier, projectIdentifier,
	)
	if err != nil {
		return fmt.Errorf("failed to remove access: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("access not found")
	}
	return nil
}

func (store *GroupProjectAccessStore) ListAccessByProject(ctx context.Context, projectIdentifier string) ([]domain.GroupProjectAccess, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT group_identifier, project_identifier, access_level
		 FROM group_project_access WHERE project_identifier = $1 ORDER BY group_identifier`, projectIdentifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query access by project: %w", err)
	}
	defer rows.Close()

	var accessList []domain.GroupProjectAccess
	for rows.Next() {
		var access domain.GroupProjectAccess
		if err := rows.Scan(&access.GroupIdentifier, &access.ProjectIdentifier, &access.AccessLevel); err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}
		accessList = append(accessList, access)
	}
	return accessList, rows.Err()
}

func (store *GroupProjectAccessStore) ListAccessByGroup(ctx context.Context, groupIdentifier string) ([]domain.GroupProjectAccess, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT group_identifier, project_identifier, access_level
		 FROM group_project_access WHERE group_identifier = $1 ORDER BY project_identifier`, groupIdentifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query access by group: %w", err)
	}
	defer rows.Close()

	var accessList []domain.GroupProjectAccess
	for rows.Next() {
		var access domain.GroupProjectAccess
		if err := rows.Scan(&access.GroupIdentifier, &access.ProjectIdentifier, &access.AccessLevel); err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}
		accessList = append(accessList, access)
	}
	return accessList, rows.Err()
}

func (store *GroupProjectAccessStore) GetUserAccessLevel(ctx context.Context, userIdentifier string, projectIdentifier string) (*domain.AccessLevel, error) {
	var maxLevel int
	err := store.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(CASE
			WHEN gpa.access_level = 'admin' THEN 3
			WHEN gpa.access_level = 'write' THEN 2
			WHEN gpa.access_level = 'read'  THEN 1
			ELSE 0
		 END), 0)
		 FROM group_memberships gm
		 INNER JOIN group_project_access gpa ON gm.group_identifier = gpa.group_identifier
		 WHERE gm.user_identifier = $1 AND gpa.project_identifier = $2`,
		userIdentifier, projectIdentifier,
	).Scan(&maxLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to query user access level: %w", err)
	}

	if maxLevel == 0 {
		return nil, nil
	}

	var accessLevel domain.AccessLevel
	switch maxLevel {
	case 3:
		accessLevel = domain.AccessLevelAdmin
	case 2:
		accessLevel = domain.AccessLevelWrite
	default:
		accessLevel = domain.AccessLevelRead
	}
	return &accessLevel, nil
}
