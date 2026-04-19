package domain

import "time"

// User represents an authenticated user in the system.
type User struct {
	UserIdentifier           string
	Email                    string
	DisplayName              string
	ExternalIdentityProvider string // e.g. "office365", "google"
	ExternalIdentitySubject  string // subject claim from the identity provider
}

// Group represents a collection of users with shared project access.
type Group struct {
	GroupIdentifier string
	GroupName       string
}

// GroupMembership links a user to a group.
type GroupMembership struct {
	GroupIdentifier string
	UserIdentifier  string
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
	GroupIdentifier   string
	ProjectIdentifier string
	AccessLevel       AccessLevel
}

// APIKey represents a hashed API key for CLI/automation access.
type APIKey struct {
	APIKeyIdentifier string
	UserIdentifier   string
	HashedKey        string
	Label            string
	CreatedTime      time.Time
	LastUsedTime     *time.Time
}
