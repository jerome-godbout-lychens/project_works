package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// VersionableSnapshot is the canonical serializable form of an element's full state.
// It includes core fields, custom fields, supervisors, clients, links, and attachment metadata.
// This struct is serialized to deterministic JSON for diffing and SHA computation.
type VersionableSnapshot struct {
	ElementType   domain.ElementType   `json:"element_type"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	InterestLevel *int                 `json:"interest_level,omitempty"`
	AssigneeIdentifier    *string              `json:"assignee_identifier,omitempty"`
	TaskStatus    *domain.TaskStatus   `json:"task_status,omitempty"`
	TaskProgress  *int                 `json:"task_progress,omitempty"`
	ParentFeatureIdentifier  *string           `json:"parent_feature_identifier,omitempty"`
	StartPhaseIdentifier     *string           `json:"start_phase_identifier,omitempty"`
	DeliveryPhaseIdentifier  *string           `json:"delivery_phase_identifier,omitempty"`
	CustomFieldValues []SnapshotCustomField `json:"custom_field_values"`
	Supervisors       []string              `json:"supervisors"`
	Clients           []SnapshotContact     `json:"clients"`
	Links             []SnapshotLink        `json:"links"`
	Attachments       []SnapshotAttachment  `json:"attachments"`
}

type SnapshotCustomField struct {
	FieldDefinitionIdentifier string      `json:"field_definition_identifier"`
	FieldName         string      `json:"field_name"`
	FieldValue        interface{} `json:"field_value"`
}

type SnapshotContact struct {
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	ContactNotes string `json:"contact_notes"`
}

type SnapshotLink struct {
	SourceElementIdentifier      string          `json:"source_element_identifier"`
	DestinationElementIdentifier string          `json:"destination_element_identifier"`
	LinkType             domain.LinkType `json:"link_type"`
}

type SnapshotAttachment struct {
	AttachmentIdentifier string `json:"attachment_identifier"`
	FileName     string `json:"file_name"`
	ContentType  string `json:"content_type"`
	FileSizeBytes int64 `json:"file_size_bytes"`
}

// BuildSnapshot assembles a VersionableSnapshot from the element and its related data.
func BuildSnapshot(
	element *domain.Element,
	links []domain.ElementLink,
	attachments []domain.Attachment,
) *VersionableSnapshot {
	snapshot := &VersionableSnapshot{
		ElementType:     element.ElementType,
		Title:           element.Title,
		Description:     element.Description,
		InterestLevel:   element.InterestLevel,
		AssigneeIdentifier:      element.AssigneeIdentifier,
		TaskStatus:      element.TaskStatus,
		TaskProgress:    element.TaskProgress,
		ParentFeatureIdentifier: element.ParentFeatureIdentifier,
		StartPhaseIdentifier:    element.StartPhaseIdentifier,
		DeliveryPhaseIdentifier: element.DeliveryPhaseIdentifier,
	}

	// Custom fields — sorted by definition ID for deterministic output
	snapshot.CustomFieldValues = make([]SnapshotCustomField, len(element.CustomFieldValues))
	for i, customField := range element.CustomFieldValues {
		snapshot.CustomFieldValues[i] = SnapshotCustomField{
			FieldDefinitionIdentifier: customField.FieldDefinitionIdentifier,
			FieldName:         customField.FieldName,
			FieldValue:        customField.FieldValue,
		}
	}
	sort.Slice(snapshot.CustomFieldValues, func(i, j int) bool {
		return snapshot.CustomFieldValues[i].FieldDefinitionIdentifier < snapshot.CustomFieldValues[j].FieldDefinitionIdentifier
	})

	// Supervisors — sorted for determinism
	snapshot.Supervisors = make([]string, len(element.Supervisors))
	copy(snapshot.Supervisors, element.Supervisors)
	sort.Strings(snapshot.Supervisors)

	// Clients — sorted by name
	snapshot.Clients = make([]SnapshotContact, len(element.Clients))
	for i, client := range element.Clients {
		snapshot.Clients[i] = SnapshotContact{
			ContactName:  client.ContactName,
			ContactEmail: client.ContactEmail,
			ContactNotes: client.ContactNotes,
		}
	}
	sort.Slice(snapshot.Clients, func(i, j int) bool {
		return snapshot.Clients[i].ContactName < snapshot.Clients[j].ContactName
	})

	// Links — sorted by destination + type
	snapshot.Links = make([]SnapshotLink, len(links))
	for i, link := range links {
		snapshot.Links[i] = SnapshotLink{
			SourceElementIdentifier:      link.SourceElementIdentifier,
			DestinationElementIdentifier: link.DestinationElementIdentifier,
			LinkType:             link.LinkType,
		}
	}
	sort.Slice(snapshot.Links, func(i, j int) bool {
		if snapshot.Links[i].DestinationElementIdentifier != snapshot.Links[j].DestinationElementIdentifier {
			return snapshot.Links[i].DestinationElementIdentifier < snapshot.Links[j].DestinationElementIdentifier
		}
		return string(snapshot.Links[i].LinkType) < string(snapshot.Links[j].LinkType)
	})

	// Attachments — sorted by ID
	snapshot.Attachments = make([]SnapshotAttachment, len(attachments))
	for i, attachment := range attachments {
		snapshot.Attachments[i] = SnapshotAttachment{
			AttachmentIdentifier:  attachment.AttachmentIdentifier,
			FileName:      attachment.FileName,
			ContentType:   attachment.ContentType,
			FileSizeBytes: attachment.FileSizeBytes,
		}
	}
	sort.Slice(snapshot.Attachments, func(i, j int) bool {
		return snapshot.Attachments[i].AttachmentIdentifier < snapshot.Attachments[j].AttachmentIdentifier
	})

	return snapshot
}

// SerializeSnapshot converts a snapshot to deterministic JSON bytes.
func SerializeSnapshot(snapshot *VersionableSnapshot) ([]byte, error) {
	return json.Marshal(snapshot)
}

// DeserializeSnapshot parses JSON bytes back into a VersionableSnapshot.
func DeserializeSnapshot(data []byte) (*VersionableSnapshot, error) {
	var snapshot VersionableSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// ComputeContentSha returns the SHA-256 hex digest of the serialized snapshot.
func ComputeContentSha(snapshotJson []byte) string {
	hash := sha256.Sum256(snapshotJson)
	return fmt.Sprintf("%x", hash)
}
