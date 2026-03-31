package domain

// TaskStatus represents the lifecycle state of a task element.
type TaskStatus string

const (
	TaskStatusBacklog    TaskStatus = "backlog"
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusInReview   TaskStatus = "in_review"
	TaskStatusTesting    TaskStatus = "testing"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusRejected   TaskStatus = "rejected"
)

// AllTaskStatuses returns every valid task status.
func AllTaskStatuses() []TaskStatus {
	return []TaskStatus{
		TaskStatusBacklog,
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusInReview,
		TaskStatusTesting,
		TaskStatusBlocked,
		TaskStatusDone,
		TaskStatusRejected,
	}
}

// IsValid checks whether the task status is one of the known values.
func (taskStatus TaskStatus) IsValid() bool {
	for _, validStatus := range AllTaskStatuses() {
		if taskStatus == validStatus {
			return true
		}
	}
	return false
}

// IsOpenStatus returns true if the task status represents an open (non-closed) state.
func (taskStatus TaskStatus) IsOpenStatus() bool {
	return taskStatus != TaskStatusDone && taskStatus != TaskStatusRejected
}

// IsClosedStatus returns true if the task status represents a closed state.
func (taskStatus TaskStatus) IsClosedStatus() bool {
	return !taskStatus.IsOpenStatus()
}
