package domain

import "time"

// Project represents a trackable project, organized in a folder hierarchy.
type Project struct {
	ProjectIdentifier  string
	ProjectName        string
	ProjectDescription string
	FolderPath         string // ltree path, e.g. "engineering.firmware.sensors"
	CreationTime       time.Time
	ModificationTime   time.Time
}
