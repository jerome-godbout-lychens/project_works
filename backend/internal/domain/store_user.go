package domain

import "context"

// UserStore handles user persistence.
type UserStore interface {
	GetUserByIdentifier(context context.Context, userIdentifier string) (*User, error)
	GetUserByExternalIdentity(context context.Context, provider string, subject string) (*User, error)
	CreateUser(context context.Context, user *User) error
	UpdateUser(context context.Context, user *User) error
	ListUsers(context context.Context, limit int, offset int) ([]User, error)
}

// GroupStore handles group and membership persistence.
type GroupStore interface {
	GetGroupByIdentifier(context context.Context, groupIdentifier string) (*Group, error)
	ListGroups(context context.Context) ([]Group, error)
	CreateGroup(context context.Context, group *Group) error
	DeleteGroup(context context.Context, groupIdentifier string) error
	AddUserToGroup(context context.Context, groupIdentifier string, userIdentifier string) error
	RemoveUserFromGroup(context context.Context, groupIdentifier string, userIdentifier string) error
	ListGroupsByUser(context context.Context, userIdentifier string) ([]Group, error)
	ListUsersByGroup(context context.Context, groupIdentifier string) ([]User, error)
}

// GroupProjectAccessStore handles group-level project permissions.
type GroupProjectAccessStore interface {
	SetAccess(context context.Context, access *GroupProjectAccess) error
	RemoveAccess(context context.Context, groupIdentifier string, projectIdentifier string) error
	ListAccessByProject(context context.Context, projectIdentifier string) ([]GroupProjectAccess, error)
	ListAccessByGroup(context context.Context, groupIdentifier string) ([]GroupProjectAccess, error)
	GetUserAccessLevel(context context.Context, userIdentifier string, projectIdentifier string) (*AccessLevel, error)
}

// APIKeyStore handles API key persistence.
type APIKeyStore interface {
	CreateAPIKey(context context.Context, apiKey *APIKey) error
	GetAPIKeyByHash(context context.Context, hashedKey string) (*APIKey, error)
	ListAPIKeysByUser(context context.Context, userIdentifier string) ([]APIKey, error)
	DeleteAPIKey(context context.Context, apiKeyIdentifier string) error
	UpdateLastUsedTime(context context.Context, apiKeyIdentifier string) error
}
