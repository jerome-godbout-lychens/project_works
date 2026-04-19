package domain

import "time"

// Phase represents a project phase used for scheduling tasks.
type Phase struct {
	PhaseIdentifier   string
	ProjectIdentifier string
	PhaseName         string
	PhaseOrder        int
	PlannedStartDate  *time.Time
	PlannedEndDate    *time.Time
}
