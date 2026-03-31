package domain

import "time"

// User represents an authenticated user in the system.
type User struct {
	UserId           string
	Email                    string
	DisplayName              string
	ExternalIdentityProvider string // e.g. "office365", "google"
	ExternalIdentitySubject  string // subject claim from the identity provider
}

// Group represents a collection of users with shared project access.
type Group struct {
	GroupId string
	GroupName       string
}

// GroupMembership links a user to a group.
type GroupMembership struct {
	GroupId string
	UserId  string
}

// AccessLevel defines the permission a group has on a project.
type AccessLevel string

const (
	AccessLevelRead  AccessLevel = "read"
	AccessLevelWrite AccessLevel = "write"
	AccessLevelAdmin AccessLevel = "admin"
)

// GroupProjectAccess links a group to a project with a specific access level.
type GroupProjectAccess struct {
	GroupId   string
	ProjectId string
	AccessLevel       AccessLevel
}

// APIKey represents a hashed API key for CLI/automation access.
type APIKey struct {
	APIKeyId string
	UserId   string
	HashedKey        string
	Label            string
	CreatedTime      time.Time
	LastUsedTime     *time.Time
}
