package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type GroupStore struct {
	db *sql.DB
}

func NewGroupStore(db *sql.DB) domain.GroupStore {
	return &GroupStore{
		db: db,
	}
}

func (store *GroupStore) GetGroupById(groupId domain.GroupId) (*domain.Group, error) {
	group := &domain.Group{}

	err := store.db.QueryRow(
		`SELECT id, group_name
		 FROM groups
		 WHERE id = $1`,
		groupId,
	).Scan(
		&group.GroupId,
		&group.GroupName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group not found: %w", err)
		}
		return nil, fmt.Errorf("failed to query group by id: %w", err)
	}

	return group, nil
}

func (store *GroupStore) ListGroups(limit int, offset int) ([]*domain.Group, error) {
	rows, err := store.db.Query(
		`SELECT id, group_name
		 FROM groups
		 ORDER BY id
		 LIMIT $1 OFFSET $2`,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group

	for rows.Next() {
		group := &domain.Group{}
		err := rows.Scan(
			&group.GroupId,
			&group.GroupName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group rows: %w", err)
	}

	return groups, nil
}

func (store *GroupStore) CreateGroup(group *domain.Group) (*domain.Group, error) {
	createdGroup := &domain.Group{}

	err := store.db.QueryRow(
		`INSERT INTO groups (group_name)
		 VALUES ($1)
		 RETURNING id, group_name`,
		group.GroupName,
	).Scan(
		&createdGroup.GroupId,
		&createdGroup.GroupName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return createdGroup, nil
}

func (store *GroupStore) DeleteGroup(groupId domain.GroupId) error {
	result, err := store.db.Exec(
		`DELETE FROM groups
		 WHERE id = $1`,
		groupId,
	)

	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

func (store *GroupStore) AddUserToGroup(groupId domain.GroupId, userId domain.UserId) error {
	_, err := store.db.Exec(
		`INSERT INTO group_memberships (group_id, user_id)
		 VALUES ($1, $2)
		 ON CONFLICT (group_id, user_id) DO NOTHING`,
		groupId,
		userId,
	)

	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

func (store *GroupStore) RemoveUserFromGroup(groupId domain.GroupId, userId domain.UserId) error {
	result, err := store.db.Exec(
		`DELETE FROM group_memberships
		 WHERE group_id = $1 AND user_id = $2`,
		groupId,
		userId,
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

func (store *GroupStore) ListGroupsByUser(userId domain.UserId, limit int, offset int) ([]*domain.Group, error) {
	rows, err := store.db.Query(
		`SELECT g.id, g.group_name
		 FROM groups g
		 INNER JOIN group_memberships gm ON g.id = gm.group_id
		 WHERE gm.user_id = $1
		 ORDER BY g.id
		 LIMIT $2 OFFSET $3`,
		userId,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query groups by user: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group

	for rows.Next() {
		group := &domain.Group{}
		err := rows.Scan(
			&group.GroupId,
			&group.GroupName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group rows: %w", err)
	}

	return groups, nil
}

func (store *GroupStore) ListUsersByGroup(groupId domain.GroupId, limit int, offset int) ([]*domain.User, error) {
	rows, err := store.db.Query(
		`SELECT u.id, u.email, u.display_name, u.external_identity_provider, u.external_identity_subject
		 FROM users u
		 INNER JOIN group_memberships gm ON u.id = gm.user_id
		 WHERE gm.group_id = $1
		 ORDER BY u.id
		 LIMIT $2 OFFSET $3`,
		groupId,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query users by group: %w", err)
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

type GroupProjectAccessStore struct {
	db *sql.DB
}

func NewGroupProjectAccessStore(db *sql.DB) domain.GroupProjectAccessStore {
	return &GroupProjectAccessStore{
		db: db,
	}
}

func (store *GroupProjectAccessStore) SetAccess(groupId domain.GroupId, projectId domain.ProjectId, accessLevel domain.AccessLevel) error {
	_, err := store.db.Exec(
		`INSERT INTO group_project_access (group_id, project_id, access_level)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (group_id, project_id) DO UPDATE SET access_level = $3`,
		groupId,
		projectId,
		accessLevel,
	)

	if err != nil {
		return fmt.Errorf("failed to set access: %w", err)
	}

	return nil
}

func (store *GroupProjectAccessStore) RemoveAccess(groupId domain.GroupId, projectId domain.ProjectId) error {
	result, err := store.db.Exec(
		`DELETE FROM group_project_access
		 WHERE group_id = $1 AND project_id = $2`,
		groupId,
		projectId,
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

func (store *GroupProjectAccessStore) ListAccessByProject(projectId domain.ProjectId, limit int, offset int) ([]*domain.GroupProjectAccess, error) {
	rows, err := store.db.Query(
		`SELECT group_id, project_id, access_level
		 FROM group_project_access
		 WHERE project_id = $1
		 ORDER BY group_id
		 LIMIT $2 OFFSET $3`,
		projectId,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query access by project: %w", err)
	}
	defer rows.Close()

	var accessList []*domain.GroupProjectAccess

	for rows.Next() {
		access := &domain.GroupProjectAccess{}
		err := rows.Scan(
			&access.GroupId,
			&access.ProjectId,
			&access.AccessLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}
		accessList = append(accessList, access)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating access rows: %w", err)
	}

	return accessList, nil
}

func (store *GroupProjectAccessStore) ListAccessByGroup(groupId domain.GroupId, limit int, offset int) ([]*domain.GroupProjectAccess, error) {
	rows, err := store.db.Query(
		`SELECT group_id, project_id, access_level
		 FROM group_project_access
		 WHERE group_id = $1
		 ORDER BY project_id
		 LIMIT $2 OFFSET $3`,
		groupId,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query access by group: %w", err)
	}
	defer rows.Close()

	var accessList []*domain.GroupProjectAccess

	for rows.Next() {
		access := &domain.GroupProjectAccess{}
		err := rows.Scan(
			&access.GroupId,
			&access.ProjectId,
			&access.AccessLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access: %w", err)
		}
		accessList = append(accessList, access)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating access rows: %w", err)
	}

	return accessList, nil
}

func (store *GroupProjectAccessStore) GetUserAccessLevel(userId domain.UserId, projectId domain.ProjectId) (domain.AccessLevel, error) {
	var accessLevel domain.AccessLevel

	err := store.db.QueryRow(
		`SELECT COALESCE(MAX(CASE
			WHEN gpa.access_level = 'admin' THEN 3
			WHEN gpa.access_level = 'write' THEN 2
			WHEN gpa.access_level = 'read' THEN 1
			ELSE 0
		 END), 0) as max_level
		 FROM group_memberships gm
		 INNER JOIN group_project_access gpa ON gm.group_id = gpa.group_id
		 WHERE gm.user_id = $1 AND gpa.project_id = $2`,
		userId,
		projectId,
	).Scan(&accessLevel)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no access found: %w", err)
		}
		return "", fmt.Errorf("failed to query user access level: %w", err)
	}

	return accessLevel, nil
}
