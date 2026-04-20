package service

import (
	"testing"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

func stringPointer(value string) *string {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func sampleTaskElement() *domain.Element {
	status := domain.TaskStatusInProgress
	return &domain.Element{
		ElementIdentifier:       "element-1",
		ProjectIdentifier:       "project-1",
		ElementType:             domain.ElementTypeTask,
		Title:                   "Write tests",
		Description:             "Cover the snapshot builder",
		AssigneeIdentifier:      stringPointer("user-alpha"),
		TaskStatus:              &status,
		TaskProgress:            intPointer(42),
		ParentFeatureIdentifier: stringPointer("feature-1"),
		StartPhaseIdentifier:    stringPointer("phase-1"),
		DeliveryPhaseIdentifier: stringPointer("phase-2"),
		Supervisors:             []string{"zeta", "alpha", "mike"},
		Clients: []domain.ContactInfo{
			{ContactName: "Zelda", ContactEmail: "z@x.com"},
			{ContactName: "Alice", ContactEmail: "a@x.com"},
		},
		CustomFieldValues: []domain.CustomFieldValue{
			{FieldDefinitionIdentifier: "field-b", FieldName: "B", FieldValue: 2},
			{FieldDefinitionIdentifier: "field-a", FieldName: "A", FieldValue: 1},
		},
	}
}

func TestBuildSnapshot_CopiesCoreFields(t *testing.T) {
	element := sampleTaskElement()

	snapshot := BuildSnapshot(element, nil, nil)

	if snapshot.ElementType != domain.ElementTypeTask {
		t.Errorf("ElementType mismatch: got %q", snapshot.ElementType)
	}
	if snapshot.Title != "Write tests" {
		t.Errorf("Title mismatch: got %q", snapshot.Title)
	}
	if snapshot.TaskStatus == nil || *snapshot.TaskStatus != domain.TaskStatusInProgress {
		t.Errorf("TaskStatus mismatch: got %v", snapshot.TaskStatus)
	}
	if snapshot.TaskProgress == nil || *snapshot.TaskProgress != 42 {
		t.Errorf("TaskProgress mismatch: got %v", snapshot.TaskProgress)
	}
	if snapshot.AssigneeIdentifier == nil || *snapshot.AssigneeIdentifier != "user-alpha" {
		t.Errorf("AssigneeIdentifier mismatch: got %v", snapshot.AssigneeIdentifier)
	}
}

func TestBuildSnapshot_SortsSupervisorsClientsAndCustomFields(t *testing.T) {
	element := sampleTaskElement()

	snapshot := BuildSnapshot(element, nil, nil)

	expectedSupervisors := []string{"alpha", "mike", "zeta"}
	for i, value := range expectedSupervisors {
		if snapshot.Supervisors[i] != value {
			t.Errorf("Supervisors not sorted: %v", snapshot.Supervisors)
			break
		}
	}

	if snapshot.Clients[0].ContactName != "Alice" || snapshot.Clients[1].ContactName != "Zelda" {
		t.Errorf("Clients not sorted by name: %+v", snapshot.Clients)
	}

	if snapshot.CustomFieldValues[0].FieldDefinitionIdentifier != "field-a" ||
		snapshot.CustomFieldValues[1].FieldDefinitionIdentifier != "field-b" {
		t.Errorf("CustomFieldValues not sorted by definition id: %+v", snapshot.CustomFieldValues)
	}
}

func TestBuildSnapshot_SortsLinksAndAttachments(t *testing.T) {
	element := sampleTaskElement()
	links := []domain.ElementLink{
		{SourceElementIdentifier: "s", DestinationElementIdentifier: "z", LinkType: domain.LinkTypeRelated},
		{SourceElementIdentifier: "s", DestinationElementIdentifier: "a", LinkType: domain.LinkTypeChild},
		{SourceElementIdentifier: "s", DestinationElementIdentifier: "a", LinkType: domain.LinkTypeImplement},
	}
	attachments := []domain.Attachment{
		{AttachmentIdentifier: "z-id", FileName: "z.txt"},
		{AttachmentIdentifier: "a-id", FileName: "a.txt"},
	}

	snapshot := BuildSnapshot(element, links, attachments)

	if snapshot.Links[0].DestinationElementIdentifier != "a" ||
		snapshot.Links[1].DestinationElementIdentifier != "a" ||
		snapshot.Links[2].DestinationElementIdentifier != "z" {
		t.Errorf("Links not sorted by destination: %+v", snapshot.Links)
	}
	// For destination "a", sort ties by link type name: "child" < "implement".
	if snapshot.Links[0].LinkType != domain.LinkTypeChild ||
		snapshot.Links[1].LinkType != domain.LinkTypeImplement {
		t.Errorf("Links with same destination not sorted by link type: %+v", snapshot.Links)
	}

	if snapshot.Attachments[0].AttachmentIdentifier != "a-id" ||
		snapshot.Attachments[1].AttachmentIdentifier != "z-id" {
		t.Errorf("Attachments not sorted by identifier: %+v", snapshot.Attachments)
	}
}

func TestBuildSnapshot_DeterministicJSONForUnorderedInputs(t *testing.T) {
	firstElement := sampleTaskElement()
	// Same data but fields given in different order.
	secondElement := sampleTaskElement()
	secondElement.Supervisors = []string{"mike", "alpha", "zeta"}
	secondElement.Clients = []domain.ContactInfo{
		{ContactName: "Alice", ContactEmail: "a@x.com"},
		{ContactName: "Zelda", ContactEmail: "z@x.com"},
	}
	secondElement.CustomFieldValues = []domain.CustomFieldValue{
		{FieldDefinitionIdentifier: "field-a", FieldName: "A", FieldValue: 1},
		{FieldDefinitionIdentifier: "field-b", FieldName: "B", FieldValue: 2},
	}

	firstJson, err := SerializeSnapshot(BuildSnapshot(firstElement, nil, nil))
	if err != nil {
		t.Fatalf("first serialize error: %v", err)
	}
	secondJson, err := SerializeSnapshot(BuildSnapshot(secondElement, nil, nil))
	if err != nil {
		t.Fatalf("second serialize error: %v", err)
	}

	if string(firstJson) != string(secondJson) {
		t.Errorf("expected deterministic serialization, got:\n%s\nvs\n%s", firstJson, secondJson)
	}
}

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	element := sampleTaskElement()
	links := []domain.ElementLink{
		{SourceElementIdentifier: "s", DestinationElementIdentifier: "d", LinkType: domain.LinkTypeRelated},
	}
	attachments := []domain.Attachment{
		{AttachmentIdentifier: "att-1", FileName: "doc.pdf", ContentType: "application/pdf", FileSizeBytes: 1024},
	}

	snapshot := BuildSnapshot(element, links, attachments)
	snapshotJson, err := SerializeSnapshot(snapshot)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	reconstructed, err := DeserializeSnapshot(snapshotJson)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}

	if reconstructed.Title != snapshot.Title {
		t.Errorf("Title mismatch after round-trip")
	}
	if len(reconstructed.Links) != len(snapshot.Links) {
		t.Errorf("Link count mismatch: got %d, want %d", len(reconstructed.Links), len(snapshot.Links))
	}
	if len(reconstructed.Attachments) != 1 || reconstructed.Attachments[0].FileName != "doc.pdf" {
		t.Errorf("Attachment round-trip failed: %+v", reconstructed.Attachments)
	}
}

func TestComputeContentSha_DeterministicAndDifferentiatesInputs(t *testing.T) {
	firstHash := ComputeContentSha([]byte(`{"a":1}`))
	sameHash := ComputeContentSha([]byte(`{"a":1}`))
	differentHash := ComputeContentSha([]byte(`{"a":2}`))

	if firstHash != sameHash {
		t.Errorf("expected deterministic SHA, got %q and %q", firstHash, sameHash)
	}
	if firstHash == differentHash {
		t.Error("expected different inputs to produce different SHAs")
	}
	if len(firstHash) != 64 {
		t.Errorf("expected SHA-256 hex length 64, got %d", len(firstHash))
	}
}
