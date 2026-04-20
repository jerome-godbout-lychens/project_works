package domain

import "testing"

func intPointer(value int) *int {
	return &value
}

func TestComputeFeatureProgress_EmptyReturnsZero(t *testing.T) {
	if got := ComputeFeatureProgress(nil); got != 0 {
		t.Errorf("expected 0 for nil slice, got %d", got)
	}
	if got := ComputeFeatureProgress([]Element{}); got != 0 {
		t.Errorf("expected 0 for empty slice, got %d", got)
	}
}

func TestComputeFeatureProgress_AveragesTaskProgress(t *testing.T) {
	tasks := []Element{
		{TaskProgress: intPointer(20)},
		{TaskProgress: intPointer(40)},
		{TaskProgress: intPointer(60)},
	}
	if got := ComputeFeatureProgress(tasks); got != 40 {
		t.Errorf("expected average 40, got %d", got)
	}
}

func TestComputeFeatureProgress_IgnoresTasksWithNilProgress(t *testing.T) {
	tasks := []Element{
		{TaskProgress: intPointer(100)},
		{TaskProgress: nil},
		{TaskProgress: intPointer(50)},
	}
	if got := ComputeFeatureProgress(tasks); got != 75 {
		t.Errorf("expected 75 (avg of 100 and 50), got %d", got)
	}
}

func TestComputeFeatureProgress_AllNilReturnsZero(t *testing.T) {
	tasks := []Element{
		{TaskProgress: nil},
		{TaskProgress: nil},
	}
	if got := ComputeFeatureProgress(tasks); got != 0 {
		t.Errorf("expected 0 when all tasks have nil progress, got %d", got)
	}
}

func TestComputeFeatureProgress_IntegerDivisionTruncates(t *testing.T) {
	tasks := []Element{
		{TaskProgress: intPointer(50)},
		{TaskProgress: intPointer(51)},
	}
	if got := ComputeFeatureProgress(tasks); got != 50 {
		t.Errorf("expected 50 from integer division of 101/2, got %d", got)
	}
}
