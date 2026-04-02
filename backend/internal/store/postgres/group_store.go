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

// NewGroupStore creates a new GroupStore instance.
func NewGroupStore(db *sql.DB) domain.GroupStore {
	return &GroupStore{db: db}
}

func (store *GroupStore) GetGroupById(ctx context.Context, groupId string) (*domain.Group, error) {
	group := &domain.Group{}
	err := store.db.QueryRowContext(ctx,
		`SELECT id, group_name FROM groups WHERE id = $1`, groupId,
	).Scan(&group.GroupId, &group.GroupName)
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
		`SELECT id, group_name FROM groups ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(&group.GroupId, &group.GroupName); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (store *GroupStore) CreateGroup(ctx context.Context, group *domain.Group) error {
	err := store.db.QueryRowContext(ctx,
		`INSERT INTO groups (group_name) VALUES ($1) RETURNING id`,
		group.GroupName,
	).Scan(&group.GroupId)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (store *GroupStore) DeleteGroup(ctx context.Context, groupId string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM groups WHERE id = $1`, groupId)
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

func (store *GroupStore) AddUserToGroup(ctx context.Context, groupId string, userId string) error {
	_, err := store.db.ExecContext(ctx,
		`INSERT INTO group_memberships (group_id, user_id)
		 VALUES ($1, $2) ON CONFLICT (group_id, user_id) DO NOTHING`,
		groupId, userId,
	)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	return nil
}

func (store *GroupStore) RemoveUserFromGroup(ctx context.Context, groupId string, userId string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM group_memberships WHERE group_id = $1 AND user_id = $2`,
		groupId, userId,
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

func (store *GroupStore) ListGroupsByUser(ctx context.Context, userId string) ([]domain.Group, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT g.id, g.group_name
		 FROM groups g
		 INNER JOIN group_memberships gm ON g.id = gm.group_id
		 WHERE gm.user_id = $1
		 ORDER BY g.id`, userId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups by user: %w", err)
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(&group.GroupId, &group.GroupName); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (store *GroupStore) ListUsersByGroup(ctx context.Context, groupId string) ([]domain.User, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.display_name, u.external_identity_provider, u.external_identity_subject
		 FROM users u
		 INNER JOIN group_memberships gm ON u.id = gm.user_id
		 WHERE gm.group_id = $1
		 ORDER BY u.id`, groupId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users by group: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserId, &user.Email, &user.DisplayName,
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

// NewGroupProjectAccessStore creates a new GroupProjectAccessStore instance.
func NewGroupProjectAccessStore(db *sql.DB) domain.GroupProjectAccessStore {
	return &GroupProjectAccessStore{db: db}
}

func (store *GroupProjectAccessStore) SetAccess(ctx context.Context, access *domain.GroupProjectAccess) error {
	_, err := store.db.ExecContext(ctx,
		`INSERT INTO group_project_access (group_id, project_id, access_level)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (group_id, project_id) DO UPDATE SET access_level = EXCLUDED.access_level`,
		access.GroupId, access.ProjectId, access.AccessLevel,
	)
	if err != nil {
		return fmt.Errorf("failed to set access: %w", err)
	}
	return nil
}

func (store *GroupProjectAccessStore) RemoveAccess(ctx context.Context, groupId string, projectId string) error {
	result, err := store.db.ExecContext(ctx,
		`DELETE FROM group_project_access WHERE group_id = $1 AND project_id = $2`,
		groupId, projectId,
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

func (store *GroupProjectAccessStore) ListAccessByProject(ctx context.Context, projectId string) ([]domain.GroupProjectAccess, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT group_id, project_id, access_level
		 FROM group_project_access WHERE project_id = $1 ORDER BY group_id`, projectId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query access by project: %w", err)
	}
	defer rows.Close()

	var accessList []domain.GroupProjectAccess
	for rows.Next() {
		var access domain.GroupProjectAccess
		if err := rows.Scan(&access.GroupId, &access.ProjectId, &access.AccessLevel); err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}
		accessList = append(accessList, access)
	}
	return accessList, rows.Err()
}

func (store *GroupProjectAccessStore) ListAccessByGroup(ctx context.Context, groupId string) ([]domain.GroupProjectAccess, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT group_id, project_id, access_level
		 FROM group_project_access WHERE group_id = $1 ORDER BY project_id`, groupId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query access by group: %w", err)
	}
	defer rows.Close()

	var accessList []domain.GroupProjectAccess
	for rows.Next() {
		var access domain.GroupProjectAccess
		if err := rows.Scan(&access.GroupId, &access.ProjectId, &access.AccessLevel); err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}
		accessList = append(accessList, access)
	}
	return accessList, rows.Err()
}

func (store *GroupProjectAccessStore) GetUserAccessLevel(ctx context.Context, userId string, projectId string) (*domain.AccessLevel, error) {
	// Returns the highest access level the user has via any of their groups.
	// Uses CASE ordering: admin=3, write=2, read=1 to find the maximum.
	var maxLevel int
	err := store.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(CASE
			WHEN gpa.access_level = 'admin' THEN 3
			WHEN gpa.access_level = 'write' THEN 2
			WHEN gpa.access_level = 'read'  THEN 1
			ELSE 0
		 END), 0)
		 FROM group_memberships gm
		 INNER JOIN group_project_access gpa ON gm.group_id = gpa.group_id
		 WHERE gm.user_id = $1 AND gpa.project_id = $2`,
		userId, projectId,
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
