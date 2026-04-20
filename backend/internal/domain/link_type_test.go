package domain

import "testing"

func TestLinkType_IsValid(t *testing.T) {
	tests := []struct {
		linkType LinkType
		valid    bool
	}{
		{LinkTypeRelated, true},
		{LinkTypeChild, true},
		{LinkTypeImplement, true},
		{LinkType(""), false},
		{LinkType("parent"), false},
		{LinkType("Related"), false},
	}
	for _, testCase := range tests {
		if got := testCase.linkType.IsValid(); got != testCase.valid {
			t.Errorf("IsValid(%q) = %v, want %v", testCase.linkType, got, testCase.valid)
		}
	}
}

func TestAllLinkTypes_ContainsAllThreeValuesUnique(t *testing.T) {
	all := AllLinkTypes()
	if len(all) != 3 {
		t.Fatalf("expected 3 link types, got %d", len(all))
	}
	seen := make(map[LinkType]bool, len(all))
	for _, linkType := range all {
		if seen[linkType] {
			t.Errorf("duplicate link type %q", linkType)
		}
		seen[linkType] = true
	}
}
