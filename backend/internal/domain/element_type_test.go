package domain

import "testing"

func TestElementType_IsValid(t *testing.T) {
	tests := []struct {
		elementType ElementType
		valid       bool
	}{
		{ElementTypeRequirement, true},
		{ElementTypeFeature, true},
		{ElementTypeTask, true},
		{ElementTypeBug, true},
		{ElementTypeEvaluation, true},
		{ElementTypeRisk, true},
		{ElementType(""), false},
		{ElementType("Task"), false},
		{ElementType("unknown"), false},
	}
	for _, testCase := range tests {
		if got := testCase.elementType.IsValid(); got != testCase.valid {
			t.Errorf("IsValid(%q) = %v, want %v", testCase.elementType, got, testCase.valid)
		}
	}
}

func TestAllElementTypes_ContainsAllSixValuesUnique(t *testing.T) {
	all := AllElementTypes()
	if len(all) != 6 {
		t.Fatalf("expected 6 element types, got %d", len(all))
	}
	seen := make(map[ElementType]bool, len(all))
	for _, elementType := range all {
		if seen[elementType] {
			t.Errorf("duplicate element type %q", elementType)
		}
		seen[elementType] = true
	}
}
