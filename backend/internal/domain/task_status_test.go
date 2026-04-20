package domain

import "testing"

func TestTaskStatus_IsValid(t *testing.T) {
	tests := []struct {
		status TaskStatus
		valid  bool
	}{
		{TaskStatusBacklog, true},
		{TaskStatusTodo, true},
		{TaskStatusInProgress, true},
		{TaskStatusInReview, true},
		{TaskStatusTesting, true},
		{TaskStatusBlocked, true},
		{TaskStatusDone, true},
		{TaskStatusRejected, true},
		{TaskStatus(""), false},
		{TaskStatus("unknown"), false},
		{TaskStatus("DONE"), false},
	}
	for _, testCase := range tests {
		if got := testCase.status.IsValid(); got != testCase.valid {
			t.Errorf("IsValid(%q) = %v, want %v", testCase.status, got, testCase.valid)
		}
	}
}

func TestTaskStatus_IsOpenAndClosed(t *testing.T) {
	openStatuses := []TaskStatus{
		TaskStatusBacklog,
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusInReview,
		TaskStatusTesting,
		TaskStatusBlocked,
	}
	for _, status := range openStatuses {
		if !status.IsOpenStatus() {
			t.Errorf("expected %q to be open", status)
		}
		if status.IsClosedStatus() {
			t.Errorf("expected %q not to be closed", status)
		}
	}

	closedStatuses := []TaskStatus{TaskStatusDone, TaskStatusRejected}
	for _, status := range closedStatuses {
		if status.IsOpenStatus() {
			t.Errorf("expected %q not to be open", status)
		}
		if !status.IsClosedStatus() {
			t.Errorf("expected %q to be closed", status)
		}
	}
}

func TestAllTaskStatuses_ContainsAllEightValues(t *testing.T) {
	all := AllTaskStatuses()
	if len(all) != 8 {
		t.Fatalf("expected 8 task statuses, got %d", len(all))
	}

	seen := make(map[TaskStatus]bool, len(all))
	for _, status := range all {
		if seen[status] {
			t.Errorf("duplicate task status %q", status)
		}
		seen[status] = true
	}
}
