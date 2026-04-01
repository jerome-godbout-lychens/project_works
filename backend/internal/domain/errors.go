package domain

import "errors"

// Sentinel errors returned by store implementations.
var (
	ErrNotFound        = errors.New("not found")
	ErrElementNotFound = errors.New("element not found")
	ErrProjectNotFound = errors.New("project not found")
	ErrLinkNotFound    = errors.New("link not found")
	ErrPhaseNotFound   = errors.New("phase not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrGroupNotFound   = errors.New("group not found")
	ErrAPIKeyNotFound  = errors.New("api key not found")
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrDuplicateLink   = errors.New("duplicate link")
	ErrAccessDenied    = errors.New("access denied")
)
