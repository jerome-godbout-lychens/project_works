package domain

import "context"

// UserStore handles user persistence.
type UserStore interface {
	GetUserById(context context.Context, userId string) (*User, error)
	GetUserByExternalIdentity(context context.Context, provider string, subject string) (*User, error)
	CreateUser(context context.Context, user *User) error
	UpdateUser(context context.Context, user *User) error
	ListUsers(context context.Context, limit int, offset int) ([]User, error)
}

// GroupStore handles group and membership persistence.
type GroupStore interface {
	GetGroupById(context context.Context, groupId string) (*Group, error)
	ListGroups(context context.Context) ([]Group, error)
	CreateGroup(context context.Context, group *Group) error
	DeleteGroup(context context.Context, groupId string) error
	AddUserToGroup(context context.Context, groupId string, userId string) error
	RemoveUserFromGroup(context context.Context, groupId string, userId string) error
	ListGroupsByUser(context context.Context, userId string) ([]Group, error)
	ListUsersByGroup(context context.Context, groupId string) ([]User, error)
}

// GroupProjectAccessStore handles group-level project permissions.
type GroupProjectAccessStore interface {
	SetAccess(context context.Context, access *GroupProjectAccess) error
	RemoveAccess(context context.Context, groupId string, projectId string) error
	ListAccessByProject(context context.Context, projectId string) ([]GroupProjectAccess, error)
	ListAccessByGroup(context context.Context, groupId string) ([]GroupProjectAccess, error)
	GetUserAccessLevel(context context.Context, userId string, projectId string) (*AccessLevel, error)
}

// APIKeyStore handles API key persistence.
type APIKeyStore interface {
	CreateAPIKey(context context.Context, apiKey *APIKey) error
	GetAPIKeyByHash(context context.Context, hashedKey string) (*APIKey, error)
	ListAPIKeysByUser(context context.Context, userId string) ([]APIKey, error)
	DeleteAPIKey(context context.Context, apiKeyId string) error
	UpdateLastUsedTime(context context.Context, apiKeyId string) error
}
