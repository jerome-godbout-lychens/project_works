package domain

import "time"

// Element is the core domain entity representing any trackable item in a project.
// Type-specific fields use pointers and are non-nil only when relevant to the ElementType.
type Element struct {
	ElementIdentifier string
	ProjectIdentifier string
	ElementType       ElementType
	Title             string
	Description       string
	ContentSha        string
	CreationTime      time.Time
	ModificationTime  time.Time

	// Requirement-specific fields.
	InterestLevel *int // 1–10 scale, nil for non-requirements.

	// Task-specific fields.
	AssigneeIdentifier      *string
	TaskStatus              *TaskStatus
	TaskProgress            *int // 0–100 percentage, nil for non-tasks.
	CloseTime               *time.Time
	ParentFeatureIdentifier *string
	StartPhaseIdentifier    *string
	DeliveryPhaseIdentifier *string

	// Populated via separate queries (not stored inline in the elements table).
	CustomFieldValues []CustomFieldValue
	Supervisors       []string // user identifiers
	Clients           []ContactInfo
}

// ContactInfo represents a client contact associated with a requirement.
type ContactInfo struct {
	ContactName  string
	ContactEmail string
	ContactNotes string
}

// ComputeFeatureProgress calculates a feature's progress as the average of its
// child task progress values. This is a derived value — never stored in the database.
// Returns 0 if childTasks is empty.
func ComputeFeatureProgress(childTasks []Element) int {
	if len(childTasks) == 0 {
		return 0
	}
	totalProgress := 0
	countWithProgress := 0
	for _, task := range childTasks {
		if task.TaskProgress != nil {
			totalProgress += *task.TaskProgress
			countWithProgress++
		}
	}
	if countWithProgress == 0 {
		return 0
	}
	return totalProgress / countWithProgress
}
